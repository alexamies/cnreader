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
	"io"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"	
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
		Templates: generator.NewTemplateMap(config.AppConfig{}),
	}
}

func mockDictionaryConfig() dicttypes.DictionaryConfig {
	return dicttypes.DictionaryConfig{
		AvoidSubDomains: map[string]bool{},
		DictionaryDir: "data",
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

func TestGetChunks(t *testing.T) {
	testCases := []struct {
		name string
		in  string
		expectedLen int
		expectedOut string
	}{
		{
			name: "One chunk",
			in: "中文", 
			expectedLen: 1,
			expectedOut: "中文",
		},
		{
			name: "Two chunks",
			in: "a中文", 
			expectedLen: 2,
			expectedOut: "a",
		},
		{
			name: "Three chunks",
			in: "a中文b", 
			expectedLen: 3,
			expectedOut: "a",
		},
		{
			name: "Simplified Chinese",
			in: "简体中文", 
			expectedLen: 1,
			expectedOut: "简体中文",
		},
	}
	for _, tc := range testCases {
		chunks := getChunks(tc.in)
		if chunks.Len() != tc.expectedLen {
			t.Errorf("TestGetChunks %s: Expected length of chunks got %d, want %d",
				tc.name, chunks.Len(), tc.expectedLen)
		}
		chunk := chunks.Front().Value.(string)
		if chunk != tc.expectedOut {
			t.Errorf("TestGetChunks %s: Expected first element of chunk %s, want %s",
				tc.name, chunk, tc.expectedOut)
		}
	}
}


func TestGetHeadwords(t *testing.T) {
	s1 := "繁体中文"
	t1 := "繁體中文"
	hw1 := dicttypes.Word{
		HeadwordId:  	1,
		Simplified:  	s1,
		Traditional: 	t1,
		Pinyin:      	"fántǐ zhōngwén",
		Senses:  			[]dicttypes.WordSense{},
	}
	oneWordDict := map[string]dicttypes.Word {
		s1: hw1,
		t1: hw1,
	}
	testCases := []struct {
		name string
		wdict  map[string]dicttypes.Word
		expectedLen int
	}{
		{
			name: "Empty dict",
			wdict: map[string]dicttypes.Word{}, 
			expectedLen: 0,
		},
		{
			name: "One word dict",
			wdict: oneWordDict, 
			expectedLen: 1,
		},
	}
	for _, tc := range testCases {
		words := getHeadwords(tc.wdict)
		if len(words) != tc.expectedLen {
			t.Errorf("TestGetHeadwords %s: Expected length of wdict got %d, want %d",
				tc.name, len(words), tc.expectedLen)
		}
	}
}

func TestReadText(t *testing.T) {
	var buf bytes.Buffer
	io.WriteString(&buf, "繁體中文")
	text := corpus.ReadText(&buf)
	expected := "繁體中文"
	if text != expected {
		t.Error("Expected ", expected, ", got ", text)
	}
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
	usageMap := map[string]*[]wordUsage{}
	usageMap = sampleUsage(usageMap)
	l := len(usageMap)
	expected := 0
	if l != expected {
		t.Error("Expected to get length ", expected, ", got ", l)
	}
}

// Basic test with minimal data
func TestSampleUsage2(t *testing.T) {
	wu := wordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "A Big Snake",
	}
	wuArray := []wordUsage{wu}
	usageMap := map[string]*[]wordUsage{"大": &wuArray}
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
	wu1 := wordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "蛇",
		Example:    "大蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "Some Snakes",
	}
	wu2 := wordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "蛇",
		Example:    "小蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 2",
		ColTitle:   "Some Snakes",
	}
	wuArray := []wordUsage{wu1, wu2}
	usageMap := map[string]*[]wordUsage{"蛇": &wuArray}
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
	wu1 := wordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大蛇",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "Some Big Animals",
	}
	wu2 := wordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大老虎",
		File:       "afile.txt",
		EntryTitle: "Scroll 2",
		ColTitle:   "Some Big Animals",
	}
	wu3 := wordUsage{
		Freq:       1,
		RelFreq:    0.01,
		Word:       "大",
		Example:    "大树",
		File:       "afile.txt",
		EntryTitle: "Scroll 1",
		ColTitle:   "Some Big Trees",
	}
	wuArray := []wordUsage{wu1, wu2, wu3}
	usageMap := map[string]*[]wordUsage{"大": &wuArray}
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
	var outBuf bytes.Buffer
	err := writeAnalysis(results, srcFile, glossFile, "Test Collection", "Test Doc",
			mockOutputConfig(), wdict, &outBuf)
	if err != nil {
		t.Errorf("could write analysis: %v", err)
	}
	t.Log("analysis.TestWriteAnalysis: End +++++++++++")
}

func TestWriteHwFile(t *testing.T) {
	const s = "繁体中文"
	const trad = "繁體中文"
	const p = "fántǐ zhōngwén"
	hw0 := dicttypes.Word{
		HeadwordId:  	1,
		Simplified:  	s,
		Traditional: 	trad,
		Pinyin:      	p,
		Senses:  			[]dicttypes.WordSense{},
	}
	dictEntry0 := DictEntry{
		Headword: hw0,
		RelevantDocs: nil,
		ContainsByDomain: nil,
		Contains: nil,
		Collocations: nil,
		UsageArr: nil,
		DateUpdated: "",
	}
	ws := dicttypes.WordSense{
		Simplified: s,
		Traditional: trad,
		Pinyin: p,
		English: "Traditional Chinese text",
		Grammar: "phrase",
		ConceptCN: "",
		Concept: "",
		DomainCN: "现代汉语",
		Domain: "Modern Chinese",
		SubdomainCN: "",
		Subdomain: "",
		Notes: "",
	}
	hw1 := dicttypes.Word{
		HeadwordId:  	1,
		Simplified:  	s,
		Traditional: 	trad,
		Pinyin:      	p,
		Senses:  			[]dicttypes.WordSense{ws},
	}
	dictEntry1 := DictEntry{
		Headword: hw1,
		RelevantDocs: nil,
		ContainsByDomain: nil,
		Contains: nil,
		Collocations: nil,
		UsageArr: nil,
		DateUpdated: "",
	}
	type test struct {
		name string
		dictEntry DictEntry
		wantToInclude string
		wantToNotInclude []string
  }
  tests := []test{
		{
			name: "No word senses",
			dictEntry: dictEntry0,
			wantToInclude: s,
			wantToNotInclude: []string{"hello"},
		},
		{
			name: "One word sense",
			dictEntry: dictEntry1,
			wantToInclude: s,
			wantToNotInclude: []string{"Subdomain", "Concept"},
		},
	}
  for _, tc := range tests {
		var buf bytes.Buffer
		templates := generator.NewTemplateMap(config.AppConfig{})
		tmpl, ok := templates["headword-template.html"]
		if !ok {
			t.Fatalf("%s, template not found", tc.name)
		}
		err := writeHwFile(&buf, tc.dictEntry, *tmpl)
		if err != nil {
			t.Fatalf("%s, Unexpected error: %v", tc.name, err)
		}
		got := buf.String()
		if len(got) == 0 {
			t.Fatalf("%s, no data written", tc.name)
		}
		if !strings.Contains(got, tc.wantToInclude) {
			t.Errorf("%s, got %s\n but wantToInclude %s", tc.name, got, tc.wantToInclude)
		}
		for _, notInclude := range tc.wantToNotInclude {
		  if strings.Contains(got, notInclude) {
			  t.Errorf("%s, got %s\n but wantToNotInclude %s", tc.name, got, notInclude)
		  }
		}
	}
}
