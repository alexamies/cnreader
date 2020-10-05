/*
Library for bigram type
*/
package ngram

import (
	"fmt"
	"log"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

var NULL_BIGRAM_PTR *Bigram

// A struct to hold an instance of a Bigram
// Since they could be either simplified or traditional, index by the headword
//ids. Also, include an example of the bigram so that usage context can be
// investigated
type Bigram struct {
	HeadwordDef1 *dicttypes.Word  // First headword
	HeadwordDef2 *dicttypes.Word  // Second headword
	Example, ExFile, ExDocTitle, ExColTitle *string
}

// Constructor for a Bigram struct
func NewBigram(hw1, hw2 dicttypes.Word,
		example, exFile, exDocTitle, exColTitle string) *Bigram {
	if hw1.Senses == nil || hw2.Senses == nil {
		msg := "bigram.NewBigram: nil reference"
		if len(hw1.Simplified) != 0 {
			msg += fmt.Sprintf(", Simplified1 = %s", hw1.Simplified)
		}
		if len(hw2.Simplified) != 0 {
			msg += fmt.Sprintf(", Simplified2 = %s", hw2.Simplified)
		}
		if len(hw1.Traditional) != 0 {
			msg += fmt.Sprintf(", Traditional1 = %s", hw1.Traditional)
		}
		if len(hw2.Traditional) != 0 {
			msg += fmt.Sprintf(", Traditional2 = %s", hw2.Traditional)
		}
		msg += fmt.Sprintf("in %s, %s", exFile, exColTitle)
		log.Printf(msg)
	}
	return &Bigram{
		HeadwordDef1: &hw1,
		HeadwordDef2: &hw2,
		Example: &example,
		ExFile: &exFile,
		ExDocTitle: &exDocTitle,
		ExColTitle: &exColTitle,
	}
}

func NullBigram() *Bigram {
	if NULL_BIGRAM_PTR == nil {
		hw1 := dicttypes.Word{
			HeadwordId: 0,
			Simplified: "", 
			Traditional: "",
			Pinyin: "",
			Senses: []dicttypes.WordSense{},
		}
		NULL_BIGRAM_PTR = NewBigram(hw1, hw1, "", "", "", "")
	}
	return NULL_BIGRAM_PTR
}

// For comparison of bigrams
func bigramKey(id1, id2 int) string {
	return fmt.Sprintf("%.7d %.7d", id1, id2)
}

// Bigrams that contain function words should be excluded
func (bigram *Bigram) ContainsFunctionWord() bool {
	if bigram.HeadwordDef1.Senses == nil || bigram.HeadwordDef2.Senses == nil {
		msg := "bigram.ContainsFunctionWord: no value"
		if len(bigram.HeadwordDef1.Simplified) != 0 {
			msg += fmt.Sprintf(", Simplified1 = %s",  
				bigram.HeadwordDef1.Simplified)
		}
		if len(bigram.HeadwordDef2.Simplified) != 0 {
			msg += fmt.Sprintf(", Simplified2 = %s",  
				bigram.HeadwordDef1.Simplified)
		}
		log.Printf(msg)
		return false
	}
	ws1 := bigram.HeadwordDef1.Senses[0]
	ws2 := bigram.HeadwordDef2.Senses[0]
	return ws1.IsFunctionWord() || ws2.IsFunctionWord()
}

// The simplified text of the bigram
func (bigram *Bigram) Simplified() string {
	if len(bigram.HeadwordDef1.Simplified) == 0 || len(bigram.HeadwordDef2.Simplified) == 0 {
		msg := "bigram.Simplified no value"
		log.Printf(msg)
		return msg
	}
	return fmt.Sprintf("%s%s", bigram.HeadwordDef1.Simplified,
		bigram.HeadwordDef2.Simplified)
}

// Override string method for comparison
func (bigram *Bigram) String() string {
	return bigramKey(bigram.HeadwordDef1.HeadwordId, bigram.HeadwordDef2.HeadwordId)
}

// The traditional text of the bigram
func (bigram *Bigram) Traditional() string {
	if len(bigram.HeadwordDef1.Traditional) == 0 || len(bigram.HeadwordDef2.Traditional) == 0 {
		msg := "bigram.Traditional(): no value"
		log.Printf(msg)
		return msg
	}
	t1 := bigram.HeadwordDef1.Traditional
	if t1 == "\\N" {
		t1 = bigram.HeadwordDef1.Simplified
	}
	t2 := bigram.HeadwordDef2.Traditional
	if t2 == "\\N" {
		t2 = bigram.HeadwordDef2.Simplified
	}
	return fmt.Sprintf("%s%s", t1, t2)
}
