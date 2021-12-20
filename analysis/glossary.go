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
	"sort"

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

// The content for a corpus entry
type Glossary struct {
	Domain string
	Words  dicttypes.Words
}

// Makes a glossary by filtering by the domain label and sorting by Chinese
// pinyin.
func MakeGlossary(domain string, headwords []dicttypes.Word) Glossary {
	hws := dicttypes.Words{}
	if len(domain) == 0 {
		return Glossary{domain, hws}
	}
	for _, hw := range headwords {
		if len(hw.Senses) > 0 {
			ws := hw.Senses[0]
			if ws.Domain == domain && ws.Grammar != "proper noun" {
				hws = append(hws, hw)
			}
		}
	}
	sort.Sort(hws)
	return Glossary{domain, hws}
}

// Makes a list of proper nouns, sorted by Pinyin
func makePNList(vocab map[string]int, wdict map[string]dicttypes.Word) dicttypes.Words {
	hws := dicttypes.Words{}
	for w := range vocab {
		hw, ok := wdict[w]
		if ok && hw.IsProperNoun() {
			hws = append(hws, hw)
		}
	}
	sort.Sort(hws)
	return hws
}
