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

func TestGetHeadwordArray(t *testing.T) {
	wdict := map[string]dicttypes.Word{}
	simp := "多"
	trad := "\\N"
	pinyin := "duō"
	wsArray := []dicttypes.WordSense{}
	hw := dicttypes.Word{
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
	wdict["多"] = hw
	duo := Keyword{"多", 1.1}
	tests := []struct {
		name string
		keywords Keywords
		nExpected int
		pinyin string
	}{
		{
			name: "empty",
			keywords: Keywords{},
			nExpected: 0,
			pinyin: "",
		},
		{
			name: "single term",
			keywords: Keywords{duo},
			nExpected: 1,
			pinyin: "duō",
		},
	}
	for _, tc := range tests {
		hws := GetHeadwordArray(tc.keywords, wdict)
		nReturned := len(hws)
		if tc.nExpected != nReturned {
			t.Errorf("%s: nExpected %d, got %d", tc.name, tc.nExpected, nReturned)
			continue
		}
		if tc.nExpected == 0 {
			continue
		}
		gotPinyin := hws[0].Pinyin
		if tc.pinyin != gotPinyin {
			t.Errorf("%s, expected tc.pinyin %s, got %s", tc.name, tc.pinyin, gotPinyin)
		}
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
