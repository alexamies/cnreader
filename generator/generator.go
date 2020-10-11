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

//Package for generating HTML files
package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/alexamies/chinesenotes-go/dicttypes"
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/corpus"
)

const collectionTemplate = `
<!DOCTYPE html>
<html lang="en">
  <body>
    <h2>{{.Title}}</h2>
    <p>{{.Summary}}</p>
    <ul>
      {{ range $element := .CorpusEntries }}
      <li><ahref='/{{$element.GlossFile}}'>{{ $element.Title }}</a></li>
      {{ end }}
    </ul>
    {{.Intro}}
    <a href="/analysis/{{.AnalysisFile}}">vocabulary analysis</a>
    <p>Page updated on {{.DateUpdated}}</p>
  </body>
</html>
`
const corpusTemplate = `
<!DOCTYPE html>
<html lang="en">
  <body>
    <h2>{{.CollectionTitle}}</h2>
    <h3>{{.EntryTitle}}</h3>
    <p>{{.CorpusText}}</p>
    <p><a href='/analysis/{{.AnalysisFile}}'>vocabulary analysis</a></p>
    <p>Page updated on {{.DateUpdated}}</p>
  </body>
</html>
`
// HTMLOutPutConfig holds parameters for writing output to HTML
type HTMLOutPutConfig struct {
	ContainsByDomain string
	Domain string
	GoStaticDir string
	TemplateDir string
	VocabFormat string
	WebDir string
}

// CorpusEntryContent holds the content for a corpus entry
type CorpusEntryContent struct {
	CorpusText, DateUpdated, CollectionURL, CollectionTitle, EntryTitle, AnalysisFile string
}

// decodeUsageExample formats usage example text into links with highlight
//   Return
//      marked up text with links and highlight
func DecodeUsageExample(usageText string, headword dicttypes.Word,
		dictTokenizer tokenizer.Tokenizer, outputConfig HTMLOutPutConfig,
		wdict map[string]dicttypes.Word) string {
	tokens := dictTokenizer.Tokenize(usageText)
	replacementText := ""
	for _, token := range tokens {
		word := token.DictEntry
		if word.Simplified == headword.Simplified || word.Traditional == headword.Traditional {
			replacementText = replacementText +
				"<span class='usage-highlight'>" + token.Token + "</span>"
		} else {
			ws, ok := wdict[word.Simplified]
			if ok {
				replacementText = replacementText + hyperlink(ws, token.Token, outputConfig.VocabFormat)
			} else {
				replacementText = replacementText + token.Token
			}
		}
	}
	return replacementText
}

// Constructs a hyperlink for a headword, including Pinyin and English in the
// title attribute for the link mouseover
func hyperlink(w dicttypes.Word, text, vocabFormat string) string {
	classTxt := "vocabulary"
	if w.IsProperNoun() {
		classTxt = classTxt + " propernoun"
	}
	pinyin := w.Pinyin
	english := ""
	if len(w.Senses) > 0 {
		english = w.Senses[0].English
	}
	if len(w.Senses) > 1 {
		english = ""
		for i, entry := range w.Senses {
			english += fmt.Sprintf("%d. %s, ", i + 1, entry.English)
		}
		english = english[0:len(english)-2]
	}
	return fmt.Sprintf(vocabFormat, pinyin, english, classTxt,
			w.HeadwordId, text)
}

// span constructs a HTML span element for a headword
// in the title attribute for the mouseover and headword id in the microdata
// 'data' attrbute.
func span(w dicttypes.Word, text string) string {
	classTxt := "vocabulary"
	if w.IsProperNoun() {
		classTxt = classTxt + " propernoun"
	}
	pinyin := w.Pinyin
	english := ""
	if len(w.Senses) > 0 {
		english = w.Senses[0].English
	}
	if len(w.Senses) > 1 {
		english = ""
		for i, entry := range w.Senses {
			english += fmt.Sprintf("%d. %s, ", i + 1, entry.English)
		}
		english = english[0:len(english)-2]
	}
	vocabFormat := `<span title="%s | %s" class="%s" itemprop="HeadwordId" value="%d">%s</span>`
	return fmt.Sprintf(vocabFormat, pinyin, english, classTxt, w.HeadwordId, text)
}

// WriteCollectionFile writes a HTML file describing the collection
// Parameters:
//   collectionFile: The name of the file describing the collection
//   baseDir: The base directory for writing the file
func WriteCollectionFile(entry corpus.CollectionEntry, analysisFile string,
		outputConfig HTMLOutPutConfig, corpusConfig corpus.CorpusConfig,
		corpusEntries []corpus.CorpusEntry, introText string) error {
	entry.CorpusEntries = corpusEntries
	entry.AnalysisFile = analysisFile
	fName := outputConfig.WebDir + "/" + entry.GlossFile
	f, err := os.Create(fName)
	if err != nil {
		return fmt.Errorf("Error creating collection output file: %v ", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	// Replace name of intro file with introduction text
	entry.Intro = introText
	entry.DateUpdated = time.Now().Format("2006-01-02")
	var tmpl *template.Template
	if len(outputConfig.TemplateDir) > 0 {
		templFile := outputConfig.TemplateDir + "/collection-template.html"
		tmpl = template.Must(template.New(
					"collection-template.html").ParseFiles(templFile))
	} else {
		tmpl = template.Must(template.New(
					"collection-template.html").Parse(collectionTemplate))
	}
	err = tmpl.Execute(w, entry)
	if err != nil {
		return fmt.Errorf("Error executing collection-template: %v ", err)
	}
	return nil
}


// WriteCorpusDoc writes a corpus document with markup for the array of tokens
// tokens: A list of tokens forming the document
// vocab: A list of word id's in the document
// filename: The file name to write to
// HTML template to use
// collectionURL: the URL of the collection that the corpus text belongs to
// collectionTitle: The collection title that the corpus entry belongs to
// aFile: The vocabulary analysis file written to or empty string for none
// sourceFormat: TEXT, or HTML used for formatting output
func WriteCorpusDoc(tokens []tokenizer.TextToken, vocab map[string]int, w io.Writer,
		collectionURL string, collectionTitle string, entryTitle string,
		aFile string, sourceFormat string, outputConfig HTMLOutPutConfig,
		corpusConfig corpus.CorpusConfig, wdict map[string]dicttypes.Word) error {

	var b bytes.Buffer
	replacer := strings.NewReplacer("\n", "<br/>")

	// Iterate over text chunks
	for _, token := range tokens {
		chunk := token.Token
		if entries, ok := wdict[chunk]; ok && !corpus.IsExcluded(corpusConfig.Excluded, chunk) {
			fmt.Fprintf(&b, span(entries, chunk))
		} else {
			if sourceFormat != "HTML" {
				chunk = replacer.Replace(chunk)
			}
			b.WriteString(chunk)
		}
	}

	textContent := b.String()
	dateUpdated := time.Now().Format("2006-01-02")
	content := CorpusEntryContent{textContent, dateUpdated, collectionURL,
		collectionTitle, entryTitle, aFile}

	var tmpl *template.Template
	if len(outputConfig.TemplateDir) > 0 {
		templFile := outputConfig.TemplateDir + "/corpus-template.html"
		tmpl = template.Must(template.New("corpus-template.html").ParseFiles(templFile))
	} else {
		tmpl = template.Must(template.New(
					"corpus-template.html").Parse(corpusTemplate))
	}
	err := tmpl.Execute(w, content)
	if err != nil {
		return fmt.Errorf("could not execute template: %v", err)
	}
	return nil
}