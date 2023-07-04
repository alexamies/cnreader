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

package generator

import (
	"fmt"
	"text/template"

	"github.com/alexamies/chinesenotes-go/config"
)

// HTML fragment for page head
const head = `
  <head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <link href="https://fonts.googleapis.com/css?family=Noto+Sans" rel="stylesheet">
    <link rel="stylesheet" href="/web/styles.css">
  </head>
`

// header block in HTML body
const header = `
<header>
  <h1>{{.Title}}</h1>
</header>
`

// navigation menu
const nav = `
<nav>
  <ul>
    <li><a href="/">Home</a></li>
    <li><a href="/findtm">Translation Memory</a></li>
    <li><a href="/findadvanced/">Full Text Search</a></li>
    <li><a href="/web/texts.html">Library</a></li>
  </ul>
</nav>
`

// Page footer
const footer = `
    <footer>
      <p>Page updated on {{.DateUpdated}}</p>
      <p>
        Copyright Fo Guang Shan 佛光山 2020.
        The Chinese-English dictionary is reproduced from the <a 
        href="http://ntireader.org/" target="_blank"
        > NTI Buddhist Text Reader</a> under the <a 
        href="https://creativecommons.org/licenses/by-sa/3.0/" target="_blank"
        >Creative Commons Attribution-Share Alike 3.0 License</a>
        (CCASE 3.0). 
        The site is powered by open source
        software under an <a 
        href="http://www.apache.org/licenses/LICENSE-2.0.html"
        >Apache 2.0 license</a>.
        Other content shown in password protected versions of this site is
        copyright protected.
      </p>
    </footer>
`

// Templates from source for zero-config usage
const collectionTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>{{.Data.Title}}</h2>
      <p>{{.Data.Summary}}</p>
      <ul>
        {{ range $element := .Data.CorpusEntries }}
        <li><a href="/{{$element.GlossFile}}">{{ $element.Title }}</a></li>
        {{ end }}
      </ul>
      <p>{{.Data.Intro}}</p>
      <a href="{{.Data.AnalysisFile}}">vocabulary analysis</a>
    </main>
    %s
  </body>
</html>
`

const corpusTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <div>
        <h2>{{.CollectionTitle}}</h2>
        <h3>{{.EntryTitle}}</h3>
      </div>
      <div id="CorpusTextDiv">
      {{.CorpusText}}
      </div>
    </main>
    %s
    <script src="/web/cnotes.js" async></script>
  <body>
</html>
`

const pageTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <div id="CorpusText">
      {{.Content}}
      </div>
    </main>
    %s
  <body>
</html>
`

const corpusAnalysisTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h3 id="lexical">Frequencies of Lexical Words</h3>
      <table>
        <thead>
          <tr>
            <th>Rank</th>
            <th>Frequency</th>
            <th>Chinese</th>
            <th>Pinyin</th>
            <th>English</th>
            <th>Example Usage</th>
          </tr>
        </thead>
        <tbody>
        {{ range $index, $wf := .LexicalWordFreq }}
          <tr>
            <td>{{ add $index 1 }}</td>
            <td>{{ $wf.Freq }}</td>
            <td>{{$wf.Chinese}}</td>
            <td>{{ $wf.Pinyin }}</td>
            <td>{{ $wf.English }}</td>
            <td>{{ $wf.Usage }}</td>
          </tr>
        {{ end }}
        </tbody>
      </table>
      </main>
    %s
  <body>
</html>
`

const textsTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <ul>
        {{ range $index, $entry := .Data.ColIEntries }}
          <li><a href="{{ $entry.GlossFile }}">{{ $entry.Title }}</a></li>
        {{ end }}
      </ul>
      <p><a href="{{ .Data.AnalysisPage }}">Corpus vocabulary analysis</a></p>
    </main>
    %s
  <body>
</html>
`

const corpusSummaryAnalysisTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h3>Frequencies of Lexical Words</h3>
      <table>
        <thead>
          <tr>
            <th>Rank</th>
            <th>Frequency</th>
            <th>Chinese</th>
            <th>Pinyin</th>
            <th>English</th>
            <th>Example Usage</th>
          </tr>
        </thead>
        <tbody>
        {{ range $index, $wf := .LexicalWordFreq }}
          <tr>
            <td>{{ add $index 1 }}</td>
            <td>{{ $wf.Freq }}</td>
            <td><a href="/words/{{$wf.HeadwordId}}.html">{{$wf.Chinese}}</a></td>
            <td>{{ $wf.Pinyin }}</td>
            <td>{{ $wf.English }}</td>
            <td>{{ $wf.Usage }}</td>
          </tr>
        {{ end }}
        </tbody>
      </table>
      </main>
    %s
  <body>
</html>
`

const headwordTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>{{ .Headword.Simplified }}
        {{if ne .Headword.Traditional "\\N"}}
          ({{.Headword.Traditional}})
        {{end}}</span>
      </h2>
        {{ if .Headword.Senses }}
        <ol>
          {{ range $index, $ws := .Headword.Senses }}
            <li>
              <span class='dict-entry-pinyin'>{{ $ws.Pinyin }}</span> 
              {{if ne $ws.MP3 "\\N"}}
                {{if eq $index 0}}
                <span><a href='/mp3/{{ $ws.MP3 }}'
                         title='Play sound'
                         target="_blank"><img src='/images/audio.gif'/></a></span>
                {{ end }}
              {{ end }}
              <span class='dict-entry-grammar'>{{if ne $ws.Grammar "\\N"}}{{ $ws.Grammar }}{{end}}</span> 
              <span class='dict-entry-definition'>{{if ne $ws.English "\\N"}}{{ $ws.English }}{{end}}</span> <br/>
              Domain: {{ $ws.Domain }} {{ $ws.DomainCN }}
              {{ if $ws.Subdomain }}
                , Subdomain: {{ $ws.Subdomain}} {{ $ws.SubdomainCN}}
              {{ end }}
              {{if $ws.Concept }}
                , Concept: {{ $ws.Concept}} {{ $ws.ConceptCN}}
              {{ end }}
              <br/>
              {{if $ws.Notes }}
                <span class='dict-entry-notes'>Notes</span>: {{ $ws.Notes}}
              {{ end }}
              {{if ne $ws.Image "\\N"}}
                <span><img src='/images/{{ $ws.Image }}'/></span>
              {{ end }}
            </li>
          {{ end }}
        </ol>
        {{ end }}
    </main>
    %s
  <body>
</html>
`

const useFileTemplate = `
<!DOCTYPE html>
<html lang="en">
  %s
  <body>
    %s
    %s
    <main>
      <h2>H2 Title</h2>
    </main>
    %s
  <body>
</html>
`

// newTemplateMap builds a template map
func NewTemplateMap(appConfig config.AppConfig) map[string]*template.Template {
	templateMap := make(map[string]*template.Template)
	templDir := appConfig.ProjectHome + "/" + appConfig.GetVar("TemplateDir")
	tNames := map[string]string{
		"about-template.html":                   pageTemplate,
		"advanced-search-template.html":         useFileTemplate,
		"collection-template.html":              collectionTemplate,
		"corpus-analysis-template.html":         corpusAnalysisTemplate,
		"corpus-summary-analysis-template.html": corpusSummaryAnalysisTemplate,
		"corpus-template.html":                  corpusTemplate,
		"headword-template.html":                headwordTemplate,
		"index-template.html":                   useFileTemplate,
		"library-template.html":                 useFileTemplate,
		"login-template.html":                   useFileTemplate,
		"lookup-template.html":                  useFileTemplate,
		"mediadetail.html-template.html":        useFileTemplate,
		"page-template.html":                    useFileTemplate,
		"reference-template.html":               useFileTemplate,
		"texts-template.html":                   textsTemplate,
		"tools-template.html":                   useFileTemplate,
	}
	funcs := template.FuncMap{
		"add":     func(x, y int) int { return x + y },
		"Deref":   func(sp *string) string { return *sp },
		"DerefNe": func(sp *string, s string) bool { return *sp != s },
	}
	if len(templDir) > 0 {
		for tName, defTmpl := range tNames {
			fileName := templDir + "/" + tName
			var tmpl *template.Template
			var err error
			tmpl, err = template.New(tName).Funcs(funcs).ParseFiles(fileName)
			if err != nil {
				//log.Printf("newTemplateMap: error parsing template, using default %s: %v",
				//    tName, err)
				t := fmt.Sprintf(defTmpl, head, header, nav, footer)
				tmpl = template.Must(template.New(tName).Funcs(funcs).Parse(t))
			}
			templateMap[tName] = tmpl
		}
	} else {
		for tName, defTmpl := range tNames {
			t := fmt.Sprintf(defTmpl, head, header, nav, footer)
			tmpl := template.Must(template.New(tName).Funcs(funcs).Parse(t))
			templateMap[tName] = tmpl
		}
	}
	return templateMap
}
