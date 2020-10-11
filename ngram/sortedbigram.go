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

package ngram

import (
	"sort"
	)

// Sorted into descending order with most frequent bigram first
type SortedBFM struct {
	bfm BigramFreqMap
	f []BigramFreq
}

func (sbf *SortedBFM) Len() int {
	return len(sbf.bfm)
}

func (sbf *SortedBFM) Less(i, j int) bool {
	return sbf.bfm.GetBigram(&sbf.f[i].BigramVal).Frequency > sbf.bfm.GetBigram(&sbf.f[j].BigramVal).Frequency
}

func NewSortedBFM(bfm BigramFreqMap) *SortedBFM {
	return &SortedBFM{bfm, []BigramFreq{}}
}

func (sbf *SortedBFM) Swap(i, j int) {
	sbf.f[i], sbf.f[j] = sbf.f[j], sbf.f[i]
}

// Get the bigram frequencies as a sorted array
func SortedFreq(bfm BigramFreqMap) []BigramFreq {
	sortedBFM := new(SortedBFM)
	sortedBFM.bfm = bfm
	sortedBFM.f = make([]BigramFreq, len(bfm))
	i := 0
	for key, _ := range bfm {
		sortedBFM.f[i] = sortedBFM.bfm[key]
		i++
	}
	sort.Sort(sortedBFM)
	return sortedBFM.f
}