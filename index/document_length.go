/*
Library for word frequencies for each doc in the corpus
*/
package index

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// Word frequencies for each document
const DOC_LENGTH_FILE = "doc_length.tsv"

// Records the document length for each document in the corpus
type DocLength struct {
	GlossFile string
	WordCount int
}

// Append document analysis to a plain text file in the index directory
func WriteDocLengthToFile(dlArray []DocLength, fileName string, indexConfig IndexConfig) {
	log.Printf("index.WriteDocLengthToFile: enter, %s, %d\n", fileName,
		len(dlArray))
	fname := indexConfig.IndexDir + "/" + fileName
	f, err := os.Create(fname)
	if err != nil {
		log.Printf("Could not open write wfFile: %v\n", err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, record := range dlArray {
		fmt.Fprintf(w, "%s\t%d\n", record.GlossFile, record.WordCount)
	}
	w.Flush()
}