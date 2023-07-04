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

// Package for translation memory index
package tmindex

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

const (
	fnameUni    = "tmindex_unigram.tsv"
	fnameDomain = "tmindex_uni_domain.tsv"
	minChars    = 3
	maxChars    = 11
)

// BuildIndexes saves unigram indexes with and without domain in a file
func BuildIndexes(indexDir string, wdict map[string]*dicttypes.Word) error {
	pathUni := fmt.Sprintf("%s/%s", indexDir, fnameUni)
	fUni, err := os.Create(pathUni)
	if err != nil {
		return fmt.Errorf("could not create index file %s, err: %v", fnameUni, err)
	}
	defer fUni.Close()
	wUni := bufio.NewWriter(fUni)
	err = buildUnigramIndex(wUni, wdict)
	if err != nil {
		return fmt.Errorf("could not write to index file %s, err: %v", fnameUni, err)
	}

	pathDomain := fmt.Sprintf("%s/%s", indexDir, fnameDomain)
	fDomain, err := os.Create(pathDomain)
	if err != nil {
		return fmt.Errorf("could not create index file %s, err: %v", fnameDomain, err)
	}
	defer fDomain.Close()
	wDomain := bufio.NewWriter(fDomain)
	err = buildUniDomainIndex(wDomain, wdict)
	if err != nil {
		return fmt.Errorf("could not write to index file %s, err: %v", fnameDomain, err)
	}
	return nil
}

// buildUniDomainIndex builds a unigram index with domain
func buildUniDomainIndex(w io.Writer, wdict map[string]*dicttypes.Word) error {
	tmindexUni := make(map[string]bool)
	for term, word := range wdict {
		for _, sense := range word.Senses {
			for _, c := range term {
				line := fmt.Sprintf("%c\t%s\t%s\n", c, term, sense.Domain)
				// log.Printf("buildUniDomainIndex, line: %s", line)
				if _, ok := tmindexUni[line]; ok {
					continue
				}
				_, err := io.WriteString(w, line)
				if err != nil {
					return err
				}
				tmindexUni[line] = true
			}
		}
	}
	return nil
}

// buildUnigramIndex builds a unigram index, skip for strings > maxChars
func buildUnigramIndex(w io.Writer, wdict map[string]*dicttypes.Word) error {
	tmindexUni := make(map[string]bool)
	for term := range wdict {
		if len([]rune(term)) > maxChars {
			continue
		}
		for _, c := range term {
			line := fmt.Sprintf("%c\t%s\n", c, term)
			if _, ok := tmindexUni[line]; ok {
				continue
			}
			_, err := io.WriteString(w, line)
			if err != nil {
				return err
			}
			tmindexUni[line] = true
		}
	}
	return nil
}
