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
	"github.com/alexamies/cnreader/ngram"
	"testing"
)

// Empty test for corpus index writing
func TestReset(t *testing.T) {
	Reset(mockIndexConfig())
}

// Trivial test for corpus-wide word frequency reading
func TestBuildIndex0(t *testing.T) {
	BuildIndex(mockIndexConfig())
}

// Trivial test for corpus-wide word frequency reading
func TestReadWFCorpus0(t *testing.T) {
	readWFCorpus(mockIndexConfig())
}

// Trivial test for corpus-wide word frequency reading
func TestReadWFCorpus1(t *testing.T) {
	w := "鐵"
	sw := SortedWordItem{w, 1}
	sortedWords := []SortedWordItem{sw}
	unknownChars := []SortedWordItem{}
	bFreq := []ngram.BigramFreq{}
	indexConfig := mockIndexConfig()
	WriteWFCorpus(sortedWords, unknownChars, bFreq, 1, indexConfig)
	readWFCorpus(indexConfig)
	entry := wf[w]
	expected := 1
	if entry.Count != expected {
		t.Error("index.TestReadWFCorpus1: Expected ", expected, " got ",
			entry.Count)
	}
}

// Trivial test for corpus index writing
func TestWriteWFCorpus0(t *testing.T) {
	sortedWords := []SortedWordItem{}
	unknownChars := []SortedWordItem{}
	bFreq := []ngram.BigramFreq{}
	WriteWFCorpus(sortedWords, unknownChars, bFreq, 0, mockIndexConfig())
}

// Simple test for corpus index writing
func TestWriteWFCorpus1(t *testing.T) {
	sw := SortedWordItem{"鐵", 1}
	sortedWords := []SortedWordItem{sw}
	uw := SortedWordItem{"𣣌", 2}
	unknownChars := []SortedWordItem{uw}
	bFreq := []ngram.BigramFreq{}
	WriteWFCorpus(sortedWords, unknownChars, bFreq, 1, mockIndexConfig())
}
