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

package analysis

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/dictionary"	
	"github.com/alexamies/chinesenotes-go/dicttypes"	
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/generator"
	"github.com/alexamies/cnreader/index"
)

func mockCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir: "corpus",
		Excluded: map[string]bool{},
		ProjectHome: ".",
	}
}

// IndexConfig encapsulates parameters for index configuration
func mockIndexConfig() index.IndexConfig {
	return index.IndexConfig{
		IndexDir: "index",
	}
}

func mockOutputConfig() generator.HTMLOutPutConfig {
	return generator.HTMLOutPutConfig{
		ContainsByDomain: "",
		Domain: "",
		GoStaticDir: "static",
		TemplateDir: "templates",
		VocabFormat: "",
		WebDir: "web-staging",
	}
}

func mockDictionaryConfig() dicttypes.DictionaryConfig {
	return dicttypes.DictionaryConfig{
		AvoidSubDomains: map[string]bool{},
		DictionaryDir: "data",
	}
}

func mockFileCorpusLoader() corpus.FileCorpusLoader {
	return corpus.FileCorpusLoader{
		FileName: "File",
		Config: mockCorpusConfig(),
	}
}

func mockSmallDict() map[string]dicttypes.Word {
	s1 := "繁体中文"
	t1 := "繁體中文"
	hw1 := dicttypes.Word{
		HeadwordId:  	1,
		Simplified:  	s1,
		Traditional: 	t1,
		Pinyin:      	"fántǐ zhōngwén",
		Senses:  			[]dicttypes.WordSense{},
	}
	s2 := "前"
	t2 := "\\N"
	hw2 := dicttypes.Word{
		HeadwordId:  	2,
		Simplified:  	s2,
		Traditional: 	t2,
		Pinyin:      	"qián",
		Senses:  			[]dicttypes.WordSense{},
	}
	s3 := "不见"
	t3 := "不見"
	hw3 := dicttypes.Word{
		HeadwordId:  	3,
		Simplified:  	s3,
		Traditional: 	t3,
		Pinyin:      	"bújiàn",
		Senses:				[]dicttypes.WordSense{},
	}
	s4 := "古人"
	t4 := "\\N"
	hw4 := dicttypes.Word{
		HeadwordId:  	4,
		Simplified:  	s4,
		Traditional: 	t4,
		Pinyin:      	"gǔrén",
		Senses:  			[]dicttypes.WordSense{},
	}
	s5 := "夫"
	t5 := "\\N"
	hw5 := dicttypes.Word{
		HeadwordId:  	5,
		Simplified:  	s5,
		Traditional: 	t5,
		Pinyin:      	"fú fū",
		Senses:  			[]dicttypes.WordSense{},
	}
	s6 := "起信论"
	t6 := "起信論"
	hw6 := dicttypes.Word{
		HeadwordId:  	6,
		Simplified:  	s6,
		Traditional: 	t6,
		Pinyin:      	"Qǐ Xìn Lùn",
		Senses:  			[]dicttypes.WordSense{},
	}
	s7 := "者"
	t7 := "\\N"
	hw7 := dicttypes.Word{
		HeadwordId:  	7,
		Simplified:  	s7,
		Traditional: 	t7,
		Pinyin:      	"zhě zhuó",
		Senses:  			[]dicttypes.WordSense{},
	}
	s8 := "乃是"
	t8 := "\\N"
	hw8 := dicttypes.Word{
		HeadwordId:  	8,
		Simplified:  	s8,
		Traditional: 	t8,
		Pinyin:      	"nǎishì",
		Senses:  			[]dicttypes.WordSense{},
	}
	return map[string]dicttypes.Word {
		s1: hw1,
		t1: hw1,
		s2: hw2,
		s3: hw3,
		t3: hw3,
		s4: hw4,
		s5: hw5,
		s6: hw6,
		t6: hw6,
		s7: hw7,
		s8: hw8,
	}
}

func mockValidator() (dictionary.Validator, error) {
	const posList = "noun\nverb\n"
	posReader := strings.NewReader(posList)
	const domainList = "艺术	Art	\\N	\\N\n佛教	Buddhism	\\N	\\N\n"
	domainReader := strings.NewReader(domainList)
	return dictionary.NewValidator(posReader, domainReader)
}

func TestGetChunks1(t *testing.T) {
	chunks := GetChunks("中文")
	if chunks.Len() != 1 {
		t.Error("TestGetChunks1: Expected length of chunks 1, got ",
			chunks.Len())
	}
	chunk := chunks.Front().Value.(string)
	if chunk != "中文" {
		t.Error("TestGetChunks1: Expected first element of chunk 中文, got ",
			chunk)
	}
}

func TestGetChunks2(t *testing.T) {
	chunks := GetChunks("a中文")
	if chunks.Len() != 2 {
		t.Error("Expected length of chunks 2, got ", chunks.Len())
	}
	chunk := chunks.Front().Value.(string)
	if chunk != "a" {
		t.Error("Expected first element of chunk a, got ", chunk)
	}
}

func TestGetChunks3(t *testing.T) {
	chunks := GetChunks("a中文b")
	if chunks.Len() != 3 {
		t.Error("Expected length of chunks 3, got ", chunks.Len())
	}
	chunk := chunks.Front().Value.(string)
	if chunk != "a" {
		t.Error("Expected first element of chunk a, got ", chunk)
	}
}

// Simplified Chinese
func TestGetChunks4(t *testing.T) {
	chunks := GetChunks("简体中文")
	if chunks.Len() != 1 {
		t.Error("Simplified Chinese 简体中文: expected length of chunks 1, got ",
			chunks.Len())
	}
	chunk := chunks.Front().Value.(string)
	if chunk != "简体中文" {
		for e := chunks.Front(); e != nil; e = e.Next() {
			t.Logf("TestGetChunks4: chunk: %s", e.Value.(string))
		}
		t.Error("Expected first element of chunk 简体中文 to be 简体中文, got ",
			chunk)
	}
}

func TestReadText1(t *testing.T) {
	//log.Printf("TestReadText1: Begin ******** \n")
	corpusLoader := mockFileCorpusLoader()
	var buf bytes.Buffer
	text := corpusLoader.ReadText(&buf)
	expected := "繁體中文"
	//log.Printf("TestReadText1: Expected  '%s', got '%s'\n", expected, text)
	if text != expected {
		t.Error("Expected ", expected, ", got ", text)
	}
	//log.Printf("TestReadText1: End ******** \n")
}

func TestParseText(t *testing.T) {
	t.Log("TestParseText: Begin ********")
	testCases := []struct {
		name string
		in  string
		expectedFirst string
		expectedTokens int
		expectedVocab int
		expectedWC int
		expectedCC int
	}{
		{
			name: "One token",
			in: "繁體中文", 
			expectedFirst: "繁體中文",
			expectedTokens: 1,
			expectedVocab: 1,
			expectedWC: 1,
			expectedCC: 4,
		},
		{
			name: "ASCII and one token",
			in: "a繁體中文", 
			expectedFirst: "a",
			expectedTokens: 2,
			expectedVocab: 1,
			expectedWC: 1,
			expectedCC: 4,
		},
		{
			name: "Three tokens",
			in: "前不见古人", 
			expectedFirst: "前",
			expectedTokens: 3,
			expectedVocab: 3,
			expectedWC: 3,
			expectedCC: 5,
		},
		{
			name: "More tokens",
			in: "夫起信論者，乃是...。", 
			expectedFirst: "夫",
			expectedTokens: 6,
			expectedVocab: 4,
			expectedWC: 4,
			expectedCC: 7,
		},
	}
	wdict := mockSmallDict()
	tok := tokenizer.DictTokenizer{wdict}
	for _, tc := range testCases {
		tokens, results := ParseText(tc.in, "", corpus.NewCorpusEntry(),
			tok, mockCorpusConfig(), wdict)
		if tokens.Len() != tc.expectedTokens {
			t.Fatalf("%s: expectedTokens %d, got %d", tc.name, tc.expectedTokens,
					tokens.Len())
		}
		first := tokens.Front().Value.(string)
		if tc.expectedFirst != first {
			t.Errorf("%s: expectedFirst: %s, got %s", tc.name, tc.expectedFirst, first)
		}
		if tc.expectedVocab != len(results.Vocab) {
			t.Errorf("%s: expectedVocab = %d, got %d", tc.name, tc.expectedVocab,
					len(results.Vocab))
		}
		if tc.expectedWC != results.WC {
			t.Errorf("%s: expectedWC: %d, got %d", tc.name, tc.expectedWC, results.WC)
		}
		if tc.expectedCC != results.CCount {
			t.Errorf("%s: expectedCC: %d, got %d", tc.name, tc.expectedCC, results.CCount)
		}
	}
	t.Log("TestParseText: End ******** ")
}

// Basic test with no data
func TestSampleUsage1(t *testing.T) {
	usageMap := map[string]*[]WordUsage{}
	usageMap = sampleUsage(usageMap)
	l := len(usageMap)
	expected := 0
	if l != expected {
		t.Error("Expected to get length ", expected, ", got ", l)
	}
}

// Basic test with minimal data
func TestSampleUsage2(t *testing.T) {
	wu := WordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "A Big Snake",
	}
	wuArray := []WordUsage{wu}
	usageMap := map[string]*[]WordUsage{"大": &wuArray}
	usageMap = sampleUsage(usageMap)
	l := len(usageMap)
	expected := 1
	if l != expected {
		t.Error("Expected to get length ", expected, ", got ", l)
	}
}

// Basic test with more data
func TestSampleUsage3(t *testing.T) {
	t.Log("TestSampleUsage3: Begin +++++++++++")
	wu1 := WordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "蛇",
		Example:    "大蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "Some Snakes",
	}
	wu2 := WordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "蛇",
		Example:    "小蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 2",
		ColTitle:   "Some Snakes",
	}
	wuArray := []WordUsage{wu1, wu2}
	usageMap := map[string]*[]WordUsage{"蛇": &wuArray}
	usageMap = sampleUsage(usageMap)
	l := len(*usageMap["蛇"])
	expected := 2
	if l != expected {
		t.Error("Expected to get length ", expected, ", got ", l)
	}
}

// Basic test with more data
func TestSampleUsage4(t *testing.T) {
	t.Logf("analysis.TestSampleUsage4: Begin +++++++++++")
	wu1 := WordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "Some Big Animals",
	}
	wu2 := WordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大老虎",
		File:       "afile.txt",
		EntryTitle: "Scroll 2",
		ColTitle:   "Some Big Animals",
	}
	wu3 := WordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大树",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "Some Big Trees",
	}
	wuArray := []WordUsage{wu1, wu2, wu3}
	usageMap := map[string]*[]WordUsage{"大": &wuArray}
	usageMap = sampleUsage(usageMap)
	l := len(*usageMap["大"])
	expected := 3
	if l != expected {
		t.Error("Expected to get length ", expected, ", got ", l)
	}
	t.Log("analysis.TestSampleUsage4: End +++++++++++")
}

func TestSortedFreq(t *testing.T) {
	wdict := make(map[string]dicttypes.Word)
	text := "夫起信論者，乃是至極大乘甚深祕典"
	_, results := ParseText(text, "", corpus.NewCorpusEntry(),
			tokenizer.DictTokenizer{}, mockCorpusConfig(), wdict)
	sortedWords := index.SortedFreq(results.Vocab)
	expected := len(results.Vocab)
	got := len(sortedWords)
	if expected != got {
		t.Fatalf("TestSortedFreq: Expected %d, got %d", expected, got)
	}
}

func TestWriteAnalysis(t *testing.T) {
	t.Log("analysis.TestWriteAnalysis: Begin +++++++++++")
	term := "繁"
	wdict := make(map[string]dicttypes.Word)
	_, results := ParseText(term, "", corpus.NewCorpusEntry(),
			tokenizer.DictTokenizer{}, mockCorpusConfig(), wdict)
	srcFile := "test.txt"
	glossFile := "test.html"
	vocab := map[string]int{
		term: 1,
	}
	df := index.NewDocumentFrequency()
	df.AddVocabulary(vocab)
	var buf bytes.Buffer
	df.Write(&buf)
	index.ReadDocumentFrequency(&buf)
	writeAnalysis(results, srcFile, glossFile, "Test Collection", "Test Doc",
			mockOutputConfig(), wdict)
	t.Log("analysis.TestWriteAnalysis: End +++++++++++")
}

func TestWriteDoc(t *testing.T) {
	t.Log("analysis.TestWriteDoc: Begin +++++++++++")
	wdict := make(map[string]dicttypes.Word)
	input1 := `
  	A test document
    繁體中文	
	`
	input2 := `
 	  <p>A test document with simplified Chinese</p>
    <p>简体中文</p>
    <p>Word with multiple senses: 中</p>
	`
	type test struct {
		name string
		input string
		expectError bool
		expectLen int
  }
  tests := []test{
		{
			name: "One character",
			input: "繁",
			expectError: false,
			expectLen: 1,
		},
		{
			name: "Four characters",
			input: input1,
			expectError: false,
			expectLen: 6,
		},
		{
			name: "Simplified characters",
			input: input2,
			expectError: false,
			expectLen: 8,
		},
  }
  for _, tc := range tests {
		tokens, results := ParseText(tc.input, "", corpus.NewCorpusEntry(),
				tokenizer.DictTokenizer{}, mockCorpusConfig(), wdict)
		if tc.expectLen != tokens.Len() {
			t.Errorf("%s, expected len %d, got %d", tc.name, tc.expectLen, tokens.Len())
		}
		var buf bytes.Buffer
		err := WriteDoc(tokens, results.Vocab, &buf, `\N`, `\N`, true, "",
				mockCorpusConfig(), wdict)
		if !tc.expectError && err != nil {
			t.Errorf("%s unexpected error: %v", tc.name, err)
		}
	}
	t.Log("analysis.TestWriteDoc: End +++++++++++")
}