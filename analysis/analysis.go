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

// Package for Chinese vocabulary analysis of a corpus
package analysis

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/generator"
	"github.com/alexamies/cnreader/index"
	"github.com/alexamies/cnreader/ngram"
	"github.com/alexamies/cnreader/library"
)

// Maximum number of word frequency entries to output to the generated
// HTML file
const MAX_WF_OUTPUT = 500

// Maximum number of unknwon characters to output to the generated
// HTML file
const MAX_UNKOWN_OUTPUT = 50

// Max usage elements for a word
const MAX_USAGE = 25

// Max number of occurrences of same title in a list of word usages
const MAX_TITLE = 5

// Max number of keywords to display
const MAX_KEYWORDS = 10

// Maximum number of containing words to output to the generated
// HTML file
const MAX_CONTAINS = 50

// Holds vocabulary analysis for a corpus text
type AnalysisResults struct {
	Title                   string
	WC, UniqueWords, CCount int
	ProperNouns				[]dicttypes.Word
	DocumentGlossary 		Glossary
	TopKeywords				dicttypes.Words
	WordFrequencies         []WFResult
	LexicalWordFreq         []WFResult
	BigramFreqSorted        []ngram.BigramFreq
	UnkownnChars            []index.SortedWordItem
	DateUpdated             string
	MaxWFOutput             int
}

// Dictionary entry content struct used for writing a dictionary entry to HTML
type DictEntry struct {
	Headword     dicttypes.Word
	RelevantDocs []index.RetrievalResult
	ContainsByDomain []dicttypes.Word
	Contains     []dicttypes.Word
	Collocations []ngram.BigramFreq
	UsageArr     []WordUsage
	DateUpdated  string
}

// HTML content for template
type HTMLContent struct {
	Content, DateUpdated, Title, FileName string
}

// Bundles up vocabulary analysis
type VocabAnalysis struct {
	UsageMap map[string]*[]WordUsage
	WFTotal map[*index.CorpusWord]index.CorpusWordFreq
	WCTotal map[string]int
	Collocations ngram.CollocationMap
}

// Word usage
type WordUsage struct {
	Freq                                      int
	RelFreq                                   float64
	Word, Example, File, EntryTitle, ColTitle string
}

// Vocabulary analysis entry for a single word
type WFResult struct {
	Freq, HeadwordId                int
	Chinese, Pinyin, English, Usage string
}

// Get a list of words that containst the given word
func ContainsWord(word string, headwords []dicttypes.Word) []dicttypes.Word {
	//log.Printf("dictionary.ContainsWord: Enter\n")
	contains := []dicttypes.Word{}
	for _, hw := range headwords {
		if len(contains) <= MAX_CONTAINS && hw.Simplified != word && strings.Contains(hw.Simplified, word) {
			contains = append(contains, hw)
		}
	}
	return contains
}


// Compute headword numbers for all lexical units listed in data/words.txt
// Return a sorted array of headwords
func GetHeadwords(wdict map[string]dicttypes.Word) []dicttypes.Word {
	hwArray := []dicttypes.Word{}
	for _, w := range wdict {
		hwArray = append(hwArray, w)
	}
	log.Printf("dictionary.GetHeadwords: hwcount = %d\n", len(hwArray))
	return hwArray
}

func GetHwMap(wdict map[string]dicttypes.Word) map[int]dicttypes.Word {
	hwIdMap := make(map[int]dicttypes.Word)
	for _, w := range wdict {
		hwIdMap[w.HeadwordId] = w
	}
	return hwIdMap
}

// Tests whether the symbol is a CJK character, excluding punctuation
// Only looks at the first charater in the string
func IsCJKChar(character string) bool {
	r := []rune(character)
	return unicode.Is(unicode.Han, r[0]) && !unicode.IsPunct(r[0])
}

// Filter the list of headwords by the given domain
// Parameters
//   hws: A list of headwords
//   domain_en: the domain to filter by, ignored if empty
// Return
//   hw: an array of headwords matching the domain, with senses not matching the
//       domain also removed
func FilterByDomain(hws []dicttypes.Word, domain string) []dicttypes.Word {
	if len(domain) == 0 {
		return hws
	}
	headwords := []dicttypes.Word{}
	for _, hw := range hws {
		wsArr := []dicttypes.WordSense{}
		for _, ws := range hw.Senses {
			if ws.Domain == domain {
				wsArr = append(wsArr, ws)
			}
		}
		if len(wsArr) > 0 {
			h := dicttypes.CloneWord(hw)
			h.Senses = wsArr
			headwords = append(headwords, h)
		}
	}
	return headwords
}

// Breaks text into a list of CJK and non CJK strings
func GetChunks(text string) list.List {
	var chunks list.List
	cjk := ""
	noncjk := ""
	for _, character := range text {
		if IsCJKChar(string(character)) {
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

// Compute word frequencies, collocations, and usage for the entire corpus
func GetWordFrequencies(libLoader library.LibraryLoader,
		dictTokenizer tokenizer.Tokenizer,
		corpusConfig corpus.CorpusConfig,
		wdict map[string]dicttypes.Word) VocabAnalysis {

	log.Printf("analysis.GetWordFrequencies: enter")

	// Overall word frequencies per corpus
	collocations := ngram.CollocationMap{}
	usageMap := map[string]*[]WordUsage{}
	ccount := 0 // character count
	wcTotal := map[string]int{}
	wfTotal := map[*index.CorpusWord]index.CorpusWordFreq{}

	corpLoader := libLoader.GetCorpusLoader()
	collectionEntries := corpLoader.LoadCorpus(corpus.COLLECTIONS_FILE)
	for _, col := range collectionEntries {
		colFile := col.CollectionFile
		//log.Printf("GetWordFrequencies: input file: %s\n", colFile)
		corpusEntries := corpLoader.LoadCollection(colFile, col.Title)
		for _, entry := range corpusEntries {
			src := corpusConfig.CorpusDir + entry.RawFile
			text := corpLoader.ReadText(src)
			ccount += utf8.RuneCountInString(text)
			_, results := ParseText(text, col.Title, &entry, dictTokenizer,
						corpusConfig, wdict)
			wcTotal[col.Corpus] += results.WC

			// Process collocations
			collocations.MergeCollocationMap(results.Collocations)

			// Find word usage for the given word
			for word, count := range results.Vocab {
				cw := &index.CorpusWord{col.Corpus, word}
				cwf := &index.CorpusWordFreq{col.Corpus, word, count}
				if cwfPrev, found := wfTotal[cw]; found {
					cwf.Freq += cwfPrev.Freq
				}
				wfTotal[cw] = *cwf
				rel_freq := 1000.0 * float64(count) / float64(results.WC)
				usage := WordUsage{cwf.Freq, rel_freq, word, results.Usage[word],
					entry.GlossFile, entry.Title, col.Title}
				usageArr, ok := usageMap[word]
				if !ok {
					usageArr = new([]WordUsage)
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

	return VocabAnalysis{usageMap, wfTotal, wcTotal, collocations}
}

// Parses a Chinese text into words
// Parameters:
//   text: the string to parse
//   ColTitle: Optional parameter used for tracing collocation usage
//   document: Optional parameter used for tracing collocation usage
// Returns:
//   tokens: the tokens for the parsed text
//   results: vocabulary analysis results
func ParseText(text string, colTitle string, document *corpus.CorpusEntry,
		dictTokenizer tokenizer.Tokenizer, corpusConfig corpus.CorpusConfig,
		wdict map[string]dicttypes.Word) (list.List, *CollectionAResults) {
	tokens := list.List{}
	vocab := map[string]int{}
	bigrams := map[string]int{}
	bigramMap := ngram.BigramFreqMap{}
	collocations := ngram.CollocationMap{}
	unknownChars := map[string]int{}
	usage := map[string]string{}
	wc := 0
	cc := 0
	chunks := GetChunks(text)
	hwIdMap := GetHwMap(wdict)
	lastHWPtr := &dicttypes.Word{}
	lastHW := *lastHWPtr
	lastHWText := ""
	//fmt.Printf("ParseText: For text %s got %d chunks\n", colTitle, chunks.Len())
	for e := chunks.Front(); e != nil; e = e.Next() {
		chunk := e.Value.(string)
		//fmt.Printf("ParseText: chunk %s\n", chunk)
		characters := strings.Split(chunk, "")
		if !IsCJKChar(characters[0]) || corpus.IsExcluded(corpusConfig.Excluded, chunk) {
			tokens.PushBack(chunk)
			lastHWPtr = &dicttypes.Word{}
			lastHW = *lastHWPtr
			lastHWText = ""
			continue
		}
		textTokens := dictTokenizer.Tokenize(chunk)
		//fmt.Printf("ParseText: len(tokens) %d\n", len(tokens))
		for _, token := range textTokens {
			w := token.Token
			if !corpus.IsExcluded(corpusConfig.Excluded , w) {
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
				if lastHW.HeadwordId != 0 {
					if len(hw.Senses) == 0 {
						log.Printf("ParseText: WordSenses nil for %s "+
									", id = %d, in %s, %s\n", w, hwid,
									document.Title, colTitle)
					}
					bigram, ok := bigramMap.GetBigramVal(lastHW.HeadwordId, hwid)
					if !ok {
						bigram = ngram.NewBigram(lastHW, hw, chunk,
								document.GlossFile, document.Title, colTitle)
					}
					bigramMap.PutBigram(bigram)
					collocations.PutBigram(bigram.HeadwordDef1.HeadwordId, bigram)
					collocations.PutBigram(bigram.HeadwordDef2.HeadwordId, bigram)
				}
				lastHW = hw
			}
		}
	}
	dl := index.DocLength{document.GlossFile, wc}
	dlArray := []index.DocLength{dl}
	results := CollectionAResults{
		Vocab:				vocab,
		Bigrams:			bigrams,
		Usage:				usage,
		BigramFrequencies:	bigramMap,
		Collocations:		collocations,
		WC:					wc,
		CCount:				cc,
		UnknownChars:		unknownChars,
		DocLengthArray:		dlArray,
	}
	return tokens, &results
}

// Sample word usage for usability making sure that the list of word usage
// samples is not dominated by any one title and truncating at MAX_USAGE
// examples.
func sampleUsage(usageMap map[string]*[]WordUsage) map[string]*[]WordUsage {
	for word, usagePtr := range usageMap {
		sampleMap := map[string]int{}
		usage := *usagePtr
		usageCapped := new([]WordUsage)
		j := 0
		for _, wu := range usage {
			count, _ := sampleMap[wu.ColTitle]
			if count < MAX_TITLE && j < MAX_USAGE {
				*usageCapped = append(*usageCapped, wu)
				sampleMap[wu.ColTitle]++
				j++
			}
		}
		usageMap[word] = usageCapped
	}
	return usageMap
}

// For the HTML template
func add(x, y int) int {
	return x + y
}

// Writes out an analysis of the entire corpus, including word frequencies
// and other data. The output file is called 'corpus-analysis.html' in the
// web/analysis directory.
// Parameters:
//   results: The results of corpus analysis
//   docFreq: document frequency for terms
// Returns: the name of the file written to
func writeAnalysisCorpus(results *CollectionAResults,
		docFreq index.DocumentFrequency, outputConfig generator.HTMLOutPutConfig,
		indexConfig index.IndexConfig, wdict map[string]dicttypes.Word) string {

	// If the web/analysis directory does not exist, then skip the analysis
	analysisDir := outputConfig.WebDir + "/analysis/"
	_, err := os.Stat(analysisDir)
	if err != nil {
		return ""
	}

	// Parse template and organize template parameters
	sortedWords := index.SortedFreq(results.Vocab)
	wfResults := results.GetWordFreq(sortedWords, wdict)
	maxWf := len(wfResults)
	if maxWf > MAX_WF_OUTPUT {
		maxWf = MAX_WF_OUTPUT
	}

	lexicalWordFreq := results.GetLexicalWordFreq(sortedWords, wdict)
	maxLex := len(lexicalWordFreq)
	if maxLex > MAX_WF_OUTPUT {
		maxLex = MAX_WF_OUTPUT
	}

	sortedUnknownWords := index.SortedFreq(results.UnknownChars)
	maxUnknown := len(sortedUnknownWords)
	if maxUnknown > MAX_UNKOWN_OUTPUT {
		maxUnknown = MAX_UNKOWN_OUTPUT
	}

	// Bigrams, also truncated
	bFreq := ngram.SortedFreq(results.BigramFrequencies)
	maxBFOutput := len(bFreq)
	if maxBFOutput > MAX_WF_OUTPUT {
		maxBFOutput = MAX_WF_OUTPUT
	}

	dateUpdated := time.Now().Format("2006-01-02")
	title := "Terminology Extraction and Vocabulary Analysis"
	aResults := AnalysisResults{
		Title:            title,
		WC:               results.WC,
		CCount:			  		results.CCount,
		ProperNouns:      dicttypes.Words{},
		DocumentGlossary: MakeGlossary("", []dicttypes.Word{}),
		TopKeywords:	  	dicttypes.Words{},
		UniqueWords:      len(results.Vocab),
		WordFrequencies:  wfResults[:maxWf],
		LexicalWordFreq:  lexicalWordFreq[:maxLex],
		BigramFreqSorted: bFreq[:maxBFOutput],
		UnkownnChars:     sortedUnknownWords[:maxUnknown],
		DateUpdated:      dateUpdated,
		MaxWFOutput:      len(wfResults),
	}
	tmplFile := outputConfig.TemplateDir + "/corpus-summary-analysis-template.html"
	funcs := template.FuncMap{
		"add": add,
		"Deref":   func(sp *string) string { return *sp },
		"DerefNe": func(sp *string, s string) bool { return *sp != s },
	}
	tmpl, err := template.New("corpus-summary-analysis-template.html").Funcs(funcs).ParseFiles(tmplFile)
	if (err != nil || tmpl == nil) {
		log.Printf("writeAnalysisCorpus: Error getting template %v)", tmplFile)
		return ""
	}
	basename := "corpus_analysis.html"
	filename := analysisDir + basename
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("writeAnalysisCorpus: error creating file %v", err)
		return ""
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	err = tmpl.Execute(w, aResults)
	if err != nil {
		log.Printf("writeAnalysisCorpus: error executing template%v", err)
	}
	w.Flush()
	log.Printf("writeAnalysisCorpus: finished executing template%v", err)

	// Write results to plain text files
	index.WriteWFCorpus(sortedWords, sortedUnknownWords, bFreq, results.WC, indexConfig)

	return basename
}

// Writes a document with vocabulary analysis of the text. The name of the
// output file will be source file with '-analysis' appended, placed in the
// web/analysis directory
// results: The results of vocabulary analysis
// collectionTitle: The title of the whole colleciton
// docTitle: The title of this specific document
// Returns the name of the file written to
func writeAnalysis(results *CollectionAResults, srcFile, glossFile,
		collectionTitle, docTitle string,
		outputConfig generator.HTMLOutPutConfig,
		wdict map[string]dicttypes.Word) string {

	//log.Printf("analysis.writeAnalysis: enter")
	analysisDir := outputConfig.WebDir + "/analysis/"
	_, err := os.Stat(analysisDir)
	if err != nil {
		return ""
	}

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
	if maxWf > MAX_WF_OUTPUT {
		maxWf = MAX_WF_OUTPUT
	}

	lexicalWordFreq := results.GetLexicalWordFreq(sortedWords, wdict)
	maxLex := len(lexicalWordFreq)
	if maxLex > MAX_WF_OUTPUT {
		maxLex = MAX_WF_OUTPUT
	}

	topKeywords := []dicttypes.Word{}
	if domain_label != "" {
		//keywords := index.SortByWeight(results.Vocab)
		//topKeywords = index.GetHeadwordArray(keywords)
		//topKeywords = dictionary.FilterByDomain(topKeywords, domain_label)
		topKeywords = index.FilterByDomain(sortedWords, domain_label, wdict)
		maxKeywords := len(topKeywords)
		if maxKeywords > MAX_KEYWORDS {
			maxKeywords = MAX_KEYWORDS
		}
		topKeywords = topKeywords[:maxKeywords]
	}

	//log.Printf("analysis.writeAnalysis: title: %s, len topKeywords: %d, " +
	//	"domain_label: %s\n", docTitle, len(topKeywords), domain_label)

	sortedUnknownWords := index.SortedFreq(results.UnknownChars)
	maxUnknown := len(sortedUnknownWords)
	if maxUnknown > MAX_UNKOWN_OUTPUT {
		maxUnknown = MAX_UNKOWN_OUTPUT
	}

	// Bigrams, also truncated
	bFreq := ngram.SortedFreq(results.BigramFrequencies)
	maxBFOutput := len(bFreq)
	if maxBFOutput > MAX_WF_OUTPUT {
		maxBFOutput = MAX_WF_OUTPUT
	}

	dateUpdated := time.Now().Format("2006-01-02")
	title := "Glossary and Vocabulary for " + collectionTitle
	if docTitle != "" {
		title += ", " + docTitle
	}

	aResults := AnalysisResults{
		Title:            title,
		WC:               results.WC,
		CCount:			  results.CCount,
		ProperNouns:      properNouns,
		DocumentGlossary: glossary,
		TopKeywords:	  topKeywords,
		UniqueWords:      len(results.Vocab),
		WordFrequencies:  wfResults[:maxWf],
		LexicalWordFreq:  lexicalWordFreq[:maxLex],
		BigramFreqSorted: bFreq[:maxBFOutput],
		UnkownnChars:     sortedUnknownWords[:maxUnknown],
		DateUpdated:      dateUpdated,
		MaxWFOutput:      len(wfResults),
	}
	tmplFile := outputConfig.TemplateDir + "/corpus-analysis-template.html"
	funcs := template.FuncMap{
		"add": add,
		"Deref":   func(sp *string) string { return *sp },
		"DerefNe": func(sp *string, s string) bool { 
			if sp != nil {
				return *sp != s 
			}
			return false
		},
	}
	tmpl, err := template.New("corpus-analysis-template.html").Funcs(funcs).ParseFiles(tmplFile)
	if (err != nil ||  tmpl == nil) {
		log.Printf("analysiswriteAnalysis: Skipping document analysis (%v)\n",
			tmplFile)
		return ""
	}

	// Write output
	i := strings.Index(srcFile, ".txt")
	if i <= 0 {
		i = strings.Index(srcFile, ".html")
		if i <= 0 {
			i = strings.Index(srcFile, ".csv")
			if i <= 0 {
				log.Fatal("writeAnalysis: Bad name for source file: ", srcFile)
			}
		}
	}
	basename := srcFile[:i] + "_analysis.html"
	filename := analysisDir + basename
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("analysis.writeAnalysis", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	err = tmpl.Execute(w, aResults)
	if err != nil {
		panic(err)
	}
	w.Flush()

	return basename
}

// Writes a corpus document collection to HTML, including all the entries
// contained in the collection
// collectionEntry: the CollectionEntry struct
// baseDir: The base directory to use
func writeCollection(collectionEntry corpus.CollectionEntry,
		outputConfig generator.HTMLOutPutConfig, libLoader library.LibraryLoader,
		dictTokenizer tokenizer.Tokenizer, corpusConfig corpus.CorpusConfig,
		wdict map[string]dicttypes.Word) (*CollectionAResults, error) {

	log.Printf("analysis.writeCollection: enter CollectionFile =" +
			collectionEntry.CollectionFile)
	corpLoader := libLoader.GetCorpusLoader()
	corpusEntries := corpLoader.LoadCollection(collectionEntry.CollectionFile,
		collectionEntry.Title)
	aResults := NewCollectionAResults()
	for _, entry := range corpusEntries {
		//log.Printf("analysis.writeCollection: entry.RawFile = " + entry.RawFile)
		src := corpusConfig.CorpusDir + "/" + entry.RawFile
		dest := outputConfig.WebDir + "/" + entry.GlossFile
		//if collectionEntry.Title == "" {
		//	log.Printf("analysis.writeCollection: collectionEntry.Title is " +
		//		"empty, input file: %s, output file: %s\n", src, dest)
		//}
		text := corpLoader.ReadText(src)
		_, results := ParseText(text, collectionEntry.Title, &entry,
				dictTokenizer, corpusConfig, wdict)
		aFile := writeAnalysis(results, entry.RawFile, entry.GlossFile,
			collectionEntry.Title, entry.Title, outputConfig, wdict)
		sourceFormat := "TEXT"
		if strings.HasSuffix(entry.RawFile, ".html") {
			sourceFormat = "HTML"
		}
		f, err := os.Create(dest)
		if err != nil {
			return nil, fmt.Errorf("writeCollection could not open file: %v", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		defer w.Flush()

		textTokens := dictTokenizer.Tokenize(text)
		err = generator.WriteCorpusDoc(textTokens, results.Vocab, w, collectionEntry.GlossFile,
				collectionEntry.Title, entry.Title, aFile, sourceFormat, outputConfig,
				corpusConfig, wdict)
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
	aFile := writeAnalysis(&aResults, collectionEntry.CollectionFile,
		collectionEntry.GlossFile, collectionEntry.Title, "", outputConfig, wdict)
	introText := corpus.ReadIntroFile(collectionEntry.Intro, corpusConfig)
	err := generator.WriteCollectionFile(collectionEntry, aFile, outputConfig,
			corpusConfig, corpusEntries, introText)
	if err != nil {
		return nil, fmt.Errorf("Error writing collection file: %v ", err)
	}
	return &aResults, nil
}

// Write all the collections in the given corpus
// collections: The set of collections to write to HTML
// baseDir: The base directory to use to write the files
func WriteCorpus(collections []corpus.CollectionEntry,
		outputConfig generator.HTMLOutPutConfig,
		libLoader library.LibraryLoader, dictTokenizer tokenizer.Tokenizer,
		corpusConfig corpus.CorpusConfig, indexConfig index.IndexConfig,
		wdict map[string]dicttypes.Word) error {
	log.Printf("analysis.WriteCorpus: enter")
	index.Reset(indexConfig)
	wfDocMap := index.TermFreqDocMap{}
	bigramDocMap := index.TermFreqDocMap{}
	docFreq := index.NewDocumentFrequency() // used to accumulate doc frequencies
	bigramDF := index.NewDocumentFrequency()
	aResults := NewCollectionAResults()
	for _, collectionEntry := range collections {
		results, err := writeCollection(collectionEntry, outputConfig, libLoader,
				dictTokenizer, corpusConfig, wdict)
		if err != nil {
			return fmt.Errorf("WriteCorpus could not open file: %v", err)
		}
		aResults.AddResults(results)
		docFreq.AddDocFreq(results.DocFreq)
		bigramDF.AddDocFreq(results.BigramDF)
		wfDocMap.Merge(results.WFDocMap)
		bigramDocMap.Merge(results.BigramDocMap)
	}
	writeAnalysisCorpus(&aResults, docFreq, outputConfig, indexConfig, wdict)
	docFreq.WriteToFile(index.DOC_FREQ_FILE, indexConfig)
	bigramDF.WriteToFile(index.BIGRAM_DOC_FREQ_FILE, indexConfig)
	wfDocMap.WriteToFile(docFreq, index.WF_DOC_FILE, indexConfig)
	bigramDocMap.WriteToFile(bigramDF, index.BF_DOC_FILE, indexConfig)
	index.WriteDocLengthToFile(aResults.DocLengthArray, index.DOC_LENGTH_FILE, indexConfig)
	index.BuildIndex(indexConfig)
	log.Printf("analysis.WriteCorpus: exit")
	return nil
}

// Write all the collections in the default corpus (collections.csv file)
func WriteCorpusAll(libLoader library.LibraryLoader,
		dictTokenizer tokenizer.Tokenizer, outputConfig generator.HTMLOutPutConfig,
		corpusConfig corpus.CorpusConfig, indexConfig index.IndexConfig,
		wdict map[string]dicttypes.Word) error {
	log.Printf("analysis.WriteCorpusAll: enter")
	corpLoader := libLoader.GetCorpusLoader()
	collections := corpLoader.LoadCorpus(corpus.COLLECTIONS_FILE)
	err := WriteCorpus(collections, outputConfig, libLoader, dictTokenizer,
			corpusConfig, indexConfig, wdict)
	if err != nil {
		return fmt.Errorf("WriteCorpusAll could not open file: %v", err)
	}
	return nil
}

// Writes a corpus document collection to HTML, including all the entries
// contained in the collection
// collectionFile: the name of the collection file
func WriteCorpusCol(collectionFile string, libLoader library.LibraryLoader,
			dictTokenizer tokenizer.Tokenizer, outputConfig generator.HTMLOutPutConfig,
			corpusConfig corpus.CorpusConfig, wdict map[string]dicttypes.Word) {
	collectionEntry, err := libLoader.GetCorpusLoader().GetCollectionEntry(collectionFile)
	if err != nil {
		log.Fatalf("analysis.WriteCorpusCol: fatal error %v", err)
	}
	writeCollection(collectionEntry, outputConfig, libLoader, dictTokenizer,
			corpusConfig, wdict)
}

// Writes a document with markup for the array of tokens
// tokens: A list of tokens forming the document
// vocab: A list of word id's in the document
// filename: The file name to write to
// GlossChinese: whether to convert the Chinese text in the file to hyperlinks
func WriteDoc(tokens list.List, vocab map[string]int, filename,
	templateName, templateFile string, glossChinese bool, title string,
		corpusConfig corpus.CorpusConfig, wdict map[string]dicttypes.Word) {
	if templateFile != `\N` {
		writeHTMLDoc(tokens, vocab, filename, templateName, templateFile,
			glossChinese, title, wdict)
		return
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	// Iterate over text chunks
	for e := tokens.Front(); e != nil; e = e.Next() {
		chunk := e.Value.(string)
		//fmt.Printf("WriteDoc: Word %s\n", word)
		if word, ok := wdict[chunk]; ok && !corpus.IsExcluded(corpusConfig.Excluded, chunk) {
			wordIds := ""
			for _, ws := range word.Senses {
				if wordIds == "" {
					wordIds = fmt.Sprintf("%d", ws.Id)
				} else {
					wordIds = fmt.Sprintf("%s,%d", wordIds, ws.Id)
				}
			}
			fmt.Fprintf(w, "<span title='%s' data-wordid='%s'"+
				" class='dict-entry' data-toggle='popover'>%s</span>",
				chunk, wordIds, chunk)
		} else {
			w.WriteString(chunk)
		}
	}
	w.Flush()
}

// Writes a document with markup for the array of tokens
// tokens: A list of tokens forming the document
// vocab: A list of word id's in the document
// filename: The file name to write to
// GlossChinese: whether to convert the Chinese text in the file to hyperlinks
func writeHTMLDoc(tokens list.List, vocab map[string]int, filename,
		templateName, templateFile string, glossChinese bool, title string,
		wdict map[string]dicttypes.Word) {
	var b bytes.Buffer

	// Iterate over text chunks
	for e := tokens.Front(); e != nil; e = e.Next() {
		chunk := e.Value.(string)
		//fmt.Printf("WriteDoc: Word %s\n", word)
		if !glossChinese {
			fmt.Fprintf(&b, chunk)
		} else if word, ok := wdict[chunk]; ok {
			// Regular HTML link
			english := ""
			if len(word.Senses) > 0 {
				english = word.Senses[0].English
			}
			mouseover := fmt.Sprintf("%s | %s", word.Pinyin, english)
			link := fmt.Sprintf("/words/%d.html", word.HeadwordId)
			fmt.Fprintf(&b, "<a title='%s' href='%s'>%s</a>", mouseover, link,
				chunk)
		} else {
			fmt.Fprintf(&b, chunk)
		}
	}
	dateUpdated := time.Now().Format("2006-01-02")
	fnameParts := strings.Split(filename, "/")
	fname := fnameParts[len(fnameParts) - 1]
	content := HTMLContent{b.String(), dateUpdated, title, fname}

	// Prepare template
	tmpl := template.Must(template.New(templateName).ParseFiles(templateFile))
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	w := bufio.NewWriter(f)
	err = tmpl.Execute(w, content)
	if err != nil {
		panic(err)
	}
	w.Flush()
	f.Close()

}

// Writes dictionary headword entries
func WriteHwFiles(loader library.LibraryLoader,
		dictTokenizer tokenizer.Tokenizer, outputConfig generator.HTMLOutPutConfig,
		corpusConfig corpus.CorpusConfig, indexConfig index.IndexConfig,
		wdict map[string]dicttypes.Word) {
	log.Printf("analysis.WriteHwFiles: Begin +++++++++++\n")
	index.BuildIndex(indexConfig)
	log.Printf("analysis.WriteHwFiles: Get headwords\n")
	hwArray := GetHeadwords(wdict)
	vocabAnalysis := GetWordFrequencies(loader, dictTokenizer, corpusConfig, wdict)
	usageMap := vocabAnalysis.UsageMap
	collocations := vocabAnalysis.Collocations
	corpusEntryMap := loader.GetCorpusLoader().LoadAll(corpus.COLLECTIONS_FILE)
	outfileMap := corpus.GetOutfileMap(corpusEntryMap)
	dateUpdated := time.Now().Format("2006-01-02")

	// Prepare template
	log.Printf("analysis.WriteHwFiles: Prepare template\n")
	templFile := outputConfig.TemplateDir + "/headword-template.html"
	fm := template.FuncMap{
		"Deref":   func(sp *string) string { return *sp },
		"DerefNe": func(sp *string, s string) bool { return *sp != s },
	}
	tmpl := template.Must(template.New("headword-template.html").Funcs(fm).ParseFiles(templFile))

	i := 0
	for _, hw := range hwArray {

		if i%1000 == 0 {
			log.Printf("analysis.WriteHwFiles: wrote %d words\n", i)
		}

		// Look for different writings of traditional form
		tradVariants := []dicttypes.WordSense{}
		for _, ws := range hw.Senses {
			if hw.Traditional != ws.Traditional {
				tradVariants = append(tradVariants, ws)
			}
		}

		// Words that contain this word
		contains := ContainsWord(hw.Simplified, hwArray)

		// Filter contains words by domain
		containsByDomain := ContainsByDomain(contains, outputConfig)
		contains = Subtract(contains, containsByDomain)

		// Sorted array of collocations
		wordCollocations := collocations.SortedCollocations(hw.HeadwordId)

		// Combine usage arrays for both simplified and traditional characters
		usageArrPtr, ok := usageMap[hw.Simplified]
		if !ok {
			usageArrPtr, ok = usageMap[hw.Traditional]
			if !ok {
				//log.Printf("WriteHwFiles: no usage for %s", hw.Simplified)
				usageArrPtr = &[]WordUsage{}
			}
		} else {
			usageArrTradPtr, ok := usageMap[hw.Traditional]
			if ok {
				usageArr := *usageArrPtr
				usageArrTrad := *usageArrTradPtr
				for j, _ := range usageArrTrad {
					usageArr = append(usageArr, usageArrTrad[j])
				}
				usageArrPtr = &usageArr
			}
		}

		// Decorate useage text
		hlUsageArr := []WordUsage{}
		for _, wu := range *usageArrPtr {
			hlText := generator.DecodeUsageExample(wu.Example, hw, dictTokenizer, outputConfig, wdict)
			hlWU := WordUsage{
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

		dictEntry := DictEntry {
			Headword:     hw,
			RelevantDocs: index.FindDocsForKeyword(hw, outfileMap),
			ContainsByDomain: containsByDomain,
			Contains:     contains,
			Collocations: wordCollocations,
			UsageArr:     hlUsageArr,
			DateUpdated:  dateUpdated,
		}
		filename := fmt.Sprintf("%s%s%d%s", outputConfig.WebDir, "/words/",
				hw.HeadwordId, ".html")
		f, err := os.Create(filename)
		if err != nil {
			log.Printf("WriteHwFiles: Error creating file for hw.Id %d, "+
				"Simplified %s", hw.HeadwordId, hw.Simplified)
			log.Fatalf("hit a problem: %v", err)
		}
		w := bufio.NewWriter(f)
		err = tmpl.Execute(w, dictEntry)
		if err != nil {
			log.Printf("analysis.WriteHwFiles: error executing template for hw.Id: %d,"+
				" filename: %s, Simplified: %s", hw.HeadwordId, filename, hw.Simplified)
			log.Fatalf("hit a different problem: %v", err)
		}
		w.Flush()
		f.Close()
		i++
	}
}

// Writes a HTML files describing the corpora in the library, both public and
// for the translation portal (requiring login)
func writeLibraryFile(lib library.Library, corpora []library.CorpusData,
		outputFile string, outputConfig generator.HTMLOutPutConfig) {
	log.Printf("analysis.writeLibraryFile: with %d corpora, outputFile = %s, " +
			"TargetStatus = %s", len(corpora), outputFile, lib.TargetStatus)
	libData := library.LibraryData{
		Title: lib.Title,
		Summary: lib.Summary,
		DateUpdated: lib.DateUpdated,
		TargetStatus: lib.TargetStatus,
		Corpora: corpora,
	}
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal("library.WriteLibraryFile: could not open file", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	templFile := outputConfig.TemplateDir + "/library-template.html"
	tmpl:= template.Must(template.New(
					"library-template.html").ParseFiles(templFile))
	err = tmpl.Execute(w, libData)
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

}

// Writes a HTML file describing the corpora in the library and for each corpus
// in the library
func WriteLibraryFiles(lib library.Library, dictTokenizer tokenizer.Tokenizer,
			outputConfig generator.HTMLOutPutConfig, corpusConfig corpus.CorpusConfig,
			indexConfig index.IndexConfig, wdict map[string]dicttypes.Word) {
	corpora := lib.Loader.LoadLibrary()
	libraryOutFile := outputConfig.WebDir + "/library.html"
	writeLibraryFile(lib, corpora, libraryOutFile, outputConfig)
	portalDir := ""
	goStaticDir := outputConfig.GoStaticDir
	if len(goStaticDir) != 0 {
		portalDir = corpusConfig.ProjectHome + "/" + goStaticDir
		_, err := os.Stat(portalDir)
		lib.TargetStatus = "translator_portal"
		if err == nil {
			portalLibraryFile := portalDir + "/portal_library.html"
			writeLibraryFile(lib, corpora, portalLibraryFile, outputConfig)
		}
	}
	for _, c := range corpora {
		outputFile := ""
		if c.Status == "public" {
			outputFile = fmt.Sprintf("%s/%s.html", outputConfig.WebDir,
					c.ShortName)
		} else if c.Status == "translator_portal" {
			outputFile = fmt.Sprintf("%s/%s.html", portalDir, c.ShortName)
		} else {
			log.Printf("library.WriteLibraryFiles: not sure what to do with status %v",
				c.Status)
			continue
		}
		fName := fmt.Sprintf(c.FileName)
		collections := lib.Loader.GetCorpusLoader().LoadCorpus(fName)
		WriteCorpus(collections, outputConfig, lib.Loader, dictTokenizer,
				corpusConfig, indexConfig, wdict)
		corpus := library.Corpus{c.Title, "", lib.DateUpdated, collections}
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatal("library.WriteLibraryFiles: could not open file", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		templFile := outputConfig.TemplateDir + "/corpus-list-template.html"
		tmpl:= template.Must(template.New(
					"corpus-list-template.html").ParseFiles(templFile))
		err = tmpl.Execute(w, corpus)
		if err != nil {
			log.Fatal(err)
		}
		w.Flush()
	}
}
