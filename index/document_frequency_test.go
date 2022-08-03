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
	"testing"
)

// Trivial test for document frequency
func TestAddVocabulary(t *testing.T) {
	t.Logf("index.DocumentFrequency.TestAddVocabulary enter")
	df := NewDocumentFrequency()
	vocab := map[string]int{}
	df.AddVocabulary(vocab)
	lenExpected := 0
	lenReturned := len(df.DocFreq)
	if lenReturned != lenExpected {
		t.Error("index.TestAddVocabulary0: lenExpected ", lenExpected, " got ",
			lenReturned)
	}
}

// Trivial test for document frequency
func TestIDF0(t *testing.T) {
	df := NewDocumentFrequency()
	vocab := map[string]int{}
	df.AddVocabulary(vocab)
	_, ok := df.IDF("")
	okExpected := false
	if ok != okExpected {
		t.Error("index.TestIDF0: okExpected ", okExpected, " got ", ok)
	}
	n := df.N
	nExpected := 1
	if n != nExpected {
		t.Error("index.TestIDF0: nExpected ", nExpected, " got ", n)
	}
}

// Simple test for document frequency
func TestIDF1(t *testing.T) {
	df := NewDocumentFrequency()
	term := "car"
	vocab := map[string]int{
		term: 1,
	}
	df.AddVocabulary(vocab)
	_, ok := df.IDF(term)
	okExpected := true
	if ok != okExpected {
		t.Error("index.TestIDF1: okExpected ", okExpected, " got ",
			ok)
	}
}

// Slightly harder test
func TestIDF2(t *testing.T) {
	df := NewDocumentFrequency()
	terms := []string{"car", "auto", "insurance", "best"}
	vocab1 := map[string]int{
		terms[0]: 27,
		terms[1]: 3,
		terms[3]: 14,
	}
	df.AddVocabulary(vocab1)
	vocab2 := map[string]int{
		terms[0]: 4,
		terms[1]: 33,
		terms[2]: 33,
	}
	df.AddVocabulary(vocab2)
	vocab3 := map[string]int{
		terms[0]: 24,
		terms[2]: 29,
		terms[3]: 17,
	}
	df.AddVocabulary(vocab3)
	_, ok0 := df.IDF(terms[0])
	okExpected0 := true
	if ok0 != okExpected0 {
		t.Error("index.TestIDF2: okExpected0 ", okExpected0, " got ",
			ok0)
	}
}

// Example from p. 109 of Manning, Christopher D., Prabhakar Raghavan, and
// Hinrich Schütze. Introduction to information retrieval, Cambridge: Cambridge
// university press, 2008.
func TestIDF3(t *testing.T) {
	df := NewDocumentFrequency()
	terms := []string{"car", "auto", "insurance", "best"}
	df.DocFreq[terms[0]] = 18165
	df.DocFreq[terms[1]] = 6723
	df.DocFreq[terms[2]] = 19241
	df.DocFreq[terms[3]] = 25235
	df.N = 806791
	v0, ok := df.IDF(terms[0])
	okExpected := true
	if ok != okExpected {
		t.Error("index.TestIDF0: okExpected ", okExpected, " got ", ok)
	}
	v1, _ := df.IDF(terms[1])
	v2, _ := df.IDF(terms[2])
	v3, _ := df.IDF(terms[3])
	t.Logf("index.TestIDF3 idf = (%v, %v, %v, %v)\n", v0, v1, v2, v3)
}

// Trivial test for tf-idf
func TestTfIdf(t *testing.T) {
	df := NewDocumentFrequency()
	term := "car"
	vocab := map[string]int{
		term: 1,
	}
	df.AddVocabulary(vocab)
	_, ok := tfIdf("car", 1, df)
	if !ok {
		t.Error("index.TestTfIdf: not ok")
	}
}

// TestWriteDocumentFrequency is a trivial test for saving document frequency data
func TestWriteDocumentFrequency(t *testing.T) {
	dfEmpty := NewDocumentFrequency()
	dfSmall := NewDocumentFrequency()
	term := "car"
	vocab := map[string]int{
		term: 1,
	}
	dfSmall.AddVocabulary(vocab)
	type test struct {
		name string
		df   DocumentFrequency
	}
	tests := []test{
		{
			name: "Empty index",
			df:   dfEmpty,
		},
		{
			name: "Small index",
			df:   dfSmall,
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		tc.df.Write(&buf)
		_, err := ReadDocumentFrequency(&buf)
		if err != nil {
			t.Errorf("%s: unexpected error %v", tc.name, err)
		}
	}
}
