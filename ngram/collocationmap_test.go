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
func TestCMPutBigram(t *testing.T) {
	t.Log("TestCMPutBigram: Begin unit test")

	// Data to test
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
	cm := CollocationMap{}

	// Method being tested
	cm.PutBigram(hw1.HeadwordId, b1)
	cm.PutBigram(hw1.HeadwordId, b1)

	// check result
	bfm := cm[hw1.HeadwordId]

	r1 := bfm.GetBigram(b1)
	e1 := 2
	if r1.Frequency != e1 {
		t.Error("TestCMPutBigram, expected ", e1, " got, ", r1.Frequency)
	}
}

// Test basic Bigram functions
func TestCMPutBigramFreq(t *testing.T) {

	// Data to test
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
	cm := CollocationMap{}
	cm.PutBigram(hw1.HeadwordId, b1)
	bf := BigramFreq{*b1, 3}

	// Method being tested
	cm.PutBigramFreq(hw1.HeadwordId, bf)

	// check result
	bfm := cm[hw1.HeadwordId]

	r1 := bfm.GetBigram(b1)
	e1 := 4
	if r1.Frequency != e1 {
		t.Error("TestCMPutBigramFreq, expected ", e1, " got, ", r1.Frequency)
	}
}


// Test basic Bigram functions
func TestMergeCollocationMap(t *testing.T) {

	// Data to test
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
	cm := CollocationMap{}
	cm.PutBigram(hw1.HeadwordId, b1)
	cm2 := CollocationMap{}
	cm2.PutBigram(hw1.HeadwordId, b1)

	// Method being tested
	cm.MergeCollocationMap(cm2)

	// check result
	bfm := cm[hw1.HeadwordId]

	r1 := bfm.GetBigram(b1)
	e1 := 2
	if r1.Frequency != e1 {
		t.Error("TestCMPutBigramFreq, expected ", e1, " got, ", r1.Frequency)
	}
}