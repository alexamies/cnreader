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

// Max collocation elements for a single word
const MAX_COLLOCATIONS = 10 // Max to report
const MAX_STORE = 100 // Max to store

// The key is the headword id, each entry is a bigram frequency map
type CollocationMap map[int]BigramFreqMap

// Put the bigram in the bigram frequency map for the specific word
func (cmPtr *CollocationMap) MergeCollocationMap(more CollocationMap) {
	for k, bfm := range more {
		for _, bigramFreq := range bfm {
    		cmPtr.PutBigramFreq(k, bigramFreq)
    	}
	}
}

// Put the bigram in the bigram frequency map for the specific word
func (cmPtr *CollocationMap) PutBigram(headwordId int, bigram *Bigram) {
	if !bigram.ContainsFunctionWord() {
		cm := *cmPtr
		if bfm, ok := cm[headwordId]; ok {
			bfm.PutBigram(bigram)
		} else {
			newBFM := BigramFreqMap{}
			newBFM.PutBigram(bigram)
			cm[headwordId] = newBFM
		}
	}
}

// Add the BigramFreq object to the CollocationMap
func (cmPtr *CollocationMap) PutBigramFreq(key int, bigramFreq BigramFreq) {
	cm := *cmPtr
	if bfm, ok := cm[key]; !ok {
		bgKey := bigramFreq.BigramVal.String()
		cm[key] = BigramFreqMap{bgKey: bigramFreq}
	} else {
		if len(bfm) < MAX_STORE {
			bfm.PutBigramFreq(bigramFreq)
		}
	}
}

// Get the sorted collocations for a given headword, making sure that there
// are at least two of each and with the total number less than MAX_COLLOCATIONS
func (cmPtr *CollocationMap) SortedCollocations(headwordId int) []BigramFreq {
	cm := *cmPtr
	collArray := SortedFreq(cm[headwordId])
	bfArray := []BigramFreq{}
	for i, bf := range collArray {
		if bf.Frequency > 1 && i < MAX_COLLOCATIONS {
			bfArray = append(bfArray, bf)
		}
	}
	return bfArray
}