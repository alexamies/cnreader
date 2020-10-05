/*
Library for constructing a glossary for a text or collection of texts

A glossary consists of all the words that occur in the collection of texts and
that match the given domain.
*/
package analysis

import (
	"sort"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

// The content for a corpus entry
type Glossary struct {
	Domain string
	Words []dicttypes.WordSense
}

// Makes a glossary by filtering by the domain label and sorting by Chinese
// pinyin.
func MakeGlossary(domain string, headwords []dicttypes.Word) Glossary {
	hws := dicttypes.WordSenses{}
	if len(domain) == 0 {
		return Glossary{domain, hws}
	}
	for _, hw := range headwords {
		for _, ws := range hw.Senses {
			if ws.Domain == domain && ws.Grammar != "proper noun" {
				hws = append(hws, ws)
			}
		}
	}
	sort.Sort(hws)
	return Glossary{domain, hws}
}

// Makes a list of proper nouns, sorted by Pinyin
func makePNList(vocab map[string]int, wdict map[string]dicttypes.Word) dicttypes.Words {
	hws := dicttypes.Words{}
	for w, _ := range vocab {
		hw, ok := wdict[w]
			if ok && hw.IsProperNoun() {
				hws = append(hws, hw)
			}
	}
	sort.Sort(hws)
	return hws
}