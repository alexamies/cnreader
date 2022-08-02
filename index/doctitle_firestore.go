// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Functions for building and accessing the document title index.

package index

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/alexamies/chinesenotes-go/find"
	"github.com/alexamies/cnreader/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FsClient defines Firestore interfaces needed
type FsClient interface {
	Collection(path string) *firestore.CollectionRef
}

// UpdateDocTitleIndex writes a flast list of document titles from the
// hierarchical corpus.
func UpdateDocTitleIndex(ctx context.Context, libLoader library.LibraryLoader, client FsClient, corpus string, generation int) error {
	fsCol := fmt.Sprintf("%s_doc_title_%d", corpus, generation)
	corpLoader := libLoader.GetCorpusLoader()
	collections, err := corpLoader.LoadCollections()
	if err != nil {
		return fmt.Errorf("UpdateDocTitleIndex could not load corpus: %v", err)
	}
	log.Printf("UpdateDocTitleIndex loaded %d collections", len(*collections))
	for _, c := range *collections {
		log.Printf("UpdateDocTitleIndex, collection: %s\n", c.Title)
		docs, err := corpLoader.LoadCollection(c.CollectionFile, c.Title)
		if err != nil {
			return fmt.Errorf("UpdateDocTitleIndex error loading %s: %v", c.CollectionFile, err)
		}
		for _, d := range *docs {
			titleZh, titleEn := splitTitle(d.Title)
			title := c.Title + ": " + d.Title
			colTitleZh, _ := splitTitle(c.Title)
			record := find.DocTitleRecord{
				RawFile:         d.RawFile,
				GlossFile:       d.GlossFile,
				DocTitle:        d.Title,
				DocTitleZh:      titleZh,
				DocTitleEn:      titleEn,
				ColGlossFile:    c.GlossFile,
				ColTitle:        c.Title,
				ColPlusDocTitle: title,
				Substrings:      titleSubtrings(colTitleZh, titleZh),
			}
			ref := client.Collection(fsCol).Doc(title)
			_, err := ref.Get(ctx)
			if err != nil {
				if status.Code(err) != codes.NotFound {
					log.Printf("UpdateDocTitleIndex, Failed getting tf for ref %v: %v", ref, err)
				} else {
					_, err = ref.Set(ctx, record)
					if err != nil {
						log.Printf("Failed setting tf for ref %v: %v", ref, err)
					}
				}
			} // else do not update entry if it exists already
		}
	}
	return nil
}

// titleSubtrings finds the union of all the substrings both input strings
func titleSubtrings(titleZh, colTitleZh string) []string {
	s1 := strings.Split(titleZh, "")
	s2 := strings.Split(colTitleZh, "")
	ss := append(s1, s2...)
	return substrings(ss)
}

func substrings(chars []string) []string {
	if len(chars) == 0 {
		return []string{}
	}
	if len(chars) == 1 {
		return []string{chars[0]}
	}
	whole := []string{strings.Join(chars, "")}
	first := chars[0]
	ss := append(whole, first)
	rest := substrings(chars[1:])
	return append(ss, rest...)
}
