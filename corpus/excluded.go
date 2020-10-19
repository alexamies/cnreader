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

package corpus

import (
	"encoding/csv"
	"fmt"
	"io"
)

// Tests whether the string should be excluded from corpus analysis
// Parameter
// chunk: the string to be tested
func IsExcluded(excluded map[string]bool, text string) bool  {
	_, ok := excluded[text]
	return ok
}

func LoadExcluded(file io.Reader) (*map[string]bool, error) {
	// log.Print("corpus.loadExcluded enter")
	excluded := make(map[string]bool)
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.Comma = rune('\t')
	reader.Comment = rune('#')
	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Error reading excluded text", err)
	}
	for _, row := range rawCSVdata {
		if len(row) < 1 {
			return nil, fmt.Errorf("corpus.loadExcluded: no columns in row")
	  }
		excluded[row[0]] = true
	}
	return &excluded, nil
}