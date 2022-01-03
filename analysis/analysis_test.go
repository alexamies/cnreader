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
	"log"
	"strings"
	"testing"

	"github.com/alexamies/chinesenotes-go/bibnotes"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/generator"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/library"
)

func mockCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir:     "corpus",
		Excluded:      map[string]bool{},
		ProjectHome:   ".",
	}
}

func mockOutputConfig() generator.HTMLOutPutConfig {
	return generator.HTMLOutPutConfig{
		ContainsByDomain: "",
		Domain:           "",
		GoStaticDir:      "static",
		TemplateDir:      "templates",
		VocabFormat:      "",
		WebDir:           "web-staging",
		Templates:        generator.NewTemplateMap(config.AppConfig{}),
	}
}

func mockSmallDict() map[string]dicttypes.Word {
	s1 := "繁体中文"
	t1 := "繁體中文"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "fántǐ zhōngwén",
		Senses:      []dicttypes.WordSense{},
	}
	s2 := "前"
	t2 := "\\N"
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      "qián",
		Senses:      []dicttypes.WordSense{},
	}
	s3 := "不见"
	t3 := "不見"
	hw3 := dicttypes.Word{
		HeadwordId:  3,
		Simplified:  s3,
		Traditional: t3,
		Pinyin:      "bújiàn",
		Senses:      []dicttypes.WordSense{},
	}
	s4 := "古人"
	t4 := "\\N"
	hw4 := dicttypes.Word{
		HeadwordId:  4,
		Simplified:  s4,
		Traditional: t4,
		Pinyin:      "gǔrén",
		Senses:      []dicttypes.WordSense{},
	}
	s5 := "夫"
	t5 := "\\N"
	hw5 := dicttypes.Word{
		HeadwordId:  5,
		Simplified:  s5,
		Traditional: t5,
		Pinyin:      "fú fū",
		Senses:      []dicttypes.WordSense{},
	}
	s6 := "起信论"
	t6 := "起信論"
	hw6 := dicttypes.Word{
		HeadwordId:  6,
		Simplified:  s6,
		Traditional: t6,
		Pinyin:      "Qǐ Xìn Lùn",
		Senses:      []dicttypes.WordSense{},
	}
	s7 := "者"
	t7 := "\\N"
	hw7 := dicttypes.Word{
		HeadwordId:  7,
		Simplified:  s7,
		Traditional: t7,
		Pinyin:      "zhě zhuó",
		Senses:      []dicttypes.WordSense{},
	}
	s8 := "乃是"
	t8 := "\\N"
	hw8 := dicttypes.Word{
		HeadwordId:  8,
		Simplified:  s8,
		Traditional: t8,
		Pinyin:      "nǎishì",
		Senses:      []dicttypes.WordSense{},
	}
	return map[string]dicttypes.Word{
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

// Implements the CorpusLoader interface with no data
type mockCorpusLoader struct {
	collections []corpus.CollectionEntry
}

func (loader mockCorpusLoader) GetCollectionEntry(fName string) (*corpus.CollectionEntry, error) {
	return &corpus.CollectionEntry{}, nil
}

func (loader mockCorpusLoader) GetConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{}
}

func (loader mockCorpusLoader) LoadAll(r io.Reader) (*map[string]corpus.CorpusEntry, error) {
	data := map[string]corpus.CorpusEntry{}
	return &data, nil
}

func (loader mockCorpusLoader) LoadCollection(fName, colTitle string) (*[]corpus.CorpusEntry, error) {
	for _, col := range loader.collections {
		if col.CollectionFile == fName {
			return &col.CorpusEntries, nil
		}
	}
	return &[]corpus.CorpusEntry{}, nil
}

func (loader mockCorpusLoader) LoadCollections() (*[]corpus.CollectionEntry, error) {
	return &loader.collections, nil
}

func (loader mockCorpusLoader) LoadCorpus(r io.Reader) (*[]corpus.CollectionEntry, error) {
	return &loader.collections, nil
}

func (loader mockCorpusLoader) ReadText(src string) (string, error) {
	return "你好 Hello!", nil
}

// A mock LibraryLoader
type mockLibraryLoader struct {
	corpLoader corpus.CorpusLoader
}

func (loader mockLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return loader.corpLoader
}

func (loader mockLibraryLoader) LoadLibrary(r io.Reader) (*[]library.CorpusData, error) {
	data := []library.CorpusData{}
	return &data, nil
}

func TestContainsWord(t *testing.T) {
	testCases := []struct {
		name        string
		word        string
		headwords   []dicttypes.Word
		expectedLen int
	}{
		{
			name:        "Empty",
			word:        "中文",
			headwords:   []dicttypes.Word{},
			expectedLen: 0,
		},
	}
	for _, tc := range testCases {
		result := containsWord(tc.word, tc.headwords)
		if len(result) != tc.expectedLen {
			t.Errorf("TestContainsWord %s: got %d, want %d", tc.name, len(result),
				tc.expectedLen)
		}
	}
}

func TestGetChunks(t *testing.T) {
	testCases := []struct {
		name        string
		in          string
		expectedLen int
		expectedOut string
	}{
		{
			name:        "One chunk",
			in:          "中文",
			expectedLen: 1,
			expectedOut: "中文",
		},
		{
			name:        "Two chunks",
			in:          "a中文",
			expectedLen: 2,
			expectedOut: "a",
		},
		{
			name:        "Three chunks",
			in:          "a中文b",
			expectedLen: 3,
			expectedOut: "a",
		},
		{
			name:        "Simplified Chinese",
			in:          "简体中文",
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

func TestGetDocFrequencies(t *testing.T) {
	wdict := mockSmallDict()
	tok := tokenizer.DictTokenizer{
		WDict: wdict,
	}
	emptyCorpLoader := mockCorpusLoader{}
	emptyLibLoader := mockLibraryLoader{emptyCorpLoader}
	entry := corpus.CorpusEntry{
		RawFile:   "raw_file.txt",
		GlossFile: "gloss_file.html",
		Title:     "標題 Entry Title",
		ColTitle:  "Collection Title",
	}
	collection := corpus.CollectionEntry{
		CollectionFile: "collection.txt",
		GlossFile:      "collection.html",
		Title:          "My collection",
		Summary:        "xyz",
		Intro:          "abc",
		CorpusEntries:  []corpus.CorpusEntry{entry},
	}
	collections := []corpus.CollectionEntry{collection}
	smallCorpLoader := mockCorpusLoader{collections}
	smallLibLoader := mockLibraryLoader{smallCorpLoader}
	testCases := []struct {
		name      string
		libLoader library.LibraryLoader
		tok       tokenizer.Tokenizer
		wdict     map[string]dicttypes.Word
	}{
		{
			name:      "Empty",
			libLoader: emptyLibLoader,
			tok:       tok,
			wdict:     wdict,
		},
		{
			name:      "small",
			libLoader: smallLibLoader,
			tok:       tok,
			wdict:     wdict,
		},
	}
	for _, tc := range testCases {
		_, err := GetDocFrequencies(tc.libLoader, tc.tok, tc.wdict)
		if err != nil {
			t.Errorf("TestGetDocFrequencies %s: unexpected error %v", tc.name, err)
		}
	}
}

func TestGetHeadwords(t *testing.T) {
	s1 := "繁体中文"
	t1 := "繁體中文"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "fántǐ zhōngwén",
		Senses:      []dicttypes.WordSense{},
	}
	oneWordDict := map[string]dicttypes.Word{
		s1: hw1,
		t1: hw1,
	}
	testCases := []struct {
		name        string
		wdict       map[string]dicttypes.Word
		expectedLen int
	}{
		{
			name:        "Empty dict",
			wdict:       map[string]dicttypes.Word{},
			expectedLen: 0,
		},
		{
			name:        "One word dict",
			wdict:       oneWordDict,
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

func TestGetHwMap(t *testing.T) {
	s1 := "了"
	t1 := "\\N"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      "le",
		Senses:      []dicttypes.WordSense{},
	}
	oneWordDict := map[string]dicttypes.Word{
		s1: hw1,
	}
	testCases := []struct {
		name        string
		wdict       map[string]dicttypes.Word
		expectedLen int
	}{
		{
			name:        "One word dict",
			wdict:       oneWordDict,
			expectedLen: 1,
		},
	}
	for _, tc := range testCases {
		words := getHwMap(tc.wdict)
		if len(words) != tc.expectedLen {
			t.Errorf("TestGetHwMap %s, Expected length of map: got %d, want %d",
				tc.name, len(words), tc.expectedLen)
		}
	}
}

func TestGetWordFrequencies(t *testing.T) {
	wdict := mockSmallDict()
	tok := tokenizer.DictTokenizer{WDict: wdict}
	emptyCorpLoader := mockCorpusLoader{}
	emptyLibLoader := mockLibraryLoader{emptyCorpLoader}
	entry := corpus.CorpusEntry{
		RawFile:   "raw_file.txt",
		GlossFile: "gloss_file.html",
		Title:     "標題 Entry Title",
		ColTitle:  "Collection Title",
	}
	collection := corpus.CollectionEntry{
		CollectionFile: "collection.txt",
		GlossFile:      "collection.html",
		Title:          "My collection",
		Summary:        "xyz",
		Intro:          "abc",
		CorpusEntries:  []corpus.CorpusEntry{entry},
	}
	collections := []corpus.CollectionEntry{collection}
	smallCorpLoader := mockCorpusLoader{collections}
	smallLibLoader := mockLibraryLoader{smallCorpLoader}
	testCases := []struct {
		name      string
		libLoader library.LibraryLoader
	}{
		{
			name:      "Empty",
			libLoader: emptyLibLoader,
		},
		{
			name:      "Small",
			libLoader: smallLibLoader,
		},
	}
	for _, tc := range testCases {
		_, err := GetWordFrequencies(tc.libLoader, tok, wdict)
		if err != nil {
			t.Errorf("TestGetWordFrequencies, unexpected error, %s: %v", tc.name, err)
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
		name           string
		in             string
		expectedFirst  string
		expectedTokens int
		expectedVocab  int
		expectedWC     int
		expectedCC     int
	}{
		{
			name:           "One token",
			in:             "繁體中文",
			expectedFirst:  "繁體中文",
			expectedTokens: 1,
			expectedVocab:  1,
			expectedWC:     1,
			expectedCC:     4,
		},
		{
			name:           "ASCII and one token",
			in:             "a繁體中文",
			expectedFirst:  "a",
			expectedTokens: 2,
			expectedVocab:  1,
			expectedWC:     1,
			expectedCC:     4,
		},
		{
			name:           "Three tokens",
			in:             "前不见古人",
			expectedFirst:  "前",
			expectedTokens: 3,
			expectedVocab:  3,
			expectedWC:     3,
			expectedCC:     5,
		},
		{
			name:           "More tokens",
			in:             "夫起信論者，乃是...。",
			expectedFirst:  "夫",
			expectedTokens: 6,
			expectedVocab:  4,
			expectedWC:     4,
			expectedCC:     7,
		},
	}
	wdict := mockSmallDict()
	tok := tokenizer.DictTokenizer{WDict: wdict}
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
		HeadwordId:  1,
		Simplified:  s,
		Traditional: trad,
		Pinyin:      p,
		Senses:      []dicttypes.WordSense{},
	}
	dictEntry0 := DictEntry{
		Headword:         hw0,
		RelevantDocs:     nil,
		ContainsByDomain: nil,
		Contains:         nil,
		Collocations:     nil,
		UsageArr:         nil,
		DateUpdated:      "",
	}
	ws := dicttypes.WordSense{
		Simplified:  s,
		Traditional: trad,
		Pinyin:      p,
		English:     "Traditional Chinese text",
		Grammar:     "phrase",
		ConceptCN:   "",
		Concept:     "",
		DomainCN:    "现代汉语",
		Domain:      "Modern Chinese",
		SubdomainCN: "",
		Subdomain:   "",
		Notes:       "",
	}
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s,
		Traditional: trad,
		Pinyin:      p,
		Senses:      []dicttypes.WordSense{ws},
	}
	dictEntry1 := DictEntry{
		Headword:         hw1,
		RelevantDocs:     nil,
		ContainsByDomain: nil,
		Contains:         nil,
		Collocations:     nil,
		UsageArr:         nil,
		DateUpdated:      "",
	}
	type test struct {
		name             string
		dictEntry        DictEntry
		wantToInclude    string
		wantToNotInclude []string
	}
	tests := []test{
		{
			name:             "No word senses",
			dictEntry:        dictEntry0,
			wantToInclude:    s,
			wantToNotInclude: []string{"hello"},
		},
		{
			name:             "One word sense",
			dictEntry:        dictEntry1,
			wantToInclude:    s,
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
			t.Fatalf("TestWriteHwFile: %s, Unexpected error: %v", tc.name, err)
		}
		got := buf.String()
		if len(got) == 0 {
			t.Fatalf("TestWriteHwFile: %s, no data written", tc.name)
		}
		if !strings.Contains(got, tc.wantToInclude) {
			t.Errorf("TestWriteHwFile: %s, got %s\n but wantToInclude %s", tc.name,
				got, tc.wantToInclude)
		}
		for _, notInclude := range tc.wantToNotInclude {
			if strings.Contains(got, notInclude) {
				t.Errorf("TestWriteHwFile %s, got %s\n but wantToNotInclude %s",
					tc.name, got, notInclude)
			}
		}
	}
}

// testHwWriter writes to a bytes buffer instead of a file
type testHwWriter struct {
	buf        *bytes.Buffer
	numWritten *int
}

// OpenWriter opens the file to write HTML
func (w testHwWriter) NewWriter(hwId int) io.Writer {
	*w.numWritten++
	return w.buf
}

// CloseWriter does nothing
func (w testHwWriter) CloseWriter(hwId int) {
	log.Printf("testHwWriter: CloseWriter called with hw.Id %d", hwId)
}

func TestWriteHwFiles(t *testing.T) {
	loader := mockLibraryLoader{}
	indexState := index.IndexState{}
	vocabAnalysis := VocabAnalysis{}
	oneWordDict := make(map[string]dicttypes.Word)
	const match = `"(T ([0-9]))(\)|,|;)","(T ([0-9]{2}))(\)|,|;)","(T ([0-9]{3}))(\)|,|;)","(T ([0-9]{4}))(\)|,|;)"`
	const replace = `"<a href="/taisho/t000${2}.html">${1}</a>${3}","<a href="/taisho/t00${2}.html">${1}</a>${3}","<a href="/taisho/t0${2}.html">${1}</a>${3}","<a href="/taisho/t${2}.html">${1}</a>${3}"`
	config := generator.HTMLOutPutConfig{
		ContainsByDomain: "",
		Domain:           "",
		GoStaticDir:      "static",
		TemplateDir:      "templates",
		VocabFormat:      "",
		WebDir:           "web-staging",
		Templates:        generator.NewTemplateMap(config.AppConfig{}),
		NotesReMatch:     match,
		NotesReplace:     replace,
	}
	const s = "金刚经"
	const tr = "金剛經"
	const p = "Jīngāng Jīng"
	ws := dicttypes.WordSense{
		Simplified:  s,
		Traditional: tr,
		Pinyin:      p,
		English:     "Diamond Sutra",
		Notes:       "(T 235)",
	}
	hw := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s,
		Traditional: tr,
		Pinyin:      p,
		Senses:      []dicttypes.WordSense{ws},
	}
	oneWordDict[s] = hw
	oneWordDict[tr] = hw
	type test struct {
		name      string
		wdict     map[string]dicttypes.Word
		config    generator.HTMLOutPutConfig
		expectNum int
		contains  string
	}
	tests := []test{
		{
			name:      "Empty",
			wdict:     map[string]dicttypes.Word{},
			config:    mockOutputConfig(),
			expectNum: 0,
			contains:  "",
		},
		{
			name:      "Small",
			wdict:     mockSmallDict(),
			config:    mockOutputConfig(),
			expectNum: 8,
			contains:  "",
		},
		{
			name:      "Has notes",
			wdict:     oneWordDict,
			config:    mockOutputConfig(),
			expectNum: 1,
			contains:  "T 235",
		},
		{
			name:      "Transforms notes",
			wdict:     oneWordDict,
			config:    config,
			expectNum: 1,
			contains:  `<a href="/taisho/t0235.html">T 235</a>`,
		},
	}
	for _, tc := range tests {
		tok := tokenizer.DictTokenizer{WDict: tc.wdict}
		var buf bytes.Buffer
		numWritten := 0
		tw := testHwWriter{
			buf:        &buf,
			numWritten: &numWritten,
		}
  	ref2FileReader := strings.NewReader("")
  	refNo2ParallelReader := strings.NewReader("")
  	refNo2TransReader := strings.NewReader("")
  	bibNotesClient, err := bibnotes.LoadBibNotes(ref2FileReader, refNo2ParallelReader, refNo2TransReader)
  	if err != nil {
  		t.Fatalf("TestWriteHwFiles error loading bibnotes: %v", err)
  	}
		hWFileDependencies := HWFileDependencies {
			Loader: loader,
			DictTokenizer: tok,
			OutputConfig: tc.config,
			IndexState: indexState,
			Wdict: tc.wdict,
			VocabAnalysis: vocabAnalysis,
			Hww: tw,
			BibNotesClient: bibNotesClient,
		}
		err = WriteHwFiles(hWFileDependencies)
		if err != nil {
			t.Fatalf("TestWriteHwFiles: %s, Unexpected error: %v", tc.name, err)
		}
		got := buf.String()
		if *tw.numWritten != tc.expectNum {
			t.Fatalf("TestWriteHwFiles %s, Got numWritten = %d, want: %d, buf:\n%s",
				tc.name, tw.numWritten, tc.expectNum, got)
		}
		if !strings.Contains(got, tc.contains) {
			t.Fatalf("TestWriteHwFiles %s, did not contain expected string %s buf:\n%s",
				tc.name, tc.contains, buf.String())
		}
	}
}
