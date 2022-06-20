// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// tfidf is an early prototype that counts term frequence in Chinese text files.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	input    = flag.String("input", "", "Location containing documents to read.")
	cnrHome    = flag.String("cnreader_home", "..", "Top level directory to search for config files.")
	corpusFN = flag.String("corpus_fn", "", "File containing list of document collections to read.")
	filter = flag.String("filter", "本作品在全世界都属于公有领域", "Regex filter pattern to use to filter out lines.")
	tfDocOut   = flag.String("tfdoc_out", "word_freq_doc.txt", "Term frequency per document output file")
	dfDocOut   = flag.String("df_out", "doc_freq.txt", "Document frequency output file")
)

var (
	charCounter = beam.NewCounter("extract", "charCounter")
)

func init() {
	beam.RegisterFunction(concatLines)
	beam.RegisterFunction(extractDocFreqFn)
	beam.RegisterFunction(extractTF)
	beam.RegisterFunction(formatTFDocEntries)
	beam.RegisterFunction(formatDFFn)
	beam.RegisterFunction(lineMap)
	beam.RegisterFunction(sumTermFreq)
	beam.RegisterType(reflect.TypeOf((*addDocEntryMetaFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*extractTermsFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*filterFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*CorpusEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*DocEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*TermFreqEntry)(nil)).Elem())
}

type CollectionEntry struct {
	CollectionFile, GlossFile string
}

// CorpusEntry contains metadata for a document that text will be read from
type CorpusEntry struct {
	RawFile string `beam:"rawFile"`
	GlossFile string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
	CorpusLen int `beam:"corpusLen"`
}

// DocEntry contains all text and metadata for the document that it was extracted from
type DocEntry struct {
	Text string `beam:"text"`
	GlossFile string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
	CorpusLen int `beam:"corpusLen"`
}

// TermFreqEntry contains term frequency and metadata for the document that it was occurred in
type TermFreqEntry struct {
	Term string `beam:"term"`
	Freq int `beam:"freq"`
	GlossFile string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
	DocLen int `beam:"docLen"`
	CorpusLen int `beam:"corpusLen"`
}

// addDocEntryMetaFn adds metadata for a doc entry
type addDocEntryMetaFn struct {
	CorpusEntry CorpusEntry `beam:"corpusEntry"`
}

func (f *addDocEntryMetaFn) ProcessElement(text string) DocEntry {
	return DocEntry {
		Text: text,
		GlossFile: f.CorpusEntry.GlossFile,
		ColFile: f.CorpusEntry.ColFile,
		CorpusLen: f.CorpusEntry.CorpusLen,
	}
}

// lineMap adds a key for the doc name
func lineMap(entry DocEntry, emit func(string, DocEntry)) {
	emit(entry.GlossFile, entry)
}

// concatLines is used for combining lines a doc
func concatLines(ctx context.Context, d1, d2 DocEntry) DocEntry {
	log.Infof(ctx, "concatLines, %s | %s", d1.Text, d2.Text)
	return DocEntry{
		Text: fmt.Sprintf("%s%s", d1.Text, d2.Text),
		GlossFile: d1.GlossFile,
		ColFile: d1.ColFile,
		CorpusLen: d1.CorpusLen,
	}
}

// filterFn is a DoFn for filtering out expressions that are not relevant to the term frequency, eg copyright statements.
type filterFn struct {
	// Filter is a regex identifying the terms to filter.
	Filter string `json:"filter"`
	re *regexp.Regexp
}

func (f *filterFn) Setup() {
	f.re = regexp.MustCompile(f.Filter)
}

func (f *filterFn) ProcessElement(ctx context.Context, line string, emit func(string)) {
	if !f.re.MatchString(line) {
		emit(line)
	}
}

// extractDocText reads the text from the files in a directory and returns a PCollection of <key, DocEntry>
func extractDocText(ctx context.Context, s beam.Scope, cnrHome, directory, corpusFN, filter string) beam.PCollection {

	// Get the list of files to read text from
	entries := readCorpusEntries(ctx, s, cnrHome, corpusFN)
	corpusLen := len(entries)

	// Read the text in the files line by line
	lDoc := []beam.PCollection{}
	for _, entry := range entries {
		fn := fmt.Sprintf("%s/%s", directory, entry.RawFile)
		lines :=  textio.Read(s, fn)
		filtered := beam.ParDo(s, &filterFn{Filter: filter}, lines)
		entry.CorpusLen = corpusLen
		lineEntries := beam.ParDo(s, &addDocEntryMetaFn{CorpusEntry: entry}, filtered)
		lDoc = append(lDoc, lineEntries)
	}

	// Combine the lines of text
	flattened := beam.Flatten(s, lDoc...)
	lineKV := beam.ParDo(s, lineMap, flattened)
	return beam.CombinePerKey(s, concatLines, lineKV)
}

// extractTermsFn is a DoFn that parses terms in a line of text
type extractTermsFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractTermsFn) ProcessElement(ctx context.Context, k string, entry DocEntry, emit func(string, TermFreqEntry)) {
	dt := tokenizer.DictTokenizer{
		WDict: f.Dict,
	}
	textTokens := dt.Tokenize(entry.Text)
	log.Infof(ctx, "extractTermsFn, %s | %d", entry.Text, len(textTokens))
	for _, token := range textTokens {
		charCounter.Inc(ctx, int64(len(token.Token)))
		key := fmt.Sprintf("%s:%s", token.Token, entry.GlossFile)
		emit(key, TermFreqEntry{
			Term: token.Token,
			Freq: 1,
			GlossFile: entry.GlossFile,
			ColFile: entry.ColFile,
			DocLen: len(textTokens),
			CorpusLen: entry.CorpusLen,
		})
	}
}

func extractDocFreqFn(ctx context.Context, key string, entry TermFreqEntry, emit func(string, int)) {
	emit(entry.Term, 1)
}

// sumTermFreq is used for forming a term + document key
func sumTermFreq(tf1, tf2 TermFreqEntry) TermFreqEntry {
	return TermFreqEntry{
		Term: tf1.Term,
		Freq: tf1.Freq + tf2.Freq,
		GlossFile: tf1.GlossFile,
		ColFile: tf1.ColFile,
		DocLen: tf1.DocLen,
		CorpusLen: tf1.CorpusLen,
	}
}

// extractTF is a DoFn that transforms the key in KV<string, TermFreqEntry> from term + doc to the term only
func extractTF(k string, entry TermFreqEntry, emit func(string, TermFreqEntry)) {
	emit(entry.Term, entry)
}

// readCorpusEntries read the list of entries listed in a collection file
func readCorpusEntries(ctx context.Context, s beam.Scope, cnrHome, corpusFN string) []CorpusEntry {
	fn := fmt.Sprintf("%s/%s", cnrHome, corpusFN)
	cf, err := os.Open(fn)
	if err != nil {
		log.Fatalf(ctx, "readCorpusEntries, could not open corpus file %s: %v", corpusFN, err)
	}
	defer cf.Close()
	collections, err := loadCorpusCollections(cf)
	if err != nil {
		log.Fatalf(ctx, "readCorpusEntries, could not read corpus file: %v", err)
	}
	entries := []CorpusEntry{}
	for _, col := range collections {
		colFN := fmt.Sprintf("%s/%s", cnrHome, col.CollectionFile)
		f, err := os.Open(colFN)
		if err != nil {
			log.Fatalf(ctx, "readCorpusEntries, could not open collection file %s: %v", colFN, err)
		}
		ent, err := loadCorpusEntries(f, colFN)
		f.Close()
		if err != nil {
			log.Fatalf(ctx, "readCorpusEntries, could not open read collection file %s: %v", colFN, err)
		}
		entries = append(entries, ent...)
	}
	return entries
}

// loadCorpusCollections gets the list of collections in the corpus
func loadCorpusCollections(r io.Reader) ([]CollectionEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadCorpusCollections, could not read collections: %v", err)
	}
	collections := make([]CollectionEntry, 0)
	// log.Printf("loadCorpusCollections, reading collections")
	for i, row := range rawCSVdata {
		//log.Printf("loadCorpusCollections, i = %d, len(row) = %d", i, len(row))
		if len(row) < 2 {
			return nil, fmt.Errorf("loadCorpusCollections: not enough fields in file line %d: got %d, want %d",
					i, len(row), 2)
	  }
		collections = append(collections, CollectionEntry{
			CollectionFile: row[0],
			GlossFile: row[1],
		})
	}
	return collections, nil
}

// loadCorpusEntries Get a list of documents for a collection
func loadCorpusEntries(r io.Reader, colFile string) ([]CorpusEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadCorpusEntries, unable to read corpus document: %v", err)
	}
	corpusEntries := []CorpusEntry{}
	for _, row := range rawCSVdata {
		if len(row) != 3 {
			return nil, fmt.Errorf("corpus.loadCorpusEntries len(row) != 3: %s", row)
		}
		entry := CorpusEntry{
			RawFile: row[0],
			GlossFile: row[1],
			ColFile: getColGlossFN(colFile),
		}
		corpusEntries = append(corpusEntries, entry)
	}
	return corpusEntries, nil
}

func getColGlossFN(colRawFN string) string {
	parts := strings.Split(colRawFN, "/")
	fn := parts[len(parts)-1]
	return strings.Replace(fn, ".tsv", ".html", 1)
}

// formatTFDocEntries is a DoFn that formats TermFreqEntry objects co-grouped by docFreq int's as a string.
func formatTFDocEntries(term string, tfIter func(*TermFreqEntry) bool, dfIter func(*int) bool) string {
	var e TermFreqEntry
	var docFreq int
	tfIter(&e)
	dfIter(&docFreq)
	idf := 0.0
	if docFreq > 0 {
		idf = math.Log10(float64(e.CorpusLen + 1) / float64(docFreq))
	}
	return fmt.Sprintf("%s\t%d\t%s\t%s\t%.4f\t%d", e.Term, e.Freq, e.ColFile, e.GlossFile, idf, e.DocLen)
}

// formatDFFn is a DoFn that formats document frequency as a string.
func formatDFFn(k string, freq int) string {
	return fmt.Sprintf("%s\t%d", k, freq)
}

// CountTerms processes DocEntries and outputs term frequencies in TermFreqEntry objects
func CountTerms(ctx context.Context, s beam.Scope, docs beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	c := config.InitConfig()
	dict, err := dictionary.LoadDictFile(c)
	if err != nil {
		log.Fatalf(ctx, "CountTerms, could not load dictionary: %v", err)
	}
	log.Infof(ctx, "CountTerms, loaded dictionary with %d terms", len(dict.Wdict))
	terms := beam.ParDo(s, &extractTermsFn{Dict: dict.Wdict}, docs)
	return beam.CombinePerKey(s, sumTermFreq, terms)
}

func main() {
	flag.Parse()
	beam.Init()
	ctx := context.Background()

	p := beam.NewPipeline()
	s := p.Root()

	// Compute term frequencies
	docText := extractDocText(ctx, s, *cnrHome, *input, *corpusFN, *filter)
	termFreq := CountTerms(ctx, s, docText)

	// Compute document frequencies
	dfTerms := beam.ParDo(s, extractDocFreqFn, termFreq)
	dfFormatted := beam.ParDo(s, formatDFFn, dfTerms)
	textio.Write(s, *dfDocOut, dfFormatted)
	log.Infof(ctx, "Document Frequency written to %s", *dfDocOut)

	// Combine term and document frequencies
	tfExtracted := beam.ParDo(s, extractTF, termFreq)
	dfDocTerms := beam.CoGroupByKey(s, tfExtracted, dfTerms)
	tfFormatted := beam.ParDo(s, formatTFDocEntries, dfDocTerms)
	textio.Write(s, *tfDocOut, tfFormatted)
	log.Infof(ctx, "Term Frequency per document written to %s", *tfDocOut)

	if err := beamx.Run(ctx, p); err != nil {
		log.Fatalf(ctx, "Failed to execute job: %v", err)
	}
}
