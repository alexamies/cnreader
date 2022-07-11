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

package termfreqio

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/alexamies/chinesenotes-go/termfreq"
)

var (
	projectID = flag.String("project_id", "", "GCP project ID")
)

func TestProcessElement(t *testing.T) {
	corpus := "cnreader"
	generation := 0
	fbCol := fmt.Sprintf("%s_wordfreqdoc%d", corpus, generation)
	f := UpdateTermFreqDoc{
		FbCol:     fbCol,
		ProjectID: *projectID,
	}
	f.Setup()
	ctx := context.Background()
	entries := []termfreq.TermFreqDoc{
		{
			Term:       "而",
			Freq:       1,
			Collection: "testcollection.html",
			Document:   "testdata/sampletest2.html",
			IDF:        0.4771,
			DocLen:     7,
		},
		{
			Term:       "而",
			Freq:       1,
			Collection: "testcollection.html",
			Document:   "testdata/sampletest3.html",
			IDF:        0.4771,
			DocLen:     3,
		},
	}
	f.ProcessElement(ctx, entries)
}
