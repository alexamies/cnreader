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
	"log"
	"sort"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// A keyword in a document
type Keyword struct {
	Term   string
	Weight float64 // from the tf-idf formula
}

type Keywords []Keyword

func (kws Keywords) Len() int {
	return len(kws)
}

func (kws Keywords) Swap(i, j int) {
	kws[i], kws[j] = kws[j], kws[i]
}

func (kws Keywords) Less(i, j int) bool {
	return kws[i].Weight > kws[j].Weight
}

// Gets the dictionary definition of a slice of strings
// Parameters
//   terms: The Chinese (simplified or traditional) text of the words
// Return
//   hws: an array of word senses
func GetHeadwordArray(keywords Keywords, wdict map[string]dicttypes.Word) ([]dicttypes.Word) {
	hws := []dicttypes.Word{}
	for _, kw := range keywords {
		hw, ok := wdict[kw.Term]
		if ok {
			hws = append(hws, hw)
		} else {
			log.Printf("index.GetHeadwordArray %s not found\n", kw.Term)
		}
	}
	return hws
}

// Orders the keyword with given frequency in a document by tf-idf weight
// Param:
//   vocab - word frequencies for a particular document
func SortByWeight(vocab map[string]int, completeDF DocumentFrequency) Keywords {
	kws := []Keyword{}
	for t, count := range vocab {
		weight, ok := tfIdf(t, count, completeDF)
		if ok {
			kws = append(kws, Keyword{Term: t, Weight: weight})
		} 
	}
	keywords := Keywords(kws)
	sort.Sort(keywords)
	return keywords
}