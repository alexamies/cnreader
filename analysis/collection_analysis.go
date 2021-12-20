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

package analysis

import (
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/ngram"
)

// A struct to hold the analysis results for the collection
type CollectionAResults struct {
	Vocab             map[string]int
	Bigrams           map[string]int
	Usage             map[string]string
	BigramFrequencies ngram.BigramFreqMap
	Collocations      ngram.CollocationMap
	WC, CCount        int
	UnknownChars      map[string]int
	WFDocMap          index.TermFreqDocMap
	BigramDocMap      index.TermFreqDocMap
	DocFreq           index.DocumentFrequency
	BigramDF          index.DocumentFrequency
	DocLengthArray    []index.DocLength
}

// Add more results to this set of results
func (results *CollectionAResults) AddResults(more *CollectionAResults) {

	for k, v := range more.Vocab {
		results.Vocab[k] += v
	}

	for k, v := range more.Bigrams {
		results.Bigrams[k] += v
	}

	for k, v := range more.Usage {
		results.Usage[k] = v
	}

	results.BigramFrequencies.Merge(more.BigramFrequencies)

	results.Collocations.MergeCollocationMap(more.Collocations)

	results.WC += more.WC
	results.CCount += more.CCount

	for k, v := range more.UnknownChars {
		results.UnknownChars[k] += v
	}
	results.DocLengthArray = append(results.DocLengthArray, more.DocLengthArray...)
}

// Returns the subset of words that are lexical (content) words
func (results *CollectionAResults) GetLexicalWordFreq(sortedWords []index.SortedWordItem,
	wdict map[string]dicttypes.Word) []wFResult {
	wfResults := make([]wFResult, 0)
	for _, value := range sortedWords {
		if word, ok := wdict[value.Word]; ok {
			for _, ws := range word.Senses {
				if !ws.IsFunctionWord() {
					wfResults = append(wfResults, wFResult{
						Freq:       value.Freq,
						HeadwordId: ws.HeadwordId,
						Chinese:    value.Word,
						Pinyin:     ws.Pinyin,
						English:    ws.English,
						Usage:      results.Usage[value.Word]})
				}
			}
		}
	}
	return wfResults
}

// Returns the subset of words that are lexical (content) words
func (results *CollectionAResults) GetHeadwords(wdict map[string]dicttypes.Word) []dicttypes.Word {
	headwords := make([]dicttypes.Word, 0, len(results.Vocab))
	for k := range results.Vocab {
		if hw, ok := wdict[k]; ok {
			headwords = append(headwords, hw)
		}
	}
	return headwords
}

// Returns the subset of words that are lexical (content) words
func (results *CollectionAResults) GetWordFreq(sortedWords []index.SortedWordItem,
	wdict map[string]dicttypes.Word) []wFResult {

	wfResults := make([]wFResult, 0)
	maxWFOutput := len(sortedWords)
	if maxWFOutput > 100 {
		maxWFOutput = 100
	}
	for _, value := range sortedWords[:maxWFOutput] {
		if word, ok := wdict[value.Word]; ok {
			for _, ws := range word.Senses {
				wfResults = append(wfResults, wFResult{
					Freq:       value.Freq,
					HeadwordId: ws.HeadwordId,
					Chinese:    value.Word,
					Pinyin:     ws.Pinyin,
					English:    ws.English,
					Usage:      results.Usage[value.Word]})
			}
		}
	}
	return wfResults
}

// Constructor for empty CollectionAResults
func NewCollectionAResults() CollectionAResults {
	return CollectionAResults{
		Vocab:             map[string]int{},
		Bigrams:           map[string]int{},
		Usage:             map[string]string{},
		BigramFrequencies: ngram.BigramFreqMap{},
		Collocations:      ngram.CollocationMap{},
		WC:                0,
		UnknownChars:      map[string]int{},
		WFDocMap:          index.TermFreqDocMap{},
		BigramDocMap:      index.TermFreqDocMap{},
		DocFreq:           index.NewDocumentFrequency(),
		BigramDF:          index.NewDocumentFrequency(),
		DocLengthArray:    []index.DocLength{},
	}
}
