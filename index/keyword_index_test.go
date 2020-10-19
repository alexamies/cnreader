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
	"bytes"
	"io"
	"testing"

	"github.com/alexamies/cnreader/ngram"
)

// Trivial test for corpus-wide word frequency reading
func TestBuildIndex(t *testing.T) {
	var wfFile bytes.Buffer
	var wfDocReader bytes.Buffer
	var indexWriter bytes.Buffer
	indexStore := IndexStore{&wfFile, &wfDocReader, &indexWriter}
	indexState, err := BuildIndex(mockIndexConfig() ,indexStore)
	if err != nil {
		t.Fatalf("unexpected error building index: %v", err)
	}
	if !indexState.KeywordIndexReady {
		t.Errorf("KeywordIndexReady is false")
	}
}

// Trivial test for corpus-wide word frequency reading
func TestReadWFCorpus(t *testing.T) {
	tests := []struct {
		name string
		input string
		expectedLen int
	}{
		{
			name: "empty",
			input: "",
			expectedLen: 0,
		},
		{
			name: "single character",
			input: "鐵\t1",
			expectedLen: 1,
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		io.WriteString(&buf, tc.input)
		wf, err := readWFCorpus(&buf)
		if err != nil {
			t.Fatalf("%s, unexpected error reading word freq: %v", tc.name, err)
		}
		if tc.expectedLen != len(*wf) {
			t.Errorf("%s, expectedLen %d, got %d", tc.name, tc.expectedLen, len(*wf))
		}
	}
}

// TestWriteWFCorpus does a trivial test for corpus index writing
func TestWriteWFCorpus(t *testing.T) {
	tests := []struct {
		name string
		input string
		expectedLen int
	}{
		{
			name: "empty",
			input: "",
			expectedLen: 0,
		},
		{
			name: "single character",
			input: "鐵",
			expectedLen: 1,
		},
	}
	for _, tc := range tests {
		var wFWriter bytes.Buffer
		var unknownCharsWriter bytes.Buffer
		var bigramWriter bytes.Buffer
		wordFreqStore := WordFreqStore{&wFWriter, &unknownCharsWriter, &bigramWriter}
		sw := SortedWordItem{tc.input, 1}
		sortedWords := []SortedWordItem{sw}
		unknownChars := []SortedWordItem{}
		bFreq := []ngram.BigramFreq{}
		err := WriteWFCorpus(wordFreqStore, sortedWords, unknownChars, bFreq, 1, IndexConfig{})
		if err != nil {
			t.Fatalf("%s, unexpected error writing word freq: %v", tc.name, err)
		}
	}
}
