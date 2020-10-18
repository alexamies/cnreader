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
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/alexamies/cnreader/ngram"
)

// File name for keyword index
const KeywordIndexFile = "keyword_index.json"

// Maximum number of keywords in index
const maxFilePerKeyword = 50

// Unknown characters file
const UnknownCharsFile = "unknown.txt"

// Word frequencies for corpus
const WfCorpusFile = "word_frequencies.txt"

// ngram frequencies for corpus
const NgramCorpusFile = "ngram_frequencies.txt"

// A word frequency entry record
type WFEntry struct {
	Chinese string
	Count   int
}

// A word frequency entry record
type IndexState struct {
	wf *map[string]WFEntry
	wfdoc *map[string][]WFDocEntry
	KeywordIndexReady bool
}

// Storage for the keyword index
type IndexStore struct {
	WfReader io.Reader
	WfDocReader io.Reader
	IndexWriter io.Writer
}

// Storage for word frequency data
type WordFreqStore struct {
	WFWriter io.Writer
	UnknownCharsWriter io.Writer
	BigramWriter io.Writer
}

// Reads word frequencies data from files into memory and builds the keyword
// index
func BuildIndex(indexConfig IndexConfig, indexStore IndexStore) (*IndexState, error) {
	indexState := IndexState{}
	var err error
	indexState.wf, err = readWFCorpus(indexStore.WfReader)
	if err != nil {
		return nil, fmt.Errorf("index.BuildIndex, error reading word freq corpus: %f",
				err)
	}
	indexState.wfdoc, err = readWFDoc(indexStore.WfDocReader)
	if err != nil {
		return nil, fmt.Errorf("index.BuildIndex, error reading word freq doc: %f",
				err)
	}
	err = writeKeywordIndex(indexState, indexStore.IndexWriter)
	if err != nil {
		return nil, fmt.Errorf("index.BuildIndex, error writing index: %f",
				err)
	}
	indexState.KeywordIndexReady = true
	return &indexState, nil
}

// readWFCorpus reads corpus-wide word frequencies from file into memory
func readWFCorpus(r io.Reader) (*map[string]WFEntry, error) {
	wf := make(map[string]WFEntry)
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("index.readWFCorpus: Could not wf file: %v", err)
	}
	for i, row := range rawCSVdata {
		if len(row) < 2 {
			return nil, fmt.Errorf("index.readWFCorpus: not enough rows in line %d: %d",
					i, len(row))
		}
		w := row[0] // Chinese text for word
		count, err := strconv.ParseInt(row[1], 10, 0)
		if err != nil {
			return nil, fmt.Errorf("Could not parse word count, %d: %v", i, err)
		}
		wfentry := WFEntry{
			Chinese: w,
			Count:   int(count),
		}
		wf[w] = wfentry
	}
	return &wf, nil
}

// Reads document-specific word frequencies from file into memory
func readWFDoc(wffile io.Reader) (*map[string][]WFDocEntry, error) {
	wfdoc := make(map[string][]WFDocEntry)
	reader := csv.NewReader(wffile)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("index.ReadWFDoc: Could not wf file: %v", err)
	}
	for i, row := range rawCSVdata {
		w := row[0] // Chinese text for word
		count, err := strconv.ParseInt(row[1], 10, 0)
		if err != nil {
			return nil, fmt.Errorf("Could not parse word count %d: %v", i, err)
		}
		filename := row[3]
		entry := WFDocEntry{filename, int(count)}
		if entryarr, ok := wfdoc[w]; !ok {
			wfslice := make([]WFDocEntry, 1)
			wfslice[0] = entry
			wfdoc[w] = wfslice
		} else {
			wfdoc[w] = append(entryarr, entry)
		}
	}
	return &wfdoc, nil
}

// Writes a JSON format keyword index to look up top documents for each keyword
func writeKeywordIndex(indexState IndexState, indexWriter io.Writer) error {
	w := bufio.NewWriter(indexWriter)
	wfdoc := *indexState.wfdoc
	for k, items := range *indexState.wfdoc {
		sort.Sort(ByFrequencyDoc(items))
		if len(items) > maxFilePerKeyword {
			wfdoc[k] = items[:maxFilePerKeyword]
		}
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(*indexState.wfdoc)
	w.Flush()
	return nil
}

// Write corpus analysis to plain text files in the index directory
func WriteWFCorpus(wfStore WordFreqStore, sortedWords,
		sortedUnknownWords []SortedWordItem,
		bFreq []ngram.BigramFreq, wc int, indexConfig IndexConfig) error {

	// Word frequencies
	wfWriter := bufio.NewWriter(wfStore.WFWriter)
	for _, wordItem := range sortedWords {
		rel_freq := 0.0
		if wc > 0 {
			rel_freq = float64(wordItem.Freq) * 10000.0 / float64(wc)
		}
		fmt.Fprintf(wfWriter, "%s\t%d\t%f\n", wordItem.Word, wordItem.Freq,
			rel_freq)
	}
	wfWriter.Flush()

	// Write unknown characters to a text file
	w := bufio.NewWriter(wfStore.UnknownCharsWriter)
	for _, wordItem := range sortedUnknownWords {
		for _, r := range wordItem.Word {
			fmt.Fprintf(w, "U+%X\t%c", r, r)
		}
		fmt.Fprintln(w)
	}
	w.Flush()

	// Write ngrams to a file
	nWriter := bufio.NewWriter(wfStore.BigramWriter)
	for _, ngramItem := range bFreq {
		rel_freq := 0.0
		if wc > 0 {
			rel_freq = float64(ngramItem.Frequency) * 10000.0 / float64(wc)
		}
		fmt.Fprintf(nWriter, "%s\t%d\t%f\n", ngramItem.BigramVal.Traditional(),
			ngramItem.Frequency,	rel_freq)
	}
	nWriter.Flush()
	return nil
}