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

// Package for vocabulary analysis of a monolingual Chinese text corpus
//
// This includes
// - reading the corpus documents from disk
// - tokenization of the corpus into multi-character arrays
// - computation of term and bigram frequencies
// - compilation of an index for later full text search
// - computation of term occurrence and usage in the corpus
package analysis

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/alexamies/chinesenotes-go/bibnotes"
	"github.com/alexamies/chinesenotes-go/config"
	"github.com/alexamies/chinesenotes-go/dictionary"
	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/generator"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/library"
	"github.com/alexamies/cnreader/ngram"
)

// Maximum number of word frequency entries to output to the generated
// HTML file
const maxWFOutput = 500

// Maximum number of unknwon characters to output to the generated
// HTML file
const maxUnknownOutput = 50

// Max usage elements for a word
const maxUsage = 25

// Max number of occurrences of same title in a list of word usages
const maxTitle = 5

// Max number of keywords to display
const maxKeywords = 10

// Maximum number of containing words to output to the generated
// HTML file
const maxContains = 50

// hwWriter manages files for writing headwords to HTML
type HeadwordWriter interface {
	NewWriter(hwId int) io.Writer
	CloseWriter(hwId int)
}

// analysisResults holds vocabulary analysis for a corpus text
type analysisResults struct {
	Title                   string
	WC, UniqueWords, CCount int
	ProperNouns             []dicttypes.Word
	DocumentGlossary        Glossary
	TopKeywords             dicttypes.Words
	WordFrequencies         []wFResult
	LexicalWordFreq         []wFResult
	BigramFreqSorted        []ngram.BigramFreq
	UnkownnChars            []index.SortedWordItem
	DateUpdated             string
	MaxWFOutput             int
}

// VocabAnalysis bundles up vocabulary analysis
type VocabAnalysis struct {
	UsageMap     map[string]*[]wordUsage
	WFTotal      map[*index.CorpusWord]index.CorpusWordFreq
	WCTotal      map[string]int
	Collocations ngram.CollocationMap
}

// DictEntry holds content used for writing a dictionary entry to HTML
type DictEntry struct {
	Title            string
	Headword         dicttypes.Word
	RelevantDocs     []index.RetrievalResult
	ContainsByDomain []dicttypes.Word
	Contains         []dicttypes.Word
	Collocations     []ngram.BigramFreq
	UsageArr         []wordUsage
	DateUpdated      string
}

// wordUsage holds details of Word usage in the corpus
type wordUsage struct {
	Freq                                      int
	RelFreq                                   float64
	Word, Example, File, EntryTitle, ColTitle string
}

// wFResult holds results of vocabulary analysis entry for a single word
type wFResult struct {
	Freq, HeadwordId                int
	Chinese, Pinyin, English, Usage string
}

type HWFileDependencies struct {
	Loader         library.LibraryLoader
	DictTokenizer  tokenizer.Tokenizer
	OutputConfig   generator.HTMLOutPutConfig
	IndexState     index.IndexState
	Dict           *dictionary.Dictionary
	VocabAnalysis  VocabAnalysis
	Hww            HeadwordWriter
	BibNotesClient bibnotes.BibNotesClient
}

type bilingualEntryMeta struct {

	// Name of file with bilingual parallel text
	ParallelTextFile string

	// Name of file with bilingual text
	BilingualTextFile string

	// Name of file with Chinese text, in case this is a bilingual or parallel file
	ChineseTextFile string
}

// containsWord gets a list of words that contain the given word
func containsWord(word string, headwords []dicttypes.Word) []dicttypes.Word {
	//log.Printf("dictionary.containsWord: Enter")
	contains := []dicttypes.Word{}
	for _, hw := range headwords {
		if len(contains) <= maxContains && hw.Simplified != word && strings.Contains(hw.Simplified, word) {
			contains = append(contains, hw)
		}
	}
	return contains
}

func getBilingualEntryMeta(bibClient bibnotes.BibNotesClient, collectionFile, glossFile string, colMap map[string]*[]corpus.CorpusEntry) bilingualEntryMeta {
	if bibClient == nil || len(collectionFile) == 0 {
		return bilingualEntryMeta{}
	}
	transRefs := bibClient.GetTransRefs(collectionFile)
	if len(transRefs) == 0 {
		return bilingualEntryMeta{}
	}
	parallelCollectionFile := ""
	if transRefs[0].Kind == "parallel" {
		parallelCollectionFile = transRefs[0].URL
	}
	parallelTextFile := ""
	// Check that the paralell text file really exists
	if colMap != nil {
		if strings.Contains(glossFile, "_en_aligned.html") {
			// Navigate from a parallel text file to the regular file
			return bilingualEntryMeta{
				ChineseTextFile: strings.Replace(glossFile, "_en_aligned.html", ".html", 1),
			}
		}
		pTextFile := strings.Replace(glossFile, ".html", "_en_aligned.html", 1)
		log.Printf("analysis.getBilingualEntryMeta: glossFile = %s, pTextFile = %s, parallelCollectionFile = %s", glossFile, pTextFile, parallelCollectionFile)
		parallelEntries, ok := colMap[parallelCollectionFile]
		if ok {
			log.Printf("analysis.getBilingualEntryMeta: glossFile = %s, len(*parallelEntries) = %d", glossFile, len(*parallelEntries))
			for _, pEntry := range *parallelEntries {
				log.Printf("analysis.getBilingualEntryMeta: pEntry.GlossFile = %s, pTextFile = %s", pEntry.GlossFile, pTextFile)
				if pEntry.GlossFile == pTextFile {
					parallelTextFile = pTextFile
					break
				}
			}
		} else {
			log.Printf("analysis.getBilingualEntryMeta: did not find parallelCol for pIndexFile = %s in colMap of %d entries", pIndexFile, len(colMap))
		}
	} else {
		log.Printf("analysis.getBilingualEntryMeta: colMap == nil")
	}
	log.Printf("analysis.getBilingualEntryMeta: len(transRefs) = %d, collectionFile = %s, parallelFile = %s, kind = %s, glossFile = %s, parallelTextFile = %s", len(transRefs), collectionFile,
		parallelCollectionFile, transRefs[0], glossFile, parallelTextFile)
	return bilingualEntryMeta{
		ParallelTextFile: parallelTextFile,
	}
}

// getHeadwords compute headword numbers for all lexical units listed in data/words.txt
// Return a sorted array of headwords
func getHeadwords(dict *dictionary.Dictionary) []dicttypes.Word {
	hwArray := []dicttypes.Word{}
	for _, w := range dict.HeadwordIds {
		hwArray = append(hwArray, *w)
	}
	log.Printf("analysis.GetHeadwords: hwcount = %d", len(hwArray))
	return hwArray
}

// isCJKChar tests whether the symbol is a CJK character, excluding punctuation
// Only looks at the first charater in the string
func isCJKChar(character string) bool {
	r := []rune(character)
	return unicode.Is(unicode.Han, r[0]) && !unicode.IsPunct(r[0])
}

// getChunks tokenizes text into a list of CJK and non CJK strings
func getChunks(text string) list.List {
	var chunks list.List
	cjk := ""
	noncjk := ""
	for _, character := range text {
		if isCJKChar(string(character)) {
			if noncjk != "" {
				chunks.PushBack(noncjk)
				noncjk = ""
			}
			cjk += string(character)
		} else if cjk != "" {
			chunks.PushBack(cjk)
			cjk = ""
			noncjk += string(character)
		} else {
			noncjk += string(character)
		}
	}
	if cjk != "" {
		chunks.PushBack(cjk)
	}
	if noncjk != "" {
		chunks.PushBack(noncjk)
	}
	return chunks
}

// getWordFrequencies compute word doc frequencies for corpus
func GetDocFrequencies(libLoader library.LibraryLoader,
	dictTokenizer tokenizer.Tokenizer,
	dict *dictionary.Dictionary) (*index.DocumentFrequency, error) {
	log.Printf("analysis.GetDocFrequencies: enter")
	df := index.NewDocumentFrequency()
	corpLoader := libLoader.GetCorpusLoader()
	corpusConfig := corpLoader.GetConfig()
	collectionEntries, err := corpLoader.LoadCollections()
	if err != nil {
		return nil, fmt.Errorf("GetDocFrequencies: Error loading corpus: %v", err)
	}
	for _, col := range *collectionEntries {
		colFile := col.CollectionFile
		corpusEntries, err := corpLoader.LoadCollection(colFile, col.Title)
		if err != nil {
			return nil, fmt.Errorf("GetDocFrequencies: Error loading collection %s: %v",
				colFile, err)
		}
		for _, entry := range *corpusEntries {
			text, err := corpLoader.ReadText(entry.RawFile)
			if err != nil {
				return nil, fmt.Errorf("GetDocFrequencies: Error reading file %s: %v",
					entry.RawFile, err)
			}
			_, results := ParseText(text, col.Title, &entry, dictTokenizer,
				corpusConfig, dict)
			df.AddDocFreq(results.DocFreq)
		}
	}
	return &df, nil
}

// getWordFrequencies compute word frequencies, collocations, and usage for corpus
func GetWordFrequencies(libLoader library.LibraryLoader,
	dictTokenizer tokenizer.Tokenizer,
	dict *dictionary.Dictionary) (*VocabAnalysis, error) {

	log.Printf("analysis.getWordFrequencies: enter")

	// Overall word frequencies per corpus
	collocations := ngram.CollocationMap{}
	usageMap := map[string]*[]wordUsage{}
	ccount := 0 // character count
	wcTotal := map[string]int{}
	wfTotal := map[*index.CorpusWord]index.CorpusWordFreq{}

	corpLoader := libLoader.GetCorpusLoader()
	corpusConfig := corpLoader.GetConfig()
	collectionEntries, err := corpLoader.LoadCollections()
	if err != nil {
		return nil, fmt.Errorf("getWordFrequencies: Error reading collections: %v",
			err)
	}
	for _, col := range *collectionEntries {
		colFile := col.CollectionFile
		corpusEntries, err := corpLoader.LoadCollection(colFile, col.Title)
		if err != nil {
			return nil, fmt.Errorf("getWordFrequencies: Error loading col %s: %v",
				colFile, err)
		}
		for _, entry := range *corpusEntries {
			text, err := corpLoader.ReadText(entry.RawFile)
			if err != nil {
				return nil, fmt.Errorf("getWordFrequencies: Error opening file %s: %v",
					entry.RawFile, err)
			}
			ccount += utf8.RuneCountInString(text)
			_, results := ParseText(text, col.Title, &entry, dictTokenizer,
				corpusConfig, dict)
			wcTotal[col.Corpus] += results.WC

			// Process collocations
			collocations.MergeCollocationMap(results.Collocations)

			// Find word usage for the given word
			for word, count := range results.Vocab {
				cw := &index.CorpusWord{
					Corpus: col.Corpus,
					Word:   word,
				}
				cwf := &index.CorpusWordFreq{
					Corpus: col.Corpus,
					Word:   word,
					Freq:   count}
				if cwfPrev, found := wfTotal[cw]; found {
					cwf.Freq += cwfPrev.Freq
				}
				wfTotal[cw] = *cwf
				rel_freq := 1000.0 * float64(count) / float64(results.WC)
				usage := wordUsage{cwf.Freq, rel_freq, word, results.Usage[word],
					entry.GlossFile, entry.Title, col.Title}
				usageArr, ok := usageMap[word]
				if !ok {
					usageArr = new([]wordUsage)
					usageMap[word] = usageArr
				}
				*usageArr = append(*usageArr, usage)
				//fmt.Fprintf(w, "%s\t%d\t%f\t%s\t%s\t%s\t%s\n", word, count, rel_freq,
				//	entry.GlossFile, col.Title, entry.Title, usage[word])
			}
		}
	}

	usageMap = sampleUsage(usageMap)

	// Print out totals for each corpus
	for corpus, count := range wcTotal {
		log.Printf("WordFrequencies: Total word count for corpus %s: %d\n",
			corpus, count)
	}
	log.Printf("WordFrequencies: len(collocations) = %d\n", len(collocations))
	log.Printf("WordFrequencies: character count = %d\n", ccount)

	return &VocabAnalysis{usageMap, wfTotal, wcTotal, collocations}, nil
}

// ParseText tokenizes a Chinese text corpus document into terms
// Parameters:
//
//	text: the string to parse
//	ColTitle: Optional parameter used for tracing collocation usage
//	document: Optional parameter used for tracing collocation usage
//
// Returns:
//
//	tokens: the tokens for the parsed text
//	results: vocabulary analysis results
func ParseText(text string, colTitle string, document *corpus.CorpusEntry, dictTokenizer tokenizer.Tokenizer, corpusConfig corpus.CorpusConfig, dict *dictionary.Dictionary) (list.List, *CollectionAResults) {
	tokens := list.List{}
	vocab := map[string]int{}
	bigrams := map[string]int{}
	bigramMap := ngram.BigramFreqMap{}
	collocations := ngram.CollocationMap{}
	unknownChars := map[string]int{}
	usage := map[string]string{}
	wc := 0
	cc := 0
	chunks := getChunks(text)
	lastHWPtr := &dicttypes.Word{}
	lastHW := lastHWPtr
	lastHWText := ""
	hwIdMap := dict.HeadwordIds
	// log.Printf("ParseText: For coll %s, doc %s got %d chunks\n", colTitle, document.RawFile, chunks.Len())
	for e := chunks.Front(); e != nil; e = e.Next() {
		chunk := e.Value.(string)
		//fmt.Printf("ParseText: chunk %s\n", chunk)
		characters := strings.Split(chunk, "")
		if !isCJKChar(characters[0]) || corpus.IsExcluded(corpusConfig.Excluded, chunk) {
			tokens.PushBack(chunk)
			lastHWPtr = &dicttypes.Word{}
			lastHW = lastHWPtr
			lastHWText = ""
			continue
		}
		textTokens := dictTokenizer.Tokenize(chunk)
		//fmt.Printf("ParseText: len(tokens) %d\n", len(tokens))
		for _, token := range textTokens {
			w := token.Token
			if !corpus.IsExcluded(corpusConfig.Excluded, w) {
				tokens.PushBack(w)
				wc++
				cc += utf8.RuneCountInString(w)
				vocab[w]++
				if lastHWText != "" {
					bg := lastHWText + w
					bigrams[bg]++
				}
				lastHWText = w
				if _, ok := usage[w]; !ok {
					usage[w] = chunk
				}
				hwid := token.DictEntry.HeadwordId
				hw := hwIdMap[hwid]
				if lastHW != nil && lastHW.HeadwordId != 0 {
					if len(hw.Senses) == 0 {
						log.Printf("ParseText: WordSenses nil for %s , id = %d, in %s, %s\n", w, hwid, document.Title, colTitle)
					}
					bigram, ok := bigramMap.GetBigramVal(lastHW.HeadwordId, hwid)
					if !ok {
						bigram = ngram.NewBigram(*lastHW, *hw, chunk, document.GlossFile, document.Title, colTitle)
					}
					bigramMap.PutBigram(bigram)
					collocations.PutBigram(bigram.HeadwordDef1.HeadwordId, bigram)
					collocations.PutBigram(bigram.HeadwordDef2.HeadwordId, bigram)
				}
				lastHW = hw
			}
		}
	}
	dl := index.DocLength{
		GlossFile: document.GlossFile,
		WordCount: wc}
	dlArray := []index.DocLength{dl}
	results := CollectionAResults{
		Vocab:             vocab,
		Bigrams:           bigrams,
		Usage:             usage,
		BigramFrequencies: bigramMap,
		Collocations:      collocations,
		WC:                wc,
		CCount:            cc,
		UnknownChars:      unknownChars,
		DocLengthArray:    dlArray,
	}
	// log.Printf("ParseText: Done coll %s, doc %s got %d tokens\n", colTitle, document.RawFile, tokens.Len())
	return tokens, &results
}

// sampleUsage finds word usage for usability, also making sure that the list of
// word usage samples is not dominated by any one title and truncating at
// maxUsage examples.
func sampleUsage(usageMap map[string]*[]wordUsage) map[string]*[]wordUsage {
	for word, usagePtr := range usageMap {
		sampleMap := map[string]int{}
		usage := *usagePtr
		usageCapped := new([]wordUsage)
		j := 0
		for _, wu := range usage {
			count := sampleMap[wu.ColTitle]
			if count < maxTitle && j < maxUsage {
				*usageCapped = append(*usageCapped, wu)
				sampleMap[wu.ColTitle]++
				j++
			}
		}
		usageMap[word] = usageCapped
	}
	return usageMap
}

// writeAnalysisCorpus writes out an analysis of the entire corpus, including
// word frequencies and other data. The output file is called
// 'corpus-analysis.html' in the web/analysis directory.
// Parameters:
//
//	results: The results of corpus analysis
//	docFreq: document frequency for terms
//
// Returns: the name of the file written to
func writeAnalysisCorpus(results *CollectionAResults,
	docFreq index.DocumentFrequency, outputConfig generator.HTMLOutPutConfig,
	indexConfig index.IndexConfig, wdict map[string]*dicttypes.Word,
	c config.AppConfig, analysisFile io.Writer) error {

	// Parse template and organize template parameters
	sortedWords := index.SortedFreq(results.Vocab)
	wfResults := results.GetWordFreq(sortedWords, wdict)
	maxWf := len(wfResults)
	if maxWf > maxWFOutput {
		maxWf = maxWFOutput
	}

	lexicalWordFreq := results.GetLexicalWordFreq(sortedWords, wdict)
	maxLex := len(lexicalWordFreq)
	if maxLex > maxWFOutput {
		maxLex = maxWFOutput
	}

	sortedUnknownWords := index.SortedFreq(results.UnknownChars)
	maxUnknown := len(sortedUnknownWords)
	if maxUnknown > maxUnknownOutput {
		maxUnknown = maxUnknownOutput
	}

	// Bigrams, also truncated
	bFreq := ngram.SortedFreq(results.BigramFrequencies)
	maxBFOutput := len(bFreq)
	if maxBFOutput > maxWFOutput {
		maxBFOutput = maxWFOutput
	}

	dateUpdated := time.Now().Format("2006-01-02")
	title := "Terminology Extraction and Vocabulary Analysis"
	aResults := analysisResults{
		Title:            title,
		WC:               results.WC,
		CCount:           results.CCount,
		ProperNouns:      dicttypes.Words{},
		DocumentGlossary: MakeGlossary("", []dicttypes.Word{}),
		TopKeywords:      dicttypes.Words{},
		UniqueWords:      len(results.Vocab),
		WordFrequencies:  wfResults[:maxWf],
		LexicalWordFreq:  lexicalWordFreq[:maxLex],
		BigramFreqSorted: bFreq[:maxBFOutput],
		UnkownnChars:     sortedUnknownWords[:maxUnknown],
		DateUpdated:      dateUpdated,
		MaxWFOutput:      len(wfResults),
	}
	const tmplFile = "corpus-summary-analysis-template.html"
	tmpl, ok := outputConfig.Templates[tmplFile]
	if !ok {
		return fmt.Errorf("writeAnalysisCorpus: no template found for %s", tmplFile)
	}
	w := bufio.NewWriter(analysisFile)
	if err := tmpl.Execute(w, aResults); err != nil {
		return fmt.Errorf("writeAnalysisCorpus: error executing summary analysis template%v", err)
	}
	w.Flush()

	// Write results to plain text files
	fname := indexConfig.IndexDir + "/" + index.WfCorpusFile
	wfFile, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("could not open wfFile: %v", err)
	}
	defer wfFile.Close()
	unknownCharsFile, err := os.Create(indexConfig.IndexDir + "/" + index.UnknownCharsFile)
	if err != nil {
		return fmt.Errorf("could not open write unknownCharsFile: %v", err)
	}
	defer unknownCharsFile.Close()
	ngramFN := indexConfig.IndexDir + "/" + index.NgramCorpusFile
	ngramFile, err := os.Create(ngramFN)
	if err != nil {
		return fmt.Errorf("could not open write ngramFile: %v", err)
	}
	defer ngramFile.Close()
	wordFreqStore := index.WordFreqStore{
		WFWriter:           wfFile,
		UnknownCharsWriter: unknownCharsFile,
		BigramWriter:       ngramFile}
	err = index.WriteWFCorpus(wordFreqStore, sortedWords, sortedUnknownWords,
		bFreq, results.WC, indexConfig)
	if err != nil {
		return fmt.Errorf("could not write inddex: %v", err)
	}

	return nil
}

// writeAnalysis writes a document with vocabulary analysis of the text. The
// name of the output file will be source file with '-analysis' appended,
// placed in the web/analysis directory
// results: The results of vocabulary analysis
// collectionTitle: The title of the whole colleciton
// docTitle: The title of this specific document
// Returns the name of the file written to
func writeAnalysis(results *CollectionAResults, srcFile, glossFile,
	collectionTitle, docTitle string,
	outputConfig generator.HTMLOutPutConfig,
	wdict map[string]*dicttypes.Word, f io.Writer) error {

	// Parse template and organize template parameters
	properNouns := makePNList(results.Vocab, wdict)

	domain_label := outputConfig.Domain
	//log.Printf("analysis.writeAnalysis: domain_label: %s\n", domain_label)

	sortedWords := index.SortedFreq(results.Vocab)
	//log.Printf("analysis.writeAnalysis: found sortedWords for %s, count %d\n",
	//	srcFile, len(sortedWords))

	glossary := MakeGlossary(domain_label, results.GetHeadwords(wdict))

	wfResults := results.GetWordFreq(sortedWords, wdict)
	maxWf := len(wfResults)
	if maxWf > maxWFOutput {
		maxWf = maxWFOutput
	}

	lexicalWordFreq := results.GetLexicalWordFreq(sortedWords, wdict)
	maxLex := len(lexicalWordFreq)
	if maxLex > maxWFOutput {
		maxLex = maxWFOutput
	}

	topKeywords := []dicttypes.Word{}
	if domain_label != "" {
		topKeywords = index.FilterByDomain(sortedWords, domain_label, wdict)
		m := len(topKeywords)
		if m > maxKeywords {
			m = maxKeywords
		}
		topKeywords = topKeywords[:m]
	}

	//log.Printf("analysis.writeAnalysis: title: %s, len topKeywords: %d, " +
	//	"domain_label: %s\n", docTitle, len(topKeywords), domain_label)

	sortedUnknownWords := index.SortedFreq(results.UnknownChars)
	maxUnknown := len(sortedUnknownWords)
	if maxUnknown > maxUnknownOutput {
		maxUnknown = maxUnknownOutput
	}

	// Bigrams, also truncated
	bFreq := ngram.SortedFreq(results.BigramFrequencies)
	maxBFOutput := len(bFreq)
	if maxBFOutput > maxWFOutput {
		maxBFOutput = maxWFOutput
	}

	dateUpdated := time.Now().Format("2006-01-02")
	title := "Glossary and Vocabulary for " + collectionTitle
	if docTitle != "" {
		title += ", " + docTitle
	}

	aResults := analysisResults{
		Title:            title,
		WC:               results.WC,
		CCount:           results.CCount,
		ProperNouns:      properNouns,
		DocumentGlossary: glossary,
		TopKeywords:      topKeywords,
		UniqueWords:      len(results.Vocab),
		WordFrequencies:  wfResults[:maxWf],
		LexicalWordFreq:  lexicalWordFreq[:maxLex],
		BigramFreqSorted: bFreq[:maxBFOutput],
		UnkownnChars:     sortedUnknownWords[:maxUnknown],
		DateUpdated:      dateUpdated,
		MaxWFOutput:      len(wfResults),
	}
	tmpl, ok := outputConfig.Templates["corpus-analysis-template.html"]
	if !ok {
		return fmt.Errorf("corpus-analysis-template.html not found")
	}

	// Write output
	w := bufio.NewWriter(f)
	err := tmpl.Execute(w, aResults)
	if err != nil {
		return fmt.Errorf("analysis.writeAnalysis could not execute template: %v", err)
	}
	w.Flush()

	return nil
}

// writeCollection writes a corpus document collection to HTML, including all
// the entries contained in the collection
// collectionEntry: the CollectionEntry struct
// baseDir: The base directory to use
func writeCollection(collectionEntry *corpus.CollectionEntry, outputConfig generator.HTMLOutPutConfig, libLoader library.LibraryLoader,
	dictTokenizer tokenizer.Tokenizer, dict *dictionary.Dictionary, c config.AppConfig, bibClient bibnotes.BibNotesClient, colMap map[string]*[]corpus.CorpusEntry) (*CollectionAResults, error) {

	log.Printf("analysis.writeCollection: enter CollectionFile =" + collectionEntry.CollectionFile)
	corpLoader := libLoader.GetCorpusLoader()
	corpusConfig := corpLoader.GetConfig()
	cFile := collectionEntry.CollectionFile
	corpusEntries, err := corpLoader.LoadCollection(cFile, collectionEntry.Title)
	if err != nil {
		return nil, fmt.Errorf("analysis.writeCollection error loading file %s: %v", cFile, err)
	}
	aResults := NewCollectionAResults()
	wdict := dict.Wdict
	for _, entry := range *corpusEntries {
		text, err := corpLoader.ReadText(entry.RawFile)
		if err != nil {
			return nil, fmt.Errorf("analysis.writeCollection error reading src %s: %v", entry.RawFile, err)
		}
		_, results := ParseText(text, collectionEntry.Title, &entry, dictTokenizer, corpusConfig, dict)

		srcFile := entry.RawFile
		i := strings.Index(srcFile, ".")
		if i <= 0 {
			return nil, fmt.Errorf("writeCollection: Bad name for source file: %s", srcFile)
		}
		basename := srcFile[:i] + "_analysis.html"
		analysisDir := c.ProjectHome + "/" + outputConfig.WebDir + "/analysis/"
		filename := analysisDir + basename
		af, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("writeCollection: count not create analysis file %s: %v", filename, err)
		}
		err = writeAnalysis(results, entry.RawFile, entry.GlossFile, collectionEntry.Title, entry.Title, outputConfig, wdict, af)
		af.Close()
		if err != nil {
			return nil, fmt.Errorf("writeCollection could write analysis: %v", err)
		}

		sourceFormat := "TEXT"
		if strings.HasSuffix(entry.RawFile, ".html") {
			sourceFormat = "HTML"
		}
		dest := c.ProjectHome + "/" + outputConfig.WebDir + "/" + entry.GlossFile
		df, err := os.Create(dest)
		if err != nil {
			return nil, fmt.Errorf("writeCollection could not open dest file %s: %v",
				dest, err)
		}
		w := bufio.NewWriter(df)

		textTokens := dictTokenizer.Tokenize(text)
		// log.Printf("writeCollection: writing corpus doc, entry.RawFile = %s, got %d tokens, analysis file: %s",
		//	entry.RawFile, len(textTokens), basename)
		bilingualMeta := getBilingualEntryMeta(bibClient, collectionEntry.CollectionFile, entry.GlossFile, colMap)
		entryMeta := generator.CorpusEntryMeta{
			CollectionURL:     collectionEntry.GlossFile,
			CollectionTitle:   collectionEntry.Title,
			EntryTitle:        entry.Title,
			AnalysisFile:      basename,
			ParallelTextFile:  bilingualMeta.ParallelTextFile,
			BilingualTextFile: bilingualMeta.BilingualTextFile,
			ChineseTextFile:   bilingualMeta.ChineseTextFile,
		}
		err = generator.WriteCorpusDoc(textTokens, results.Vocab, w,
			entryMeta, sourceFormat, outputConfig,
			corpusConfig, wdict)
		w.Flush()
		df.Close()
		if err != nil {
			return nil, fmt.Errorf("writeCollection, error writing corpus doc: %v", err)
		}
		aResults.AddResults(results)
		aResults.DocFreq.AddVocabulary(results.Vocab)
		aResults.BigramDF.AddVocabulary(results.Bigrams)
		aResults.WFDocMap.AddWF(results.Vocab, collectionEntry.GlossFile,
			entry.GlossFile, results.WC)
		aResults.BigramDocMap.AddWF(results.Bigrams, collectionEntry.GlossFile,
			entry.GlossFile, results.WC)
	}

	srcFile := collectionEntry.CollectionFile
	i := strings.Index(srcFile, ".")
	if i <= 0 {
		return nil, fmt.Errorf("writeAnalysis: no period in source file name: %s", srcFile)
	}
	basename := srcFile[:i] + "_analysis.html"
	analysisDir := c.ProjectHome + "/" + outputConfig.WebDir + "/analysis/"
	filename := analysisDir + basename
	sf, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("analysis.writeAnalysis: %v", err)
	}
	defer sf.Close()
	err = writeAnalysis(&aResults, collectionEntry.CollectionFile,
		collectionEntry.GlossFile, collectionEntry.Title, "",
		outputConfig, wdict, sf)
	if err != nil {
		return nil, fmt.Errorf("writeCollection could write analysis: %v", err)
	}

	introFN := corpusConfig.CorpusDir + "/" + collectionEntry.Intro
	infile, err := os.Open(introFN)
	if err != nil {
		return nil, fmt.Errorf("writeCollection, could not open intro file %s: %v", collectionEntry.Intro, err)
	}
	defer infile.Close()
	introText := corpus.ReadIntroFile(infile)
	colFN := c.ProjectHome + "/" + outputConfig.WebDir + "/" + collectionEntry.GlossFile
	log.Printf("writeCollection, Writing %s", colFN)
	colF, err := os.Create(colFN)
	if err != nil {
		return nil, fmt.Errorf("error creating collection output file: %v ", err)
	}
	defer colF.Close()
	collectionEntry.CorpusEntries = *corpusEntries
	collectionEntry.AnalysisFile = basename
	collectionEntry.Intro = introText
	err = generator.WriteCollectionFile(collectionEntry, outputConfig,
		corpusConfig, colF)
	if err != nil {
		return nil, fmt.Errorf("error writing collection file: %v ", err)
	}
	return &aResults, nil
}

// WriteCorpus write all the collections in the given corpus
// collections: The set of collections to write to HTML
// baseDir: The base directory to use to write the files
func WriteCorpus(collections []corpus.CollectionEntry,
	outputConfig generator.HTMLOutPutConfig,
	libLoader library.LibraryLoader, dictTokenizer tokenizer.Tokenizer,
	indexConfig index.IndexConfig, dict *dictionary.Dictionary,
	c config.AppConfig, corpusConfig corpus.CorpusConfig, bibClient bibnotes.BibNotesClient) (*index.IndexState, error) {
	log.Printf("analysis.WriteCorpus: enter %d collections", len(collections))
	corpLoader := libLoader.GetCorpusLoader()
	colMap := make(map[string]*[]corpus.CorpusEntry)
	for _, collectionEntry := range collections {
		log.Printf("analysis.WriteCorpus: indexing %s", collectionEntry.CollectionFile)
		cFile := collectionEntry.CollectionFile
		corpusEntries, err := corpLoader.LoadCollection(cFile, collectionEntry.Title)
		if err != nil {
			log.Printf("analysis.WriteCorpus error loading file %s: %v", cFile, err)
			continue
		}
		colMap[collectionEntry.GlossFile] = corpusEntries
	}
	wfDocMap := index.TermFreqDocMap{}
	bigramDocMap := index.TermFreqDocMap{}
	docFreq := index.NewDocumentFrequency() // used to accumulate doc frequencies
	bigramDF := index.NewDocumentFrequency()
	aResults := NewCollectionAResults()
	for _, collectionEntry := range collections {
		results, err := writeCollection(&collectionEntry, outputConfig, libLoader, dictTokenizer, dict, c, bibClient, colMap)
		if err != nil {
			return nil, fmt.Errorf("WriteCorpus could not write collection: %v", err)
		}
		aResults.AddResults(results)
		docFreq.AddDocFreq(results.DocFreq)
		bigramDF.AddDocFreq(results.BigramDF)
		wfDocMap.Merge(results.WFDocMap)
		bigramDocMap.Merge(results.BigramDocMap)
	}

	analysisDir := c.ProjectHome + "/" + outputConfig.WebDir + "/analysis/"
	basename := "corpus_analysis.html"
	analysisFN := analysisDir + basename
	analysisWriter, err := os.Create(analysisFN)
	if err != nil {
		return nil, fmt.Errorf("writeAnalysisCorpus: error creating summary analysis file %v", err)
	}
	defer analysisWriter.Close()
	if err := writeAnalysisCorpus(&aResults, docFreq, outputConfig, indexConfig,
		dict.Wdict, c, analysisWriter); err != nil {
		return nil, fmt.Errorf("WriteCorpus could not write analysis: %v", err)
	}

	docFreqFName := indexConfig.IndexDir + "/" + index.DocFreqFile
	f, err := os.Create(docFreqFName)
	if err != nil {
		return nil, fmt.Errorf("error writing document frequency file %s: %v",
			docFreqFName, err)
	}
	defer f.Close()
	docFreq.Write(f)

	bigramFName := indexConfig.IndexDir + "/" + index.BigramDocFreqFile
	bgF, err := os.Create(docFreqFName)
	if err != nil {
		return nil, fmt.Errorf("error writing bigram frequency file %s: %v", bigramFName, err)
	}
	defer bgF.Close()
	bigramDF.Write(bgF)

	wfDocMap.WriteToFile(docFreq, index.WfDocFile, indexConfig)
	bigramDocMap.WriteToFile(bigramDF, index.BF_DOC_FILE, indexConfig)
	docLenFN := indexConfig.IndexDir + "/" + index.DocLengthFile
	docLenFile, err := os.Create(docLenFN)
	if err != nil {
		return nil, fmt.Errorf("could not open write doc len file: %v", err)
	}
	defer docLenFile.Close()
	index.WriteDocLengthToFile(aResults.DocLengthArray, docLenFile)

	fname := indexConfig.IndexDir + "/" + index.WfCorpusFile
	wfFile, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("analysis.WriteCorpus, error opening word freq file: %f", err)
	}
	defer wfFile.Close()
	wfDocFName := indexConfig.IndexDir + "/" + index.WfDocFile
	wfDocReader, err := os.Open(wfDocFName)
	if err != nil {
		return nil, fmt.Errorf("analysis.WriteCorpus, error opening word freq doc file: %v", err)
	}
	defer wfDocReader.Close()
	indexFN := indexConfig.IndexDir + "/" + index.KeywordIndexFile
	indexWriter, err := os.Create(indexFN)
	if err != nil {
		return nil, fmt.Errorf("index.writeKeywordIndex: Could not create file: %v", err)
	}
	defer indexWriter.Close()
	indexStore := index.IndexStore{
		WfReader:    wfFile,
		WfDocReader: wfDocReader,
		IndexWriter: indexWriter}
	indexState, err := index.BuildIndex(indexConfig, indexStore)
	if err != nil {
		return nil, fmt.Errorf("error building index: %v", err)
	}

	textsFN := c.ProjectHome
	if len(outputConfig.GoStaticDir) > 0 {
		textsFN = textsFN + "/" + outputConfig.GoStaticDir
	}
	textsFN += "/texts.html"
	log.Printf("WriteCorpus, writing list of collections to %s", textsFN)
	collListWriter, err := os.Create(textsFN)
	if err != nil {
		return nil, fmt.Errorf("analysis.WriteCorpus: Could not create texts file, %s: %v",
			textsFN, err)
	}
	defer collListWriter.Close()
	if err := generator.WriteCollectionList(collections, analysisFN, outputConfig, collListWriter); err != nil {
		return nil, fmt.Errorf("analysis.WriteCorpus: Could write texts file, %s: %v", textsFN, err)
	}
	log.Println("analysis.WriteCorpus: exit")
	return indexState, nil
}

// WriteCorpusAll write all the collections in the default corpus
// (collections.csv file)
func WriteCorpusAll(libLoader library.LibraryLoader,
	dictTokenizer tokenizer.Tokenizer, outputConfig generator.HTMLOutPutConfig,
	indexConfig index.IndexConfig, dict *dictionary.Dictionary,
	c config.AppConfig, bibClient bibnotes.BibNotesClient) (*index.IndexState, error) {
	log.Printf("analysis.WriteCorpusAll: enter")
	corpLoader := libLoader.GetCorpusLoader()
	corpusConfig := corpLoader.GetConfig()
	collections, err := corpLoader.LoadCollections()
	if err != nil {
		return nil, fmt.Errorf("writeCorpusAll could not load corpus: %v", err)
	}
	indexState, err := WriteCorpus(*collections, outputConfig, libLoader, dictTokenizer,
		indexConfig, dict, c, corpusConfig, bibClient)
	if err != nil {
		return nil, fmt.Errorf("writeCorpusAll error: %v", err)
	}
	return indexState, nil
}

// WriteCorpusCol writes a corpus document collection to HTML, including all
// the entries contained in the collection
// collectionFile: the name of the collection file
func WriteCorpusCol(collectionFile string, libLoader library.LibraryLoader,
	dictTokenizer tokenizer.Tokenizer, outputConfig generator.HTMLOutPutConfig,
	corpusConfig corpus.CorpusConfig, dict *dictionary.Dictionary,
	c config.AppConfig, bibClient bibnotes.BibNotesClient) error {

	collectionEntry, err := libLoader.GetCorpusLoader().GetCollectionEntry(collectionFile)
	if err != nil {
		return fmt.Errorf("analysis.WriteCorpusCol:  could not get entry %v", err)
	}
	_, err = writeCollection(collectionEntry, outputConfig, libLoader,
		dictTokenizer, dict, c, bibClient, nil)
	if err != nil {
		return fmt.Errorf("analysis.WriteCorpusCol: error writing collection %v", err)
	}
	return nil
}

// Writes dictionary headword entries
// func WriteHwFiles(loader library.LibraryLoader,
//
//	dictTokenizer tokenizer.Tokenizer,
//	outputConfig generator.HTMLOutPutConfig,
//	indexState index.IndexState,
//	wdict map[string]dicttypes.Word,
//	vocabAnalysis VocabAnalysis,
//	hww HeadwordWriter) error {
func WriteHwFiles(dep HWFileDependencies) error {
	// hWFileDependencies := analysis.HWFileDependencies
	log.Printf("analysis.WriteHwFiles: Begin +++++++++++\n")
	outputConfig := dep.OutputConfig
	wdict := dep.Dict.Wdict
	vocabAnalysis := dep.VocabAnalysis
	hwArray := getHeadwords(dep.Dict)
	usageMap := vocabAnalysis.UsageMap
	collocations := vocabAnalysis.Collocations
	outfileMap, err := corpus.GetOutfileMap(dep.Loader.GetCorpusLoader())
	if err != nil {
		return fmt.Errorf("WriteHwFiles, Error getting outfile map: %v", err)
	}
	log.Printf("analysis.WriteHwFiles: outfileMap has %d entries",
		len(*outfileMap))
	dateUpdated := time.Now().Format("2006-01-02")

	// Prepare template
	log.Printf("analysis.WriteHwFiles: Prepare template")
	tmpl, ok := outputConfig.Templates["headword-template.html"]
	if !ok {
		return fmt.Errorf("WriteHwFiles, headword-template.html not found")
	}

	var processor dictionary.NotesProcessor
	if len(outputConfig.NotesReMatch) > 0 {
		log.Printf("analysis.WriteHwFiles: initializing notesProcessor")
		processor = dictionary.NewNotesProcessor(outputConfig.NotesReMatch,
			outputConfig.NotesReplace)
	}

	i := 0
	for _, hw := range hwArray {

		if i%1000 == 0 {
			log.Printf("analysis.WriteHwFiles: wrote %d words", i)
		}

		// Replace text in notes, if configured
		if len(outputConfig.NotesReMatch) > 0 {
			hw = processor.Process(hw)
		}

		// Words that contain this word
		contains := containsWord(hw.Simplified, hwArray)

		// Filter contains words by domain
		cByDomain := containsByDomain(contains, outputConfig)
		contains = Subtract(contains, cByDomain)

		// Sorted array of collocations
		wordCollocations := collocations.SortedCollocations(hw.HeadwordId)

		// Combine usage arrays for both simplified and traditional characters
		usageArrPtr, ok := usageMap[hw.Simplified]
		if !ok {
			usageArrPtr, ok = usageMap[hw.Traditional]
			if !ok {
				//log.Printf("WriteHwFiles: no usage for %s", hw.Simplified)
				usageArrPtr = &[]wordUsage{}
			}
		} else {
			usageArrTradPtr, ok := usageMap[hw.Traditional]
			if ok {
				usageArr := *usageArrPtr
				usageArrTrad := *usageArrTradPtr
				for j := range usageArrTrad {
					usageArr = append(usageArr, usageArrTrad[j])
				}
				usageArrPtr = &usageArr
			}
		}

		// Decorate useage text
		hlUsageArr := []wordUsage{}
		for _, wu := range *usageArrPtr {
			hlText := generator.DecodeUsageExample(wu.Example, hw, dep.DictTokenizer,
				outputConfig, wdict)
			hlWU := wordUsage{
				Freq:       wu.Freq,
				RelFreq:    wu.RelFreq,
				Word:       wu.Word,
				Example:    hlText,
				File:       wu.File,
				EntryTitle: wu.EntryTitle,
				ColTitle:   wu.ColTitle,
			}
			hlUsageArr = append(hlUsageArr, hlWU)
		}

		dictEntry := DictEntry{
			Title:    hw.Simplified,
			Headword: hw,
			RelevantDocs: index.FindDocsForKeyword(hw, *outfileMap, dep.IndexState,
				dep.BibNotesClient),
			ContainsByDomain: cByDomain,
			Contains:         contains,
			Collocations:     wordCollocations,
			UsageArr:         hlUsageArr,
			DateUpdated:      dateUpdated,
		}
		f := dep.Hww.NewWriter(hw.HeadwordId)
		err = writeHwFile(f, dictEntry, *tmpl)
		if err != nil {
			return fmt.Errorf("generator.WriteHwFile: error executing template for "+
				"hw.Id: %d, Simplified: %s, error: %v", hw.HeadwordId,
				hw.Simplified, err)
		}
		dep.Hww.CloseWriter(hw.HeadwordId)
		i++
	}
	return nil
}

// WriteLibraryFile writes a HTML files describing the corpora in the library.
//
// This is for both public and for the translation portal (requiring login).
func WriteLibraryFile(lib library.Library, corpora []library.CorpusData,
	outputFile string, outputConfig generator.HTMLOutPutConfig) {
	log.Printf("analysis.writeLibraryFile: with %d corpora, outputFile = %s, "+
		"TargetStatus = %s", len(corpora), outputFile, lib.TargetStatus)
	libData := library.LibraryData{
		Title:        lib.Title,
		Summary:      lib.Summary,
		DateUpdated:  lib.DateUpdated,
		TargetStatus: lib.TargetStatus,
		Corpora:      corpora,
	}
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal("library.WriteLibraryFile: could not open file", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	templFile := outputConfig.TemplateDir + "/library-template.html"
	tmpl := template.Must(template.New(
		"library-template.html").ParseFiles(templFile))
	err = tmpl.Execute(w, libData)
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

}

// Writes dictionary headword entry to writer
func writeHwFile(f io.Writer, dictEntry DictEntry, tmpl template.Template) error {
	w := bufio.NewWriter(f)
	err := tmpl.Execute(w, dictEntry)
	if err != nil {
		return err
	}
	w.Flush()
	return nil
}
