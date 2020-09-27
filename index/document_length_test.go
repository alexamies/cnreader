package index

import (
	"testing"
)

// Empty test for corpus index writing
func TestWriteDocLengthToFile(t *testing.T) {
	dl := DocLength{"test.html", 42}
	dlArray := []DocLength{dl}
	indexconfig := mockIndexConfig()
	WriteDocLengthToFile(dlArray, "doc_length_test.tsv", indexconfig)
}
