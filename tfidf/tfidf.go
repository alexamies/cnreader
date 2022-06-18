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

// tfidf is an early prototype that counts characters in a Chinese text file
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/transforms/stats"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	input    = flag.String("input", "", "Location containing documents to read.")
	corpusFN = flag.String("corpus_fn", "", "File containing list of documents to read.")
	output   = flag.String("output", "", "Output file (required).")
)

func init() {
	beam.RegisterFunction(formatFn)
	beam.RegisterType(reflect.TypeOf((*extractFn)(nil)))
}

var (
	termFreq = beam.NewCounter("extract", "termFreq")
)

// extractFn is a DoFn that emits the terms in a given text doc and keeps the term frequencies
type extractFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractFn) ProcessElement(ctx context.Context, doc string, emit func(string)) {
	dt := tokenizer.DictTokenizer{
		WDict: f.Dict,
	}
	textTokens := dt.Tokenize(doc)
	for _, token := range textTokens {
		termFreq.Inc(ctx, 1)
		emit(token.Token)
	}
}

// extractDocs reads the text from the files in a directory and returns a PCollection of documents
func extractDocs(s beam.Scope, directory, corpusFN string) beam.PCollection {
	fNames := readFileNames(s, directory, corpusFN)
	return textio.ReadAll(s, fNames)
}

func readFileNames(s beam.Scope, directory, corpusFN string) beam.PCollection {
	f, err := os.Open(corpusFN)
	if err != nil {
		log.Fatalf("readFileNames, could not open corpus file %s: %v", corpusFN, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	fNames := []string{}
	for scanner.Scan() {
		fName := fmt.Sprintf("%s%s", directory, scanner.Text())
		fNames = append(fNames, fName)
	}
	return beam.CreateList(s, fNames)
}

// formatFn is a DoFn that formats term count as a string.
func formatFn(w string, c int) string {
	return fmt.Sprintf("%s: %v", w, c)
}

func CountTerms(s beam.Scope, docs beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	c := config.InitConfig()
	dict, err := dictionary.LoadDictFile(c)
	if err != nil {
		log.Fatalf("CountTerms, could not load dictionary: %v", err)
	}
	log.Printf("CountTerms, loaded dictionary with %d terms", len(dict.Wdict))
	col := beam.ParDo(s, &extractFn{Dict: dict.Wdict}, docs)
	return stats.Count(s, col)
}

func main() {
	flag.Parse()
	beam.Init()

	if *output == "" {
		log.Fatal("No output provided")
	}

	p := beam.NewPipeline()
	s := p.Root()

	docs := extractDocs(s, *input, *corpusFN)
	tfPCol := CountTerms(s, docs)
	formatted := beam.ParDo(s, formatFn, tfPCol)
	textio.Write(s, *output, formatted)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}
