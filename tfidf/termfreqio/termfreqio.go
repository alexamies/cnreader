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

// termfreqio provides IO functions to write term frequency information to
// Firestore
package termfreqio

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/firestore"

	"github.com/alexamies/chinesenotes-go/termfreq"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	beam.RegisterType(reflect.TypeOf((*UpdateTermFreqDoc)(nil)))
}

// UpdateTermFreqDoc is a ParDo that writes entries to Firestore
type UpdateTermFreqDoc struct {
	FbCol     string
	ProjectID string
	client    *firestore.Client
}

// Setup establishes the client connection
func (f *UpdateTermFreqDoc) Setup() {
	ctx := context.Background()
	var err error
	f.client, err = firestore.NewClient(ctx, f.ProjectID)
	if err != nil {
		log.Fatalf(ctx, "Failed to create Firebase client: %v", err)
	}
}

// ProcessElement writes the given entry to Firestore if it does not exist already
func (f *UpdateTermFreqDoc) ProcessElement(ctx context.Context, entries []termfreq.TermFreqDoc) {
	for _, entry := range entries {
		doc := strings.ReplaceAll(entry.Document, "/", "_")
		key := fmt.Sprintf("%s_%s", entry.Term, doc)
		ref := f.client.Collection(f.FbCol).Doc(key)
		_, err := ref.Get(ctx)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				log.Infof(ctx, "Failed getting tf for ref %v: %v", ref, err)
				continue
			}
		}
		// update entry whether or not it exists already
		_, err = ref.Set(ctx, entry)
		if err != nil {
			log.Infof(ctx, "Failed setting tf for existing ref %v: %v", ref, err)
		}
	}
}
