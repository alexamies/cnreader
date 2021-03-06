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
	"io"
)

// Word frequencies for each document
const DocLengthFile = "doc_length.tsv"

// Records the document length for each document in the corpus
type DocLength struct {
	GlossFile string
	WordCount int
}

// Append document analysis to a plain text file in the index directory
func WriteDocLengthToFile(dlArray []DocLength, f io.Writer) {
	w := bufio.NewWriter(f)
	for _, record := range dlArray {
		fmt.Fprintf(w, "%s\t%d\n", record.GlossFile, record.WordCount)
	}
	w.Flush()
}