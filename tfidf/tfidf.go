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
	"os"
	"reflect"
	"regexp"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"

	"github.com/alexamies/cnreader/corpus"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/transforms/stats"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	input    = flag.String("input", "", "Location containing documents to read.")
	corpusFN = flag.String("corpus_fn", "", "File containing list of documents to read.")
	filter = flag.String("filter", "本作品在全世界都属于公有领域", "Regex filter pattern to use to filter out lines.")
	output   = flag.String("output", "", "Output file (required).")
)

func init() {
	beam.RegisterFunction(formatFn)
	beam.RegisterType(reflect.TypeOf((*extractFn)(nil)))
	beam.RegisterType(reflect.TypeOf((*filterFn)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*addFileNameFn)(nil)).Elem())
}

var (
	charFreq = beam.NewCounter("extract", "charFreq")
)

type addFileNameFn struct {
	FileName string `json:"filename"`
}

func (f *addFileNameFn) ProcessElement(x beam.X) (string, beam.X) {
	return f.FileName, x
}

// extractFn is a DoFn that emits the terms in a given text doc and keeps the term frequencies
type extractFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractFn) ProcessElement(ctx context.Context, fileName, line string, emit func(string)) {
	dt := tokenizer.DictTokenizer{
		WDict: f.Dict,
	}
	textTokens := dt.Tokenize(line)
	for _, token := range textTokens {
		charFreq.Inc(ctx, int64(len(token.Token)))
		emit(fmt.Sprintf("%s\t%s", token.Token, fileName))
	}
}

// filterFn is a DoFn for filtering out expressions that are not relevant to the text.
type filterFn struct {
	// Filter is a regex identifying the terms to filter.
	Filter string `json:"filter"`
	re *regexp.Regexp
}

func (f *filterFn) Setup() {
	f.re = regexp.MustCompile(f.Filter)
}

func (f *filterFn) ProcessElement(ctx context.Context, fileName, line string, emit func(string, string)) {
	if !f.re.MatchString(line) {
		emit(fileName, line)
	}
}

// extractLines reads the text from the files in a directory and returns a PCollection of lines
func extractLines(ctx context.Context, s beam.Scope, directory, corpusFN string) beam.PCollection {
	entries := readCorpusEntries(ctx, s, corpusFN)
	lDoc := []beam.PCollection{}
	for _, entry := range entries {
		fn := fmt.Sprintf("%s/%s", directory, entry.RawFile)
		lines :=  textio.Read(s, fn)
		ld := beam.ParDo(s, &addFileNameFn{FileName: entry.GlossFile}, lines)
		lDoc = append(lDoc, ld)
	}
	return beam.Flatten(s, lDoc...)
}

// readCorpusEntryList read the list of entries listed in a collection file
func readCorpusEntries(ctx context.Context, s beam.Scope, collectionFN string) []corpus.CorpusEntry {
	f, err := os.Open(collectionFN)
	if err != nil {
		log.Fatalf(ctx, "readCorpusEntryList, could not open collection file %s: %v", collectionFN, err)
	}
	defer f.Close()
	entries, err := loadCorpusEntries(f, "Sample Collection", collectionFN)
	if err != nil {
		log.Fatalf(ctx, "readCorpusEntryList, could not open read collection file %s: %v", collectionFN, err)
	}
	return entries
}

// loadCorpusEntries Get a list of files for a collection
func loadCorpusEntries(r io.Reader, colTitle, colFile string) ([]corpus.CorpusEntry, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("loadCorpusEntries, unable to read corpus document: %v", err)
	}
	corpusEntries := []corpus.CorpusEntry{}
	for _, row := range rawCSVdata {
		if len(row) != 3 {
			return nil, fmt.Errorf("corpus.loadCorpusEntries len(row) != 3: %s", row)
		}
		entry := corpus.CorpusEntry{
			RawFile: row[0],
			GlossFile: row[1],
			Title: row[2],
			ColTitle: colTitle,
			ColFile: colFile,
		}
		corpusEntries = append(corpusEntries, entry)
	}
	return corpusEntries, nil
}

// formatFn is a DoFn that formats term count as a string.
func formatFn(w string, c int) string {
	return fmt.Sprintf("%s: %v", w, c)
}

func CountTerms(ctx context.Context, s beam.Scope, docs beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	c := config.InitConfig()
	dict, err := dictionary.LoadDictFile(c)
	if err != nil {
		log.Fatalf(ctx, "CountTerms, could not load dictionary: %v", err)
	}
	log.Infof(ctx, "CountTerms, loaded dictionary with %d terms", len(dict.Wdict))
	terms := beam.ParDo(s, &extractFn{Dict: dict.Wdict}, docs)
	return stats.Count(s, terms)
}

func main() {
	flag.Parse()
	beam.Init()
	ctx := context.Background()

	if *output == "" {
		log.Fatal(ctx, "No output provided")
	}

	p := beam.NewPipeline()
	s := p.Root()

	lines := extractLines(ctx, s, *input, *corpusFN)
	filtered := beam.ParDo(s, &filterFn{Filter: *filter}, lines)
	termFreq := CountTerms(ctx, s, filtered)
	tfFormatted := beam.ParDo(s, formatFn, termFreq)
	textio.Write(s, *output, tfFormatted)

	if err := beamx.Run(ctx, p); err != nil {
		log.Fatalf(ctx, "Failed to execute job: %v", err)
	}

}
