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

package tmindex

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/transmemory"
)

// BuildIndexesFS builds a unigram indexes and save in Firestore
func BuildIndexesFS(ctx context.Context, client dictionary.FsClient, wdict map[string]*dicttypes.Word, corpus string, generation int) error {
	err := buildUniDomainIndexFS(ctx, client, wdict, corpus, generation)
	if err != nil {
		return fmt.Errorf("tmindex.BuildIndexesFS, could not write index, err: %v", err)
	}
	err = buildUnigramIndexFS(ctx, client, wdict, corpus, generation)
	if err != nil {
		return fmt.Errorf("tmindex.BuildIndexesFS, could not write index, err: %v", err)
	}
	return nil
}

// buildUniDomainIndexFS builds a unigram index with domain
func buildUniDomainIndexFS(ctx context.Context, client dictionary.FsClient, wdict map[string]*dicttypes.Word, corpus string, generation int) error {
	fsCol := fmt.Sprintf("%s_transmemory_dom_%d", corpus, generation)
	tmindexDom := make(map[string]bool)
	i := 0
	for term, word := range wdict {
		n := len([]rune(term))
		if n < minChars || n > maxChars {
			continue
		}
		for _, sense := range word.Senses {
			chars := strings.Split(term, "")
			for _, c := range chars {
		  	key := fmt.Sprintf("%s_%s_%s", c, term, sense.Domain)
				if _, ok := tmindexDom[key]; ok {
					continue
				}
				rec := transmemory.TMIndexDomain{
					Ch: c,
					Word: term,
					Domain: sense.Domain,
				}
				ref := client.Collection(fsCol).Doc(key)
				_, err := ref.Get(ctx)
				if err != nil {
					if status.Code(err) != codes.NotFound {
						return fmt.Errorf("UpdateDocTitleIndex, Failed getting record for ref %v: %v", ref, err)
					}
					_, err = ref.Set(ctx, rec)
	  			if err != nil {
	  				return err
	  			}
		  		tmindexDom[key] = true
					i++
					if i % 100 == 0 {
						log.Printf("tmindex.buildUniDomainIndexFS processed %d entries so far", i)
					}
				}
			}
		}
	}
	log.Printf("tmindex.buildUniDomainIndexFS processed %d entries in total", i)
	return nil
}

// buildUnigramIndexFS builds a unigram index, skip for strings > maxChars
func buildUnigramIndexFS(ctx context.Context, client dictionary.FsClient, wdict map[string]*dicttypes.Word, corpus string, generation int) error {
	fsCol := fmt.Sprintf("%s_transmemory_uni_%d", corpus, generation)
	tmindexUni := make(map[string]bool)
	i := 0
	for term := range wdict {
		n := len([]rune(term))
		if n < minChars || n > maxChars {
			continue
		}
		chars := strings.Split(term, "")
		for _, c := range chars {
	  	key := fmt.Sprintf("%s_%s", c, term)
			if _, ok := tmindexUni[key]; ok {
				continue
			}
			rec := transmemory.TMIndexUnigram{
				Ch: c,
				Word: term,
			}
			ref := client.Collection(fsCol).Doc(key)
			_, err := ref.Get(ctx)
			if err != nil {
				if status.Code(err) != codes.NotFound {
					return fmt.Errorf("UpdateDocTitleIndex, Failed getting record for ref %v: %v", ref, err)
				}
				_, err = ref.Set(ctx, rec)
	  		if err != nil {
	  			return err
	  		}
	  		tmindexUni[key] = true
				i++
				if i % 100 == 0 {
					log.Printf("tmindex.buildUnigramIndexFS processed %d entries so far", i)
				}
			}
		}
	}
	log.Printf("tmindex.buildUnigramIndexFS processed %d entries in total", i)
	return nil
}