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
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

func TestGetHeadwordArray0(t *testing.T) {
	keywords := Keywords{}
	wdict := map[string]dicttypes.Word{}
	hws := GetHeadwordArray(keywords, wdict)
	nReturned := len(hws)
	nExpected := 0
	if nReturned != nExpected {
		t.Error("index.TestGetHeadwordArray0: nExpected ", nExpected, " got ",
			nReturned)
	}
}

func TestGetHeadwordArray1(t *testing.T) {
	kw := Keyword{"多", 1.1}
	keywords := Keywords{kw}
	wdict := make(map[string]dicttypes.Word)
	hws := GetHeadwordArray(keywords, wdict)
	nReturned := len(hws)
	nExpected := 1
	if nReturned != nExpected {
		t.Error("index.TestGetHeadwordArray1: nExpected ", nExpected, " got ",
			nReturned)
	}
	nPinyin := len(hws[0].Pinyin)
	nPExpected := 1
	if nPinyin != nPExpected {
		t.Error("index.TestGetHeadwordArray1: nPExpected ", nPExpected, " got ",
			nPinyin)
	}
}

// Trivial test for keyword ordering
func TestSortByWeight0(t *testing.T) {
	vocab := map[string]int{}
	df := NewDocumentFrequency()
	keywords := SortByWeight(vocab, df)
	nReturned := len(keywords)
	nExpected := 0
	if nReturned != nExpected {
		t.Error("index.TestSortByWeight0: nExpected ", nExpected, " got ",
			nReturned)
	}
}

// Simple test for keyword ordering
func TestSortByWeight1(t *testing.T) {
	df := NewDocumentFrequency()
	term0 := "好"
	vocab := map[string]int{
		term0: 1,
	}
	df.AddVocabulary(vocab)
	keywords := SortByWeight(vocab, df)
	nReturned := len(keywords)
	nExpected := 1
	if nReturned != nExpected {
		t.Error("index.TestSortByWeight1: nExpected ", nExpected, " got ",
			nReturned)
	}
	top := keywords[0]
	topExpected := term0
	if top.Term != topExpected {
		t.Error("index.TestSortByWeight1: topExpected ", topExpected, " got ",
			top)
	}
}

// Easy test for keyword ordering
// vocab1: term1 = 1*log10(1/1) = 0.0, term2 = 2*log10(2/1) = 0.60
//
func TestSortByWeight2(t *testing.T) {
	df := NewDocumentFrequency()
	term0 := "你"
	term1 := "好"
	term2 := "嗎"
	vocab0 := map[string]int{
		term0: 1,
		term1: 2,
	}
	df.AddVocabulary(vocab0)
	vocab1 := map[string]int{
		term1: 1,
		term2: 2,
	}
	df.AddVocabulary(vocab1)
	keywords := SortByWeight(vocab1, df)
	top := keywords[0]
	topExpected := term2
	if top.Term != topExpected {
		t.Logf("index.TestSortByWeight2 keywords = %v\n", keywords)
		t.Errorf("index.TestSortByWeight2: topExpected %v, got %v",
			topExpected, top)
	}
}