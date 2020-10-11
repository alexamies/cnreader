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
	"fmt"
	"testing"
)

// Test sorting of word frequencies
func TestSortedFreq1(t *testing.T) {
	fmt.Printf("TestSortedFreq: Begin unit tests\n")
	wordFreq := map[string]int{"one": 1, "three": 3, "two": 2}
	sortedWords := SortedFreq(wordFreq)
	if sortedWords == nil {
		t.Error("Expected non-nil sortedWords")
	}
	if sortedWords[0].Word != "three" {
		t.Error("Expected that 'three' to be the most frequent word")
	}
	/*
		for _, w := range sortedWords {
			fmt.Printf("TestSortedFreq: %v : %v\n", w, wordFreq[w])
		}
	*/
}
