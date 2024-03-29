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

package index

import (
	"fmt"
	"sort"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

// A word with corpus entry label
type CorpusWord struct {
	Corpus, Word string
}

// A word frequency with corpus entry label
type CorpusWordFreq struct {
	Corpus, Word string
	Freq         int
}

// Sorted list of word frequencies
type SortedWF struct {
	wf map[string]int
	w  []SortedWordItem
}

// An entry in a sorted word array
type SortedWordItem struct {
	Word string
	Freq int
}

// For indexing counts
func (cw CorpusWord) String() string {
	return fmt.Sprintf("%s:%s", cw.Corpus, cw.Word)
}

func (sortedWF *SortedWF) Len() int {
	return len(sortedWF.wf)
}

func (sortedWF *SortedWF) Less(i, j int) bool {
	return sortedWF.wf[sortedWF.w[i].Word] > sortedWF.wf[sortedWF.w[j].Word]
}

func (sortedWF *SortedWF) Swap(i, j int) {
	sortedWF.w[i], sortedWF.w[j] = sortedWF.w[j], sortedWF.w[i]
}

/*
 * Filters a slice of sorted words by domain label if any one of the word
 * senses matches the label.
 */
func FilterByDomain(words []SortedWordItem,
		domain string, wdict map[string]*dicttypes.Word) []dicttypes.Word {
	headwords := []dicttypes.Word{}
	if len(domain) == 0 {
		return headwords
	}
	for _, sw := range words {
		w, ok := wdict[sw.Word]
		if ok {
			wsArr := []dicttypes.WordSense{}
			for _, ws := range w.Senses {
				if ws.Domain == domain {
					wsArr = append(wsArr, ws)
				}
			}
			if len(wsArr) > 0 {
				h := dicttypes.CloneWord(*w)
				h.Senses = wsArr
				headwords = append(headwords, h)
			}
		}
	}
	return headwords
}

/*
 * Sorts Word struct's based on frequency
 */
func SortedFreq(wf map[string]int) []SortedWordItem {
	sortedWF := new(SortedWF)
	sortedWF.wf = wf
	sortedWF.w = make([]SortedWordItem, len(wf))
	i := 0
	for key, _ := range wf {
		sortedWF.w[i] = SortedWordItem{key, sortedWF.wf[key]}
		i++
	}
	sort.Sort(sortedWF)
	return sortedWF.w
}
