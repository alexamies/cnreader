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

package tmindex

import (
	"bytes"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"testing"
)

func mockDictionary(t *testing.T) map[string]*dicttypes.Word {
	s1 := "结实"
	t1 := "結實"
	p1 := "jiēshi"
	ws1 := dicttypes.WordSense{
		HeadwordId: 1,
		Simplified: s1,
		Traditional: t1,
		Pinyin: p1,
		English: "solid",
		Domain: "Modern Chinese",
	}
	hw1 := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s1, 
		Traditional: t1,
		Pinyin: p1,
		Senses: []dicttypes.WordSense{ws1},
	}
	s2 := "倿"
	t2 := "\\N"
	p2 := "nìng"
	ws2 := dicttypes.WordSense{
		HeadwordId: 2,
		Simplified: s2,
		Traditional: t2,
		Pinyin: p2,
		English: "",
		Domain: "Literary Chinese",
	}
	hw2 := dicttypes.Word{
		HeadwordId: 2,
		Simplified: s2, 
		Traditional: t2,
		Pinyin: p2,
		Senses: []dicttypes.WordSense{ws2},
	}
  headwords := []*dicttypes.Word{&hw1, &hw2}
  wdict := make(map[string]*dicttypes.Word)
  for _, hw := range headwords {
  	t.Logf("mockDictionary adding simplified key %s: %v\n", hw.Simplified, hw)
  	wdict[hw.Simplified] = hw
  	trad := hw.Traditional
  	if trad != "\\N" {
  		t.Logf("mockDictionary adding trad key %s: %v\n", trad, hw)
  		wdict[trad] = hw
  	}
  }
  return wdict
}

// Test basic BuildIndex functions
func TestBuildUniDomainIndex(t *testing.T) {
	wdict := mockDictionary(t)
	for k, w := range wdict {
		for _, s := range w.Senses {
			t.Logf("TestBuildUniDomainIndex: %s %d %s: %s %s\n", k, w.HeadwordId, s.Pinyin, s.English, s.Domain)
		}
	}
	var buf bytes.Buffer
	err := buildUniDomainIndex(&buf, wdict)
	expected := 
`结	结实	Modern Chinese
实	结实	Modern Chinese
結	結實	Modern Chinese
實	結實	Modern Chinese
倿	倿	Literary Chinese
`
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := buf.String()
	if len(result) != len(expected) {
		t.Errorf("got: %d, expected: %d, got:\n%s, want:\n%s", len(result),
				len(expected), result, expected)
	}
}

// Test basic BuildIndex functions
func TestBuildUnigramIndex(t *testing.T) {
	wdict := mockDictionary(t)
	var buf bytes.Buffer
	err := buildUnigramIndex(&buf, wdict)
	expected := 
`结	结实
实	结实
結	結實
實	結實
倿	倿
`
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := buf.String()
	if len(result) != len(expected) {
		t.Errorf("got: %d, expected: %d, index:\n%s", len(result), len(expected), result)
	}
}
