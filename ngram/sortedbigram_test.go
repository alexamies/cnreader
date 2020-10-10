// Test sorting of bigram frequencies
package ngram

import (
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

// Test basic Bigram functions
func TestSortedBFM(t *testing.T) {
	t.Log("TestSortedBFM: Begin unit test")
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
	bm.PutBigram(b1)
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
	bm.PutBigram(b2)
	sbf := SortedFreq(bm)
	r1 := len(sbf)
	e1 := 2
	if r1 != e1 {
		t.Error("TestSortedBFM, expected ", e1, " got, ", r1)
	}
	r2 := sbf[0].Frequency
	e2 := 2
	if r2 != e2 {
		t.Error("TestSortedBFM, expected ", e2, " got, ", r2, "sbf[0]", sbf[0])
	}
	r3 := sbf[0].BigramVal.HeadwordDef1.Simplified
	e3 := "蓝"
	if r3 != e3 {
		t.Error("TestSortedBFM, expected ", e3, " got, ", r3, "sbf[0]", sbf[0])
	}
}
