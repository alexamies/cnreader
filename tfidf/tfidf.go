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
	"context"
	"flag"
	"fmt"
	"log"
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
	input  = flag.String("input", "", "File(s) to read.")
	output = flag.String("output", "", "Output file (required).")
)

func init() {
	beam.RegisterFunction(formatFn)
	beam.RegisterType(reflect.TypeOf((*extractFn)(nil)))
}

var (
	termFreq = beam.NewCounter("extract", "termFreq")
)

// extractFn is a DoFn that emits the terms in a given line and keeps the terms frequency
type extractFn struct {
	Dict map[string]*dicttypes.Word
}

func (f *extractFn) ProcessElement(ctx context.Context, line string, emit func(string)) {
	dt := tokenizer.DictTokenizer{f.Dict}
	textTokens := dt.Tokenize(line)
	for _, token := range textTokens {
		termFreq.Inc(ctx, 1)
		emit(token.Token)
	}
}

// formatFn is a DoFn that formats term count as a string.
func formatFn(w string, c int) string {
	return fmt.Sprintf("%s: %v", w, c)
}

func CountTerms(s beam.Scope, lines beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	c := config.InitConfig()
	dict, err := dictionary.LoadDictFile(c)
	if err != nil {
		log.Fatalf("CountTerms, could not load dictionary: %v", err)
	}
	log.Printf("CountTerms, loaded dictionary with %d terms", len(dict.Wdict))
	col := beam.ParDo(s, &extractFn{Dict: dict.Wdict}, lines)
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

	lines := textio.Read(s, *input)
	counted := CountTerms(s, lines)
	formatted := beam.ParDo(s, formatFn, counted)
	textio.Write(s, *output, formatted)

	if err := beamx.Run(context.Background(), p); err != nil {
		log.Fatalf("Failed to execute job: %v", err)
	}
}
