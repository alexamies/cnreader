// Test sorting of glossary
package analysis

import (
	"testing"

	"github.com/alexamies/chinesenotes-go/dicttypes"	
)

// make example data
func makeHW0() dicttypes.Word {
	simp := "国"
	trad := "國"
	pinyin := "guó"
	english := "a country / a state / a kingdom"
	topic_cn := "古文"
	topic_en := "Classical Chinese"
	ws0 := dicttypes.WordSense{
		Id: 1075,
		HeadwordId: 1075,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		English: english,
		Grammar: "noun",
		ConceptCN: "\\N",
		Concept: "\\N",
		DomainCN: topic_cn,
		Domain: topic_en,
		SubdomainCN: "",
		Subdomain: "",
		Image: "",
		MP3: "",
		Notes: "",
	}
	wsArray := []dicttypes.WordSense{ws0}
	return dicttypes.Word{
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// make example data
func makeHW1() dicttypes.Word {
	simp := "严净"
	trad := "嚴淨"
	pinyin := "yán jìng"
	english := "majestic and pure"
	topic_cn := "佛教"
	topic_en := "Buddhism"
	ws0 := dicttypes.WordSense{
		Id: 62267,
		HeadwordId: 62267,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		English: english,
		Grammar: "set phrase",
		ConceptCN: "\\N",
		Concept: "\\N",
		DomainCN: topic_cn,
		Domain: topic_en,
		SubdomainCN: "",
		Subdomain: "",
		Image: "",
		MP3: "",
		Notes: "",
	}
	wsArray := []dicttypes.WordSense{ws0}
	return dicttypes.Word{
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// make example data
func makeHW2() dicttypes.Word {
	simp := "缘缘"
	trad := "緣緣"
	pinyin := "yuányuán"
	english := "observed object condition"
	topic_cn := "佛教"
	topic_en := "Buddhism"
	ws0 := dicttypes.WordSense{
		Id: 62252,
		HeadwordId: 62252,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		English: english,
		Grammar: "set phrase",
		ConceptCN: "\\N",
		Concept: "\\N",
		DomainCN: topic_cn,
		Domain: topic_en,
		SubdomainCN: "",
		Subdomain: "",
		Image: "",
		MP3: "",
		Notes: "",
	}
	wsArray := []dicttypes.WordSense{ws0}
	return dicttypes.Word{
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// make example data
func makeHW3() dicttypes.Word {
	simp := "禅"
	trad := "禪"
	pinyin := "chán"
	english := "meditative concentration"
	topic_cn := "佛教"
	topic_en := "Buddhism"
	ws0 := dicttypes.WordSense{
		Id: 3182,
		HeadwordId: 3182,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		English: english,
		Grammar: "noun",
		ConceptCN: "\\N",
		Concept: "\\N",
		DomainCN: topic_cn,
		Domain: topic_en,
		SubdomainCN: "",
		Subdomain: "",
		Image: "",
		MP3: "",
		Notes: "",
	}
	wsArray := []dicttypes.WordSense{ws0}
	return dicttypes.Word{
		HeadwordId: 1,
		Simplified: simp,
		Traditional: trad,
		Pinyin: pinyin,
		Senses: wsArray,
	}
}

// Trivial test of MakeGlossary function
func TestMakeGlossary0(t *testing.T) {
	t.Log("analysis.TestMakeGlossary0: Begin ********")
	headwords := []dicttypes.Word{}
	MakeGlossary("test_label", headwords)
}

// Easy test of MakeGlossary function
func TestMakeGlossary1(t *testing.T) {
	hw := makeHW0()
	headwords := []dicttypes.Word{hw}
	glossary := MakeGlossary("test_label", headwords)
	len := len(glossary.Words)
	expected := 0
	if expected != len {
		t.Error("analysis.TestMakeGlossary2: Expected ", expected, ", got",
			len)
	}
}

// Happy path test of MakeGlossary function
func TestMakeGlossary2(t *testing.T) {
	hw0 := makeHW0()
	hw1 := makeHW1()
	headwords := []dicttypes.Word{hw0, hw1}
	glossary := MakeGlossary("Buddhism", headwords)
	len := len(glossary.Words)
	expected := 1
	if expected != len {
		t.Error("analysis.TestMakeGlossary2: Expected ", expected, ", got",
			len)
	}
}

// Test sorting in MakeGlossary method
func TestMakeGlossary3(t *testing.T) {
	hw0 := makeHW0()
	hw1 := makeHW1()
	hw2 := makeHW2()
	headwords := []dicttypes.Word{hw0, hw2, hw1}
	glossary := MakeGlossary("Buddhism", headwords)
	len := len(glossary.Words)
	expected := 2
	if expected != len {
		t.Error("TestMakeGlossary3: Expected ", expected, ", got",
			len)
	}
	//fmt.Println("analysis.TestMakeGlossary3, words: ", glossary.Words)
	firstWord := glossary.Words[0].Pinyin[0]
	pinyinExpected := hw1.Pinyin[0]
	if pinyinExpected != firstWord {
		t.Error("analysis.TestMakeGlossary3: Expected first ", pinyinExpected,
			", got", firstWord)
	}
}

// Test sorting in MakeGlossary method
func TestMakeGlossary4(t *testing.T) {
	hw0 := makeHW0()
	hw1 := makeHW1()
	hw2 := makeHW2()
	hw3 := makeHW3()
	headwords := []dicttypes.Word{hw0, hw2, hw1, hw3}
	glossary := MakeGlossary("Buddhism", headwords)
	len := len(glossary.Words)
	expected := 3
	if expected != len {
		t.Error("TestMakeGlossary4: Expected ", expected, ", got",
			len)
	}
	//fmt.Println("analysis.TestMakeGlossary4, words: ", glossary.Words)
	firstWord := glossary.Words[0].Pinyin[0]
	firstExpected := hw3.Pinyin[0]
	secondExpected := hw1.Pinyin[0]
	//result := firstExpected < secondExpected
	//fmt.Printf("analysis.TestMakeGlossary4, %s < %s = %v\n", firstExpected,
	//	secondExpected, result)
	if firstExpected != firstWord {
		t.Error("analysis.TestMakeGlossary3: Expected first ", firstExpected,
			", got", firstWord)
	}
	secondWord := glossary.Words[1].Pinyin[0]
	if secondExpected != secondWord {
		t.Error("analysis.TestMakeGlossary3: Expected second ", secondExpected,
			", got", secondWord)
	}
}