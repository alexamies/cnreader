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

// Entry point for a prototype that counts term and bigram frequency in Chinese text files
// to support computation of term frequency - inverse document frequency (TF-IDF).
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
	"strings"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/termfreq"
	"github.com/alexamies/chinesenotes-go/tokenizer"

	"github.com/alexamies/cnreader/tfidf/documentio"
	"github.com/alexamies/cnreader/tfidf/termfreqio"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/options/gcpopts"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	input    = flag.String("input", "", "Location containing documents to read.")
	cnrHome    = flag.String("cnreader_home", "..", "Top level directory to search for config files.")
	corpusFN = flag.String("corpus_fn", "", "File containing list of document collections to read.")
	corpusDataDir = flag.String("corpus_data_dir", "", "Directory containing files with list of corpus documents.")
	filter = flag.String("filter", "本作品在全世界都属于公有领域", "Regex filter pattern to use to filter out lines.")
	corpus = flag.String("corpus", "cnreader", "Firestore collection identifier")
	generation = flag.Int("generation", 0, "Firestore collection generation identifier")
)

var (
	charCounter = beam.NewCounter("extract", "charCounter")
	bigramCounter = beam.NewCounter("extract", "bigramCounter")
	docCounter = beam.NewCounter("extract", "docCounter")
	termCounter = beam.NewCounter("extract", "termCounter")
	termFreqCounter = beam.NewCounter("extract", "termFreqCounter")
	bigramFreqCounter = beam.NewCounter("extract", "bigramFreqCounter")
)

func init() {
	beam.RegisterFunction(extractDocFreqFn)
	beam.RegisterFunction(extractTF)
	beam.RegisterFunction(transformBFDocEntries)
	beam.RegisterFunction(transformTFDocEntries)
	beam.RegisterFunction(sumTermFreq)
	beam.RegisterType(reflect.TypeOf((*extractBigramsFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*extractTermsFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*DocFreqEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*TermFreqEntry)(nil)).Elem())
}

type CollectionEntry struct {
	CollectionFile, DocumentId string
}

// TermFreqEntry contains term frequency and metadata for the document that it was occurred in
type TermFreqEntry struct {
	Term string `beam:"term"`
	Freq int64 `beam:"freq"`
	DocumentId string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
	DocLen int64 `beam:"docLen"`
	CorpusLen int64 `beam:"corpusLen"`
}

// DocFreqEntry contains document frequency and metadata so that it can be combined with term frequency data
type DocFreqEntry struct {
	Term string `beam:"term"`
	DocFreq int64 `beam:"freq"`
}

// extractDocText reads the text from the files in a directory and returns a PCollection of <DocEntry>
func extractDocText(ctx context.Context, s beam.Scope, cnrHome, input, corpusFN, filter string) beam.PCollection {

	// Get the list of files to read text from
	entries := readCorpusEntries(ctx, s, cnrHome, corpusFN)
	corpusLen := len(entries)
	entriesPCol := beam.CreateList(s, entries)
	return documentio.Read(ctx, s, input, corpusLen, filter, entriesPCol)
}

// extractTermsFn is a DoFn that parses terms in a string of text
type extractTermsFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractTermsFn) ProcessElement(ctx context.Context, entry documentio.DocEntry, emit func(string, TermFreqEntry)) {
	dt := tokenizer.DictTokenizer{
		WDict: f.Dict,
	}
	textTokens := dt.Tokenize(entry.Text)
	tokens := []tokenizer.TextToken{}
	for _, token := range textTokens {
		if dicttypes.IsCJKChar(token.Token) {
			tokens = append(tokens, token)
			charCounter.Inc(ctx, int64(len(token.Token)))
		}
	}
	docLen := int64(len(tokens))
	termCounter.Inc(ctx, docLen)
	// log.Infof(ctx, "extractTermsFn, %s | %d", entry.Text, docLen)
	for _, token := range tokens {
		key := fmt.Sprintf("%s:%s", token.Token, entry.DocumentId)
		emit(key, TermFreqEntry{
			Term: token.Token,
			Freq: 1,
			DocumentId: entry.DocumentId,
			ColFile: entry.ColFile,
			DocLen: docLen,
			CorpusLen: int64(entry.CorpusLen),
		})
	}
}

func extractDocFreqFn(ctx context.Context, key string, entry TermFreqEntry, emit func(string, int)) {
	// log.Infof(ctx, "extractDocFreqFn key: %s", key)
	emit(entry.Term, 1)
}

// sumTermFreq is used for forming a term + document key
func sumTermFreq(tf1, tf2 TermFreqEntry) TermFreqEntry {
	return TermFreqEntry{
		Term: tf1.Term,
		Freq: tf1.Freq + tf2.Freq,
		DocumentId: tf1.DocumentId,
		ColFile: tf1.ColFile,
		DocLen: tf1.DocLen,
		CorpusLen: tf1.CorpusLen,
	}
}

// sumDocFreq is used for document frequency
func sumDocFreq(m, n int) int {
	return m + n
}

// extractTF is a DoFn that transforms the key in KV<string, TermFreqEntry> from term + doc to the term only
func extractTF(ctx context.Context, k string, entry TermFreqEntry, emit func(string, TermFreqEntry)) {
	// log.Infof(ctx, "extractTF k: %s", k)
	emit(entry.Term, entry)
}

// readCorpusEntries read the list of entries listed in a collection file
func readCorpusEntries(ctx context.Context, s beam.Scope, cnrHome, corpusFN string) []documentio.CorpusEntry {
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
	entries := []documentio.CorpusEntry{}
	for _, col := range collections {
		colFN := fmt.Sprintf("%s/%s", cnrHome, col.CollectionFile)
		if len(*corpusDataDir) > 0 {
			colFN = fmt.Sprintf("%s/%s/%s", cnrHome, *corpusDataDir, col.CollectionFile)
		}
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
	log.Infof(ctx, "readCorpusEntries, got %d entries", len(entries))
	docCounter.Inc(ctx, int64(len(entries)))
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
			DocumentId: row[1],
		})
	}
	return collections, nil
}

// loadCorpusEntries Get a list of documents for a collection
func loadCorpusEntries(r io.Reader, colFile string) ([]documentio.CorpusEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadCorpusEntries, unable to read corpus document: %v", err)
	}
	corpusEntries := []documentio.CorpusEntry{}
	for _, row := range rawCSVdata {
		if len(row) != 3 {
			return nil, fmt.Errorf("corpus.loadCorpusEntries len(row) != 3: %s", row)
		}
		entry := documentio.CorpusEntry{
			RawFile: row[0],
			DocumentId: row[1],
			ColFile: getColGlossFN(colFile),
		}
		corpusEntries = append(corpusEntries, entry)
	}
	return corpusEntries, nil
}

func getColGlossFN(colRawFN string) string {
	parts := strings.Split(colRawFN, "/")
	fn := parts[len(parts)-1]
	fn = strings.Replace(fn, ".tsv", ".html", 1)
	return strings.Replace(fn, ".csv", ".html", 1)
}

// transformTFDocEntries is a DoFn that transforms a TermFreqEntry object to a termfreq.TermFreqDoc
func transformTFDocEntries(ctx context.Context, key string, tfIter func(*TermFreqEntry) bool, dfIter func(*int) bool) termfreq.TermFreqDoc {
	// log.Infof(ctx, "transformTFDocEntries, transforming key: %s", key)
	tf := termfreq.TermFreqDoc{
		Term: "",
    Freq: 0,
    Collection: "",
    Document: "",
    IDF: 0.0,
    DocLen: 0,
	}
	var e TermFreqEntry
	var corpusLen int64 = 0
	for tfIter(&e) {
		// log.Infof(ctx, "transformTFDocEntries, key: %s, term: %s, doc: %s, freq: %d", key, e.Term, e.DocumentId, e.Freq)
		tf.Term = e.Term
		tf.Freq += e.Freq
		tf.Collection = e.ColFile
		tf.Document = e.DocumentId
		tf.DocLen = e.DocLen
		corpusLen = e.CorpusLen
	}
	var df int
	dfIter(&df)
	if df > 0 {
		tf.IDF = math.Log10(float64(corpusLen + 1) / float64(df))
	}
	// log.Infof(ctx, "transformTFDocEntries for key %s, docFreq: %d, IDF: %.3f", key, df, tf.IDF)
	termFreqCounter.Inc(ctx, 1)
	return tf
}

// transformBFDocEntries is a DoFn that transforms a bigram TermFreqEntry object to a termfreq.TermFreqDoc
func transformBFDocEntries(ctx context.Context, term string, tfIter func(*TermFreqEntry) bool, dfIter func(*int) bool) termfreq.TermFreqDoc {
	tf := termfreq.TermFreqDoc{
		Term: "",
    Freq: 0,
    Collection: "",
    Document: "",
    IDF: 0.0,
    DocLen: 0,
	}
	var e TermFreqEntry
	var corpusLen int64 = 0
	for tfIter(&e) {
		tf.Term = e.Term
		tf.Freq += e.Freq
		tf.Collection = e.ColFile
		tf.Document = e.DocumentId
		tf.DocLen = e.DocLen
		corpusLen = e.CorpusLen
	}
	var df int
	dfIter(&df)
	if df > 0 {
		tf.IDF = math.Log10(float64(corpusLen + 1) / float64(df))
	}
	// log.Infof(ctx, "transformBFDocEntries for term %s, docFreq: %d, corpusLen: %d, IDF: %.3f", term, df, corpusLen, tf.IDF)
	bigramFreqCounter.Inc(ctx, 1)
	return tf
}

// CountTerms processes DocEntries and outputs term frequencies in TermFreqEntry objects
func CountTerms(ctx context.Context, s beam.Scope, wdict map[string]*dicttypes.Word, docs beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	terms := beam.ParDo(s, &extractTermsFn{Dict: wdict}, docs)
	return beam.CombinePerKey(s, sumTermFreq, terms)
}

// extractBigramsFn is a DoFn that parses bigrams in a string of text
type extractBigramsFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractBigramsFn) ProcessElement(ctx context.Context, entry documentio.DocEntry, emit func(string, TermFreqEntry)) {
	dt := tokenizer.DictTokenizer{
		WDict: f.Dict,
	}
	textTokens := dt.Tokenize(entry.Text)
	// log.Infof(ctx, "extractBigramsFn, %s | %d", entry.Text, len(textTokens))
	lastToken := ""
	bigrams := []string{}
	for _, token := range textTokens {
		if len(lastToken) == 0 {
			lastToken = token.Token
			continue
		}
		if !dicttypes.IsCJKChar(token.Token) {
			lastToken = ""
			continue
		}
		bigram := fmt.Sprintf("%s%s", lastToken, token.Token)
		bigrams = append(bigrams, bigram)
		lastToken = token.Token
		bigramCounter.Inc(ctx, 1)
	}
	for _, bigram := range bigrams {
		key := fmt.Sprintf("%s:%s", bigram, entry.DocumentId)
		emit(key, TermFreqEntry{
			Term: bigram,
			Freq: 1,
			DocumentId: entry.DocumentId,
			ColFile: entry.ColFile,
			DocLen: int64(len(bigrams)),
			CorpusLen: int64(entry.CorpusLen),
		})
	}
}

// CountBigrams processes DocEntries and outputs term frequencies in TermFreqEntry objects
func CountBigrams(ctx context.Context, s beam.Scope, wdict map[string]*dicttypes.Word, docs beam.PCollection) beam.PCollection {
	s = s.Scope("CountBigrams")
	terms := beam.ParDo(s, &extractBigramsFn{Dict: wdict}, docs)
	return beam.CombinePerKey(s, sumTermFreq, terms)
}

func main() {
	flag.Parse()
	beam.Init()
	ctx := context.Background()

	// Dictionary initialization
	c := config.InitConfig()
	dict, err := dictionary.LoadDictFile(c)
	if err != nil {
		log.Fatalf(ctx, "main, could not load dictionary: %v", err)
	}
	log.Infof(ctx, "main, loaded dictionary with %d terms", len(dict.Wdict))
	project := gcpopts.GetProjectFromFlagOrEnvironment(ctx)
	log.Infof(ctx, "project: %s", project)

	// Create pipeline
	p := beam.NewPipeline()
	s := p.Root()

	// Compute term frequencies
	docText := extractDocText(ctx, s, *cnrHome, *input, *corpusFN, *filter)
	termFreq := CountTerms(ctx, s, dict.Wdict, docText)

	// Compute document frequencies
	dfListTerms := beam.ParDo(s, extractDocFreqFn, termFreq)
	dfTerms := beam.CombinePerKey(s, sumDocFreq, dfListTerms)

	// Combine term and document frequencies
	tfExtracted := beam.ParDo(s, extractTF, termFreq)
	dfGroupedTerms := beam.CoGroupByKey(s, tfExtracted, dfTerms)

	// Write term frequencies to Firestore
	tfDoc := beam.ParDo(s, transformTFDocEntries, dfGroupedTerms)
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", *corpus, *generation)
	uf := termfreqio.UpdateTermFreqDoc{
		FbCol: fbCol,
		ProjectID: project,
	}
	beam.ParDo0(s, &uf, tfDoc)

	// Compute bigram frequencies
	bigramFreq := CountBigrams(ctx, s, dict.Wdict, docText)

	// Compute bigram document frequencies
	dfListBigrams := beam.ParDo(s, extractDocFreqFn, bigramFreq)
	dfBigrams := beam.CombinePerKey(s, sumDocFreq, dfListBigrams)

	// Combine bigram and document frequencies
	bfExtracted := beam.ParDo(s, extractTF, bigramFreq)
	dfGroupedBigrams := beam.CoGroupByKey(s, bfExtracted, dfBigrams)

	// Write bigram frequencies to Firestore
	bfDoc := beam.ParDo(s, transformBFDocEntries, dfGroupedBigrams)
	fbBigramCol := fmt.Sprintf("%s_bigram_doc_freq%d", *corpus, *generation)
	bf := termfreqio.UpdateTermFreqDoc{
		FbCol: fbBigramCol,
		ProjectID: project,
	}
	beam.ParDo0(s, &bf, bfDoc)

	if err := beamx.Run(ctx, p); err != nil {
		log.Fatalf(ctx, "Failed to execute job: %v", err)
	}
	log.Infof(ctx, "Term Frequency per document written to Firestore collection %s", fbCol)
	log.Infof(ctx, "Bigram Frequency per document written to Firestore collection %s", fbBigramCol)
}
