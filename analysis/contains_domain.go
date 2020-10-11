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

package analysis

import (
	"github.com/alexamies/chinesenotes-go/dicttypes"	
	"github.com/alexamies/cnreader/generator"
	"strings"
)

// Max number of words to display for contained in
const MAX_CONTAINED_BY = 50

// Filters the list of headwords to only those in the configured domain
func ContainsByDomain(contains []dicttypes.Word,
		outputConfig generator.HTMLOutPutConfig) []dicttypes.Word {
	domains := outputConfig.ContainsByDomain
	containsBy := []dicttypes.Word{}
	containsSet := make(map[int]bool)
	count := 0
	for _, hw := range contains {
		for _, ws := range hw.Senses {
		  _, ok := containsSet[hw.HeadwordId]
			if !ok && count < MAX_CONTAINED_BY && strings.Contains(domains, ws.Domain) {
				containsBy = append(containsBy, hw)
				containsSet[hw.HeadwordId] = true  // Do not add it twice
				count++ // don't go over max number
			}
		}
	}
	return containsBy
}

// Subtract the items in the second list from the first
func Subtract(headwords, subtract []dicttypes.Word) []dicttypes.Word {
	subtracted := []dicttypes.Word{}
	subtractSet := make(map[int]bool)
	for _, hw := range subtract {
    subtractSet[hw.HeadwordId] = true
	}
	for _, hw := range headwords {
		if _, ok := subtractSet[hw.HeadwordId]; !ok {
			subtracted = append(subtracted, hw)
		}
	}
	return subtracted
}
