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
	corpusFN = flag.String("corpus_fn", "", "File containing list of documents to read.")
	filter = flag.String("filter", "本作品在全世界都属于公有领域", "Regex filter pattern to use to filter out lines.")
	tfDocOut   = flag.String("tfdoc_out", "word_freq_doc.txt", "Term frequency per document output file")
	dfDocOut   = flag.String("df_out", "doc_freq.txt", "Document frequency output file")
)

var (
	charCounter = beam.NewCounter("extract", "charCounter")
)

func init() {
	beam.RegisterFunction(extractDocFreqFn)
	beam.RegisterFunction(formatFn)
	beam.RegisterFunction(formatDFFn)
	beam.RegisterFunction(sumTermFreq)
	beam.RegisterType(reflect.TypeOf((*addLineEntryMetaFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*extractFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*filterFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*CorpusEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*LineEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*TermFreqEntry)(nil)).Elem())
}

// CorpusEntry contains metadata for a document that text will be read from
type CorpusEntry struct {
	RawFile string `beam:"rawFile"`
	GlossFile string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
	CorpusLen int `beam:"corpusLen"`
}

// LineEntry contains a line of text and metadata for the document that it was extracted from
type LineEntry struct {
	Line string `beam:"line"`
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
	CorpusLen int `beam:"corpusLen"`
	DocFreq int `beam:"docLen"`
}

// addLineEntryMetaFn adds metadata for an entry
type addLineEntryMetaFn struct {
	CorpusEntry CorpusEntry `beam:"corpusEntry"`
}

func (f *addLineEntryMetaFn) ProcessElement(line string) LineEntry {
	return LineEntry {
		Line: line,
		GlossFile: f.CorpusEntry.GlossFile,
		ColFile: f.CorpusEntry.ColFile,
		CorpusLen: f.CorpusEntry.CorpusLen,
	}
}

// extractFn is a DoFn that parses terms in a line of text
type extractFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractFn) ProcessElement(ctx context.Context, entry LineEntry, emit func(string, TermFreqEntry)) {
	dt := tokenizer.DictTokenizer{
		WDict: f.Dict,
	}
	textTokens := dt.Tokenize(entry.Line)
	for _, token := range textTokens {
		charCounter.Inc(ctx, int64(len(token.Token)))
		key := fmt.Sprintf("%s:%s", token.Token, entry.GlossFile)
		emit(key, TermFreqEntry{
			Term: token.Token,
			Freq: 1,
			GlossFile: entry.GlossFile,
			ColFile: entry.ColFile,
			CorpusLen: entry.CorpusLen,
		})
	}
}

func extractDocFreqFn(ctx context.Context, key string, entry TermFreqEntry, emit func(string, int)) {
	emit(entry.Term, 1)
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

// sumTermFreq is used for forming a term + document key
func sumTermFreq(tf1, tf2 TermFreqEntry) TermFreqEntry {
	return TermFreqEntry{
		Term: tf1.Term,
		Freq: tf1.Freq + tf2.Freq,
		GlossFile: tf1.GlossFile,
		ColFile: tf1.ColFile,
		CorpusLen: tf1.CorpusLen,
	}
}

// extractLines reads the text from the files in a directory and returns a PCollection of LineEntry
func extractLines(ctx context.Context, s beam.Scope, directory, corpusFN, filter string) beam.PCollection {
	entries := readCorpusEntries(ctx, s, corpusFN)
	corpusLen := len(entries)
	lDoc := []beam.PCollection{}
	for _, entry := range entries {
		fn := fmt.Sprintf("%s/%s", directory, entry.RawFile)
		lines :=  textio.Read(s, fn)
		filtered := beam.ParDo(s, &filterFn{Filter: filter}, lines)
		entry.CorpusLen = corpusLen
		lineEntries := beam.ParDo(s, &addLineEntryMetaFn{CorpusEntry: entry}, filtered)
		lDoc = append(lDoc, lineEntries)
	}
	return beam.Flatten(s, lDoc...)
}

// readLineEntryList read the list of entries listed in a collection file
func readCorpusEntries(ctx context.Context, s beam.Scope, collectionFN string) []CorpusEntry {
	f, err := os.Open(collectionFN)
	if err != nil {
		log.Fatalf(ctx, "readLineEntryList, could not open collection file %s: %v", collectionFN, err)
	}
	defer f.Close()
	entries, err := loadCorpusEntries(f, "Sample Collection", collectionFN)
	if err != nil {
		log.Fatalf(ctx, "readLineEntryList, could not open read collection file %s: %v", collectionFN, err)
	}
	return entries
}

// loadCorpusEntries Get a list of documents for a collection
func loadCorpusEntries(r io.Reader, colTitle, colFile string) ([]CorpusEntry, error) {
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

// formatFn is a DoFn that formats term frequency as a string.
func formatFn(k string, e TermFreqEntry) string {
	idf := 0.0
	if e.DocFreq > 0 {
		idf = math.Log10(float64(e.CorpusLen + 1) / float64(e.DocFreq))
	}
	return fmt.Sprintf("%s\t%d\t%s\t%s\t%.4f", e.Term, e.Freq, e.ColFile, e.GlossFile, idf)
}

// formatDFFn is a DoFn that formats document frequency as a string.
func formatDFFn(k string, freq int) string {
	return fmt.Sprintf("%s\t%d", k, freq)
}

// CountTerms processes lines and outputs term frequencies in TermFreqEntry objects
func CountTerms(ctx context.Context, s beam.Scope, lines beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	c := config.InitConfig()
	dict, err := dictionary.LoadDictFile(c)
	if err != nil {
		log.Fatalf(ctx, "CountTerms, could not load dictionary: %v", err)
	}
	log.Infof(ctx, "CountTerms, loaded dictionary with %d terms", len(dict.Wdict))
	terms := beam.ParDo(s, &extractFn{Dict: dict.Wdict}, lines)
	return beam.CombinePerKey(s, sumTermFreq, terms)
}



func main() {
	flag.Parse()
	beam.Init()
	ctx := context.Background()

	p := beam.NewPipeline()
	s := p.Root()

	// Compute term frequencies
	lines := extractLines(ctx, s, *input, *corpusFN, *filter)
	termFreq := CountTerms(ctx, s, lines)
	tfFormatted := beam.ParDo(s, formatFn, termFreq)
	textio.Write(s, *tfDocOut, tfFormatted)
	log.Infof(ctx, "Term Frequency per document written to %s", *tfDocOut)

	// Compute document frequencies
	dfTerms := beam.ParDo(s, extractDocFreqFn, termFreq)
	dfFormatted := beam.ParDo(s, formatDFFn, dfTerms)
	textio.Write(s, *dfDocOut, dfFormatted)
	log.Infof(ctx, "Document Frequency written to %s", *dfDocOut)

	if err := beamx.Run(ctx, p); err != nil {
		log.Fatalf(ctx, "Failed to execute job: %v", err)
	}
}
