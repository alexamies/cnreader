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
	"strings"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/textio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/transforms/stats"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/x/beamx"
)

var (
	input = flag.String("input", "", "File(s) to read.")
	output = flag.String("output", "", "Output file (required).")
)

func init() {
	beam.RegisterFunction(formatFn)
	beam.RegisterType(reflect.TypeOf((*extractFn)(nil)))
}

var (
	empty           = beam.NewCounter("extract", "emptyLines")
	smallWordLength = flag.Int("small_word_length", 9, "length of small words (default: 9)")
	smallWords      = beam.NewCounter("extract", "smallWords")
	lineLen         = beam.NewDistribution("extract", "lineLenDistro")
)

// extractFn is a DoFn that emits the terms in a given line and keeps a count for small terms.
type extractFn struct {}

func (f *extractFn) ProcessElement(ctx context.Context, line string, emit func(string)) {
	lineLen.Update(ctx, int64(len(line)))
	if len(strings.TrimSpace(line)) == 0 {
		empty.Inc(ctx, 1)
	}
	for _, word := range tokenize(line) {
		emit(word)
	}
}

// formatFn is a DoFn that formats term count as a string.
func formatFn(w string, c int) string {
	return fmt.Sprintf("%s: %v", w, c)
}

func CountTerms(s beam.Scope, lines beam.PCollection) beam.PCollection {
	s = s.Scope("CountTerms")
	col := beam.ParDo(s, &extractFn{}, lines)
	return stats.Count(s, col)
}

func tokenize(text string) []string {
	return strings.Split(text, "")	
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
