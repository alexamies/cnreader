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
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
)

// File name for document index
const DocFreqFile = "doc_freq.txt"
const BigramDocFreqFile = "bigram_doc_freq.txt"

// Map from term to number of documents referencing the term
type DocumentFrequency struct {
	DocFreq map[string]int
	N       int // total number of documents
}

// Initializes a DocumentFrequency struct
func NewDocumentFrequency() DocumentFrequency {
	return DocumentFrequency{
		DocFreq: map[string]int{},
		N: 0,
	}
}

// Adds the given vocabulary to the map and increments the document count
// Param:
//   vocab - word frequencies are ignored, only the presence of the term is 
//           important
func (df *DocumentFrequency) AddVocabulary(vocab map[string]int) {
	for k, _ := range vocab {
		_, ok := df.DocFreq[k]
		if ok {
			df.DocFreq[k]++
		} else {
			df.DocFreq[k] = 1
		}
	}
	df.N += 1
}

// Merges the given document frequency to the map and increments the counts
// Param:
//   vocab - word frequencies are ignored, only the presence of the term is 
//           important
func (df *DocumentFrequency) AddDocFreq(otherDF DocumentFrequency) {
	for k, v := range otherDF.DocFreq {
		count, ok := df.DocFreq[k]
		if ok {
			df.DocFreq[k] = count + v
		} else {
			df.DocFreq[k] = 1
		}
	}
	df.N += otherDF.N
}

// Computes the inverse document frequency for the given term
// Param:
//   term: the term to find the idf for
func (df *DocumentFrequency) IDF(term string) (val float64, ok bool) {
	ndocs, ok := df.DocFreq[term]
	if ok && ndocs > 0 {
		val = math.Log10(float64(df.N + 1) / float64(ndocs))
	//log.Println("index.IDF: term, val, df.n, ", term, val, df.N)
	} 
	return val, ok
}

// ReadDocumentFrequency a document frequency object from a CSV file
func ReadDocumentFrequency(r io.Reader) (*DocumentFrequency, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadDocumentFrequency: Could not wf file: %v", err)
	}
	dfMap := map[string]int{}
	for i, row := range rawCSVdata {
		w := row[0] // Chinese text for word
		count, err := strconv.ParseInt(row[1], 10, 0)
		if err != nil {
			return nil, fmt.Errorf("ReadDocumentFrequency: could not parse word count %d: %v",
					i, err)
		}
		dfMap[w] = int(count)
	}
	df := DocumentFrequency{dfMap, -1}
	return &df, nil
}

// term frequency - inverse document frequency for the string
// Params
//   term: The term (word) to compute the tf-idf from
//   count: The count of the word in a specific document
func tfIdf(term string, count int, completeDF DocumentFrequency) (val float64, ok bool) {
	idf, ok := completeDF.IDF(term)
	//log.Println("index.tfIdf: idf, term, ", idf, term)
	if ok {
		val = float64(count) * idf
	} else {
		//log.Println("index.tfIdf: could not compute tf-idf for, ", term)
	}
	return val, ok
}

// WriteToFile writes the document frequency
func (df *DocumentFrequency) Write(w io.Writer) {
	writer := bufio.NewWriter(w)
	for k, v := range df.DocFreq {
		fmt.Fprintf(writer, "%s\t%d\n", k, v)
	}
	writer.Flush()
}