// Test sorting of word frequencies
package analysis

import (
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/ngram"
)

// Test sorting of word frequencies
func TestAddResults(t *testing.T) {
	t.Log("TestAddResults: Begin unit tests")

	// Setup
	vocab := map[string]int{"one": 1, "three": 3, "two": 2}
	usage := map[string]string{"one": "one banana"}
	unknown := map[string]int{"x": 1}
	s1 := "蓝"
	s2 := "藍"
	ws1 := dicttypes.WordSense{
		Id:          1,
		Simplified:  s1,
		Traditional: s2,
		Pinyin:      "lán",
		Grammar:     "adjective",
	}
	hw1 := dicttypes.Word{
		HeadwordId:          1,
		Simplified:  s1,
		Traditional: s2,
		Pinyin:      "",
		Senses:  []dicttypes.WordSense{ws1},
	}
	s3 := "天"
	s4 := "\\N"
	ws2 := dicttypes.WordSense{
		Id:          1,
		Simplified:  s3,
		Traditional: s4,
		Pinyin:      "tiān",
		Grammar:     "noun",
	}
	hw2 := dicttypes.Word{
		HeadwordId:          2,
		Simplified:  s3,
		Traditional: s4,
		Pinyin:      "",
		Senses:  []dicttypes.WordSense{ws2},
	}
	example := ""
	exFile := ""
	exDocTitle := ""
	exColTitle := ""
	b1 := ngram.NewBigram(hw1, hw2, example, exFile, exDocTitle, exColTitle)
	bm := ngram.BigramFreqMap{}
	bm.PutBigram(b1)
	bm.PutBigram(b1)
	results := CollectionAResults{
		Vocab:             vocab,
		Usage:             usage,
		BigramFrequencies: bm,
		WC:                3,
		UnknownChars:      unknown,
	}
	moreVocab := map[string]int{"one": 1, "three": 1, "four": 4}
	moreUsage := map[string]string{"two": "two banana"}
	s5 := "海"
	s6 := "\\N"
	ws3 := dicttypes.WordSense{
		Id:          3,
		Simplified:  s5,
		Traditional: s6,
		Pinyin:      "hǎi",
		Grammar:     "noun",
	}
	hw3 := dicttypes.Word{
		HeadwordId:          3,
		Simplified:  s5,
		Traditional: s6,
		Pinyin:      "",
		Senses:  []dicttypes.WordSense{ws3},
	}
	b2 := ngram.NewBigram(hw1, hw3, example, exFile, exDocTitle, exColTitle)
	bm2 := ngram.BigramFreqMap{}
	bm2.PutBigram(b1)
	bm2.PutBigram(b1)
	bm2.PutBigram(b2)
	unknown1 := map[string]int{"x": 1}
	more := CollectionAResults{
		Vocab:             moreVocab,
		Usage:             moreUsage,
		BigramFrequencies: bm2,
		WC:                4,
		UnknownChars:      unknown1,
	}

	// Method to test
	results.AddResults(&more)

	// Check results
	r := results.Vocab["three"]
	e := 4
	if r != e {
		t.Error("TestAddResults, three expected ", e, " got, ", r)
	}
	r = results.Vocab["four"]
	e = 4
	if r != e {
		t.Error("TestAddResults, four expected ", e, " got, ", r)
	}
	r = results.WC
	e = 7
	if r != e {
		t.Error("TestAddResults, word count expected ", e, " got, ", r)
	}
	bgResult := results.BigramFrequencies.GetBigram(b1)
	r = bgResult.Frequency
	e = 4
	if r != e {
		t.Error("TestAddResults, bigram count expected ", e, " got, ", r)
	}
	rTrad1 := bgResult.BigramVal.Traditional()
	eTrad1 := "藍天"
	if rTrad1 != eTrad1 {
		t.Error("TestAddResults, bigram traditional text expected ", eTrad1,
			" got, ", rTrad1)
	}
	bgResult2 := results.BigramFrequencies.GetBigram(b2)
	r = bgResult2.Frequency
	e = 1
	if r != e {
		t.Error("TestAddResults, bigram2 count expected ", e, " got, ", r)
	}
	rTrad2 := bgResult2.BigramVal.Traditional()
	eTrad2 := "藍海"
	if rTrad2 != eTrad2 {
		t.Error("TestAddResults, bigram traditional text expected ", eTrad2,
			" got, ", rTrad2)
	}
}

// Test sorting of word frequencies
func TestGetLexicalWordFreq0(t *testing.T) {

	// Test setup
	vocab := map[string]int{"one": 1, "three": 3, "two": 2}
	usage := map[string]string{"one": "one banana"}
	unknown := map[string]int{"x": 1}
	sortedWords := []index.SortedWordItem{}
	bm := ngram.BigramFreqMap{}
	results := CollectionAResults{
		Vocab:             vocab,
		Usage:             usage,
		BigramFrequencies: bm,
		WC:                3,
		UnknownChars:      unknown,
	}

	// Method to test
	wdict := make(map[string]dicttypes.Word)
	lexicalWF := results.GetLexicalWordFreq(sortedWords, wdict)

	// Check results
	r := len(lexicalWF)
	e := 0
	if r != e {
		t.Error("TestGetLexicalWordFreq0, expected ", e, " got, ", r)
	}
}

// Test sorting of word frequencies
func TestGetWordFreq0(t *testing.T) {

	// Test setup
	vocab := map[string]int{"one": 1, "three": 3, "two": 2}
	usage := map[string]string{"one": "one banana"}
	unknown := map[string]int{"x": 1}
	sortedWords := []index.SortedWordItem{}
	bm := ngram.BigramFreqMap{}
	results := CollectionAResults{
		Vocab:             vocab,
		Usage:             usage,
		BigramFrequencies: bm,
		WC:                3,
		UnknownChars:      unknown,
	}

	// Method to test
	wdict := make(map[string]dicttypes.Word)
	lexicalWF := results.GetWordFreq(sortedWords, wdict)

	// Check results
	r := len(lexicalWF)
	e := 0
	if r != e {
		t.Error("TestGetLexicalWordFreq0, expected ", e, " got, ", r)
	}
}
