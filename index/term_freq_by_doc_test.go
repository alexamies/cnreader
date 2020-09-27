package index

import (
	"testing"
)

// Trivial test for document index writing
func TestWriteWFDoc0(t *testing.T) {
	wfMap := TermFreqDocMap{}
	df := DocumentFrequency{}
	wfMap.WriteToFile(df, "terms_doc_test.txt", mockIndexConfig())
}

// Simple test for document index writing
func TestWriteWFDoc1(t *testing.T) {
	wfRec := TermFreqDocRecord{"鐵", 1, 10, "test.html", "testDoc.html"}
	wfMap := TermFreqDocMap{}
	wfMap.Put(wfRec)
	df := DocumentFrequency{}
	wfMap.WriteToFile(df, "terms_doc_test.txt", mockIndexConfig())
}
