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

// tfidf provides IO functions to read text files and also remembers the source
// of the text.
package documentio

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"
)

func init() {
	beam.RegisterType(reflect.TypeOf((*CorpusEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*DocEntry)(nil)).Elem())
	beam.RegisterType(reflect.TypeOf((*readFn)(nil)))
}

// CorpusEntry contains metadata for a document that text will be read from
type CorpusEntry struct {
	RawFile string `beam:"rawFile"`
	DocumentId string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
}

// DocEntry contains all text and metadata for the document that it was extracted from
type DocEntry struct {
	Text string `beam:"text"`
	DocumentId string `beam:"glossFile"`
	ColFile string `beam:"colFile"`
	CorpusLen int `beam:"corpusLen"`
}

// Read a PCollection of CorpusEntry and return a PCollection of DocEntry with the
// contents of the file contained in the Text field.
func Read(ctx context.Context, s beam.Scope, input string, corpusLen int, filter string, entries beam.PCollection) beam.PCollection {
	s = s.Scope("tfidf.Read")
	return beam.ParDo(s, &readFn{Input: input, CorpusLen: corpusLen, Filter: filter}, entries)
}


// readFn is a DoFn for reading files text and filtering out expressions that are not relevant
// to the term frequency, eg copyright statements.
type readFn struct {
	// Input is either a file path or GCS path to read the corpus entries from
	Input string
	// Filter is a regex identifying the terms to filter.
	Filter string `json:"filter"`
	CorpusLen int `beam:"corpusLen"`
	re *regexp.Regexp
	client *storage.Client
}

func (f *readFn) Setup() {
	f.re = regexp.MustCompile(f.Filter)

	ctx := context.Background()
	var err error
	f.client, err = storage.NewClient(ctx, option.WithScopes(storage.ScopeReadWrite))
	if err != nil {
		log.Fatalf(ctx, "Unable to access GCS: %v", err)
	}
}

func (f *readFn) ProcessElement(ctx context.Context, entry CorpusEntry, emit func(DocEntry)) {
	log.Infof(ctx, "Reading text from %v", entry.RawFile)

	fname := fmt.Sprintf("%s/%s", f.Input, entry.RawFile)
	fd, err := openRead(ctx, f.client, fname)
	if err != nil {
		log.Fatalf(ctx, "readFn could not open %s: %v", fname, err)
	}
	defer fd.Close()

	r := bufio.NewReader(fd)
	var text strings.Builder
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			if len(line) != 0 && !f.re.MatchString(line) {
				text.WriteString(line)
			}
			break
		}
		if err != nil {
			log.Fatalf(ctx, "readFn could not read %s: %v", fname, err)
		}
		text.WriteString(line)
	}
	emit(DocEntry {
		Text: text.String(),
		DocumentId: entry.DocumentId,
		ColFile: entry.ColFile,
		CorpusLen: f.CorpusLen,
	})
}

func openRead(ctx context.Context, client *storage.Client, filename string) (io.ReadCloser, error) {
	scheme := getScheme(filename)
	if scheme == "gs" {
		bucket, object, err := parseObject(filename)
		if err != nil {
			return nil, err
		}
		return client.Bucket(bucket).Object(object).NewReader(ctx)
	}
	return os.Open(filename)
}

func getScheme(fname string) string {
	if i := strings.Index(fname, "://"); i > 0 {
		return fname[:i]
	}
	return "file"
}

// parseObject extracts bucket and object name
func parseObject(object string) (bucket, path string, err error) {
	p, err := url.Parse(object)
	if err != nil {
		return "", "", err
	}
	return p.Host, p.Path[1:], nil
}