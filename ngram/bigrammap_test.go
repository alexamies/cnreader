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
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

// Test basic Bigram functions
func TestMerge(t *testing.T) {
	t.Log("TestMerge: Begin unit test")

	// Set up test
	s1 := "蓝"
	s2 := "藍"
	ws1 := dicttypes.WordSense{
		HeadwordId: 1,
		Simplified: s1, 
		Traditional: s2,
		Pinyin: "lán",
		Grammar: "adjective",
	}
	hw1 := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s1, 
		Traditional: s2,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws1},
	}
	s3 := "天"
	s4 := "\\N"
	ws2 := dicttypes.WordSense{
		HeadwordId: 1,
		Simplified: s3, 
		Traditional: s4,
		Pinyin: "tiān",
		Grammar: "noun",
	}
	hw2 := dicttypes.Word{
		HeadwordId: 2,
		Simplified: s3, 
		Traditional: s4,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws2},
	}
	example := ""
	exFile := ""
	exDocTitle := ""
	exColTitle := ""
	b1 := NewBigram(hw1, hw2, example, exFile, exDocTitle, exColTitle)
	bm := BigramFreqMap{}

	// Invoke twice
	bm.PutBigram(b1)
	bm.PutBigram(b1)

	// Second map
	s5 := "海"
	s6 := "\\N"
	ws3 := dicttypes.WordSense{
		Id: 3,
		Simplified: s5, 
		Traditional: s6,
		Pinyin: "hǎi",
		Grammar: "noun",
	}
	hw3 := dicttypes.Word{
		HeadwordId: 3,
		Simplified: s5,
		Traditional: s6,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws3},
	}
	b2 := NewBigram(hw1, hw3, example, exFile, exDocTitle, exColTitle)
	bm2 := BigramFreqMap{}
	bm2.PutBigram(b1)
	bm2.PutBigram(b2)

	// Method to test
	bm.Merge(bm2)

	// Check result
	r1 := bm.GetBigram(b1)
	e1 := 3
	if r1.Frequency != e1 {
		t.Error("TestPutBigram, expected ", e1, " got, ", r1.Frequency)
	}
	r2 := bm.GetBigram(b2)
	e2 := 1
	if r2.Frequency != e2 {
		t.Error("TestPutBigram, expected ", e2, " got, ", r2.Frequency)
	}
}

// Test basic Bigram functions
func TestPutBigram(t *testing.T) {

	// Set up test
	s1 := "蓝"
	s2 := "藍"
	ws1 := dicttypes.WordSense{
		Id: 1,
		Simplified: s1, 
		Traditional: s2,
		Pinyin: "lán",
		Grammar: "adjective",
	}
	hw1 := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s1, 
		Traditional: s2,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws1},
	}
	s3 := "天"
	s4 := "\\N"
	ws2 := dicttypes.WordSense{
		Id: 1,
		Simplified: s3, 
		Traditional: s4,
		Pinyin: "tiān",
		Grammar: "noun",
	}
	hw2 := dicttypes.Word{
		HeadwordId: 2,
		Simplified: s3, 
		Traditional: s4,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws2},
	}
	example := ""
	exFile := ""
	exDocTitle := ""
	exColTitle := ""
	b1 := NewBigram(hw1, hw2, example, exFile, exDocTitle, exColTitle)
	bm := BigramFreqMap{}

	// Method to test (invoke twice)
	bm.PutBigram(b1)
	bm.PutBigram(b1)

	// Check result
	r1 := bm.GetBigram(b1)
	e1 := 2
	if r1.Frequency != e1 {
		t.Error("TestPutBigram, expected ", e1, " got, ", r1.Frequency)
	}
	s5 := "海"
	s6 := "\\N"
	ws3 := dicttypes.WordSense{
		Id: 3,
		Simplified: s5, 
		Traditional: s6,
		Pinyin: "hǎi",
		Grammar: "noun",
	}
	hw3 := dicttypes.Word{
		HeadwordId: 3,
		Simplified: s5,
		Traditional: s6,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws3},
	}
	b2 := NewBigram(hw1, hw3, example, exFile, exDocTitle, exColTitle)
	r2 := bm.GetBigram(b2)
	e2 := 0
	if r2.Frequency != e2 {
		t.Error("TestPutBigram, expected ", e2, " got, ", r2.Frequency)
	}
}

// Test basic Bigram functions
func TestPutBigramFreq(t *testing.T) {

	// Set up test
	s1 := "蓝"
	s2 := "藍"
	ws1 := dicttypes.WordSense{
		Id: 1,
		Simplified: s1, 
		Traditional: s2,
		Pinyin: "lán",
		Grammar: "adjective",
	}
	hw1 := dicttypes.Word{
		HeadwordId: 1,
		Simplified: s1, 
		Traditional: s2,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws1},
	}
	s3 := "天"
	s4 := "\\N"
	ws2 := dicttypes.WordSense{
		Id: 1,
		Simplified: s3, 
		Traditional: s4,
		Pinyin: "tiān",
		Grammar: "noun",
	}
	hw2 := dicttypes.Word{
		HeadwordId: 2,
		Simplified: s3, 
		Traditional: s4,
		Pinyin: "",
		Senses: []dicttypes.WordSense{ws2},
	}
	example := ""
	exFile := ""
	exDocTitle := ""
	exColTitle := ""
	b1 := NewBigram(hw1, hw2, example, exFile, exDocTitle, exColTitle)
	bm := BigramFreqMap{}
	bm.PutBigram(b1)
	bf := BigramFreq{*b1, 2}

	// Method to test (invoke twice)
	bm.PutBigramFreq(bf)

	// Check result
	r1 := bm.GetBigram(b1)
	e1 := 3
	if r1.Frequency != e1 {
		t.Error("TestPutBigramFreq, expected ", e1, " got, ", r1.Frequency)
	}
}