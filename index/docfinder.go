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

// Library for documents retrieval
package index

import (
	"log"

	"github.com/alexamies/chinesenotes-go/bibnotes"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/cnreader/corpus"
)

// Maximum number of documents displayed on a web page
const maxDocsDisplayed = 10

// A document-specific word frequency entry record
type RetrievalResult struct {
	HTMLFile, Title, ColTitle string
	Count int
	HasEngTrans bool
}

// IndexConfig encapsulates parameters for index configuration
type IndexConfig struct {
	IndexDir string
}

// Retrieves documents with title for a single keyword
func FindDocsForKeyword(keyword dicttypes.Word,
		corpusEntryMap map[string]corpus.CorpusEntry,
		indexState IndexState, bibnotesClient bibnotes.BibNotesClient) []RetrievalResult {
	docs := make([]RetrievalResult, 0)
	if indexState.wfdoc == nil {
		log.Print("FindDocsForKeyword: indexState.wfdoc == nil")
		return []RetrievalResult{}
	}
	wfdoc := *indexState.wfdoc
	// TODO - separate corpora into simplified and traditional. At the moment
	// only traditional will work
	kw := keyword.Simplified
	if keyword.Traditional != "\\N" {
		kw = keyword.Traditional
	}
	i := 0
	//log.Printf("index.FindForKeyword, wfdoc[kw] %v\n", wfdoc[kw])
	//log.Printf("index.FindForKeyword, len(wfdoc[kw]) %d\n", len(wfdoc[kw]))
	for _, raw := range wfdoc[kw] {
		//log.Printf("index.FindForKeyword, raw.Filename %s\n", raw.Filename)
		if i < maxDocsDisplayed {
			corpusEntry, ok := corpusEntryMap[raw.Filename]
			if !ok {
				log.Printf("index.FindForKeyword, no entry for %s",
					raw.Filename)
				continue
			}
			var hasEngTrans bool
			if bibnotesClient != nil {
				hasEngTrans = len(bibnotesClient.GetTransRefs(corpusEntry.ColFile)) > 0
				log.Printf("index.FindForKeyword, colFile %s for entry file %s, eng %t",
					corpusEntry.ColFile, corpusEntry.RawFile, hasEngTrans)
			}
			item := RetrievalResult{
				HTMLFile: corpusEntry.GlossFile,
				Title: corpusEntry.Title,
				ColTitle: corpusEntry.ColTitle,
				Count: raw.Count,
				HasEngTrans: hasEngTrans,
			}
			docs = append(docs, item)
			i++
		} else {
			break
		}
	}
	//log.Printf("index.FindForKeyword, len(docs) = %d", len(docs))
	return docs
}

// Retrieves raw results for a single keyword
func FindForKeyword(keyword string, wfdoc map[string][]WFDocEntry) []WFDocEntry {
	return wfdoc[keyword]
}
