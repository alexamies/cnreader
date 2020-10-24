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
  "log"
  "text/template"

	"github.com/alexamies/chinesenotes-go/config"
)

// Templates from source for zero-config usage
const collectionTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <h2>{{.Title}}</h2>
    <p>{{.Summary}}</p>
    <ul>
      {{ range $element := .CorpusEntries }}
      <li><a href="/{{$element.GlossFile}}"">{{ $element.Title }}</a></li>
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
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <main>
      <h1>{{.Title}}</h1>
      <header>
        <h2>{{.CollectionTitle}}</h2>
        <h3>{{.EntryTitle}}</h3>
      </header>
      <div>
      {{.CorpusText}}
      </div>
    </main>
    <footer>
      <div>Page updated on {{.DateUpdated}}</div>
    </footer>
  <body>
</html>
`

const pageTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <main>
      <h1>{{.Title}}</h1>
      <div>
      {{.Content}}
      </div>
    </main>
    <footer>
      <div>Page updated on {{.DateUpdated}}</div>
    </footer>
  <body>
</html>
`

const corpusAnalysisTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <main>
      <h2>{{.Title}}</h2>
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
            <td><a href="/words/{{$wf.HeadwordId}}.html">{{$wf.Chinese}}</a></td>
            <td>{{ $wf.Pinyin }}</td>
            <td>{{ $wf.English }}</td>
            <td>{{ $wf.Usage }}</td>
          </tr>
        {{ end }}
        </tbody>
      </table>
      </main>
    <footer>
      <div>Page updated on {{.DateUpdated}}</div>
    </footer>
  <body>
</html>
`

const textsTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <main>
      <h2>{{.Title}}</h2>
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
            <td><a href="/words/{{$wf.HeadwordId}}.html">{{$wf.Chinese}}</a></td>
            <td>{{ $wf.Pinyin }}</td>
            <td>{{ $wf.English }}</td>
            <td>{{ $wf.Usage }}</td>
          </tr>
        {{ end }}
        </tbody>
      </table>
      </main>
    <footer>
      <div>Page updated on {{.DateUpdated}}</div>
    </footer>
  <body>
</html>
`

const corpusSummaryAnalysisTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <main>
      <h2>{{.Title}}</h2>
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
    <footer>
      <div>Page updated on {{.DateUpdated}}</div>
    </footer>
  <body>
</html>
`

// newTemplateMap builds a template map
func NewTemplateMap(appConfig config.AppConfig) map[string]*template.Template {
  templateMap := make(map[string]*template.Template)
  templDir := appConfig.GetVar("TemplateDir")
  tNames := map[string]string{
    "collection-template.html": collectionTemplate,
    "corpus-template.html": corpusTemplate,
    "texts-template.html": pageTemplate,
    "corpus-analysis-template.html": corpusAnalysisTemplate,
    "corpus-summary-analysis-template.html": corpusSummaryAnalysisTemplate,
  }
  funcs := template.FuncMap{
    "add": func(x, y int) int { return x + y },
    "Deref":   func(sp *string) string { return *sp },
    "DerefNe": func(sp *string, s string) bool { return *sp != s },
  }
  if len(templDir) > 0 {
    for tName, defTmpl := range tNames {
      fileName := templDir + "/" + tName
      var tmpl *template.Template
      var err error
      tmpl, err = template.New(tName).ParseFiles(fileName)
      if err != nil {
        log.Printf("newTemplateMap: error parsing template, using default %s: %v",
            tName, err)
        tmpl = template.Must(template.New(tName).Funcs(funcs).Parse(defTmpl))
      }
      templateMap[tName] = tmpl
    }
  } else {
    for tName, defTmpl := range tNames {
      tmpl := template.Must(template.New(tName).Funcs(funcs).Parse(defTmpl))
      templateMap[tName] = tmpl
    }
  }
  return templateMap
}
