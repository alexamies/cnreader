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

package index

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// Word frequencies for each document
const WF_DOC_FILE = "word_freq_doc.txt"

// Bigram frequencies for each file
const BF_DOC_FILE = "bigram_freq_doc.txt"

// Remembers the word frequency for each term for each document in the corpus
type TermFreqDocRecord struct {
	Word string
	Freq, DocLength int
	CollectionFile string
	GlossFile string
}

// Remembers the word frequency for each term for each document in the corpus
type TermFreqDocMap map[string]TermFreqDocRecord

// Ads a map of word frequencies for a given document to the map
func (wfDocMap TermFreqDocMap) AddWF(vocab map[string]int, corpusFile,
		glossFile string, wc int) {
	//log.Printf("index.AddWF: enter %d, %d", len(wfDocMap), len(vocab))
	for word, count := range vocab {
		record := TermFreqDocRecord{
			Word:			word, 
			Freq:			count, 
			DocLength:		wc,
			CollectionFile:	corpusFile, 
			GlossFile:		glossFile}
		wfDocMap.Put(record)
	}
	//log.Printf("index.AddWF: exit %d", len(wfDocMap))
}

//Merge two TermFreqDocMap struct's together
func (wfDocMap TermFreqDocMap) Merge(wfDocMap2 TermFreqDocMap) {
	//log.Printf("index.Merge: enter %d, %d", len(wfDocMap), len(wfDocMap2))
	for _, record := range wfDocMap2 {
		wfDocMap.Put(record)
	}
}

// Adds a record to the map
func (wfDocMap TermFreqDocMap) Put(record TermFreqDocRecord) {
	key := record.Word + record.GlossFile
	_, ok := wfDocMap[key]
	if ok {
		log.Printf("index.TermFreqDocRecord: key, %s %s is already in map", 
			record.Word, record.GlossFile)
		return
	}
	wfDocMap[key] = record
}

// Append document analysis to a plain text file in the index directory
func (termFreqDocMap TermFreqDocMap) WriteToFile(df DocumentFrequency,
		fileName string, indexConfig IndexConfig) {
	log.Printf("index.WriteToFile: enter, %s, %d, %d", fileName,
			len(termFreqDocMap), len(df.DocFreq))
	fname := indexConfig.IndexDir + "/" + fileName
	wfFile, err := os.Create(fname)
	if err != nil {
		log.Printf("Could not open write wfFile: %v", err)
		return
	}
	defer wfFile.Close()
	wfWriter := bufio.NewWriter(wfFile)
	i := 0
	for _, record := range termFreqDocMap {
		idf, ok := df.IDF(record.Word)
		if !ok {
			log.Printf("WriteToFile, %s: Could not compute document frequency " +
				"for %s\n", fileName, record.Word)
			idf = 0.0
		}
		if i < 3 {
			log.Printf("WriteToFile, %s: document length %d\n", fileName,
				record.DocLength)
		}
		i++
		fmt.Fprintf(wfWriter, "%s\t%d\t%s\t%s\t%.4f\t%d\n", record.Word, record.Freq,
			record.CollectionFile, record.GlossFile, idf, record.DocLength)
	}
	wfWriter.Flush()
}