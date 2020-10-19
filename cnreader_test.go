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

package main

import (
	"container/list"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/alexamies/chinesenotes-go/config"	
	"github.com/alexamies/chinesenotes-go/fileloader"	
	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/analysis"
	"github.com/alexamies/cnreader/corpus"
)
	
var integration = flag.Bool("integration", false, "run an integration test")

func listToString(l list.List) string {
	text := ""
	for e := l.Front(); e != nil; e = e.Next() {
		text += e.Value.(string)
	}
	return text
}

func printList(t *testing.T, l list.List) {
	for e := l.Front(); e != nil; e = e.Next() {
		t.Log(e.Value)
	}
}

func testCorpusConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{
		CorpusDataDir: "data/corpus",
		CorpusDir: "corpus",
		Excluded: map[string]bool{},
		ProjectHome: ".",
	}
}

// TestMain runs integration tests if the flag -integration is set
func TestMain(m *testing.M) {
	flag.Parse()
	if *integration {
		fmt.Println("Running integration test")
		os.Exit(m.Run())
	}
	fmt.Println("Skipping integration test")
}

func TestIntegration(t *testing.T) {
	t.Log("TestIntegration begin")
	appConfig := config.AppConfig{
		LUFileNames: []string{"testdata/testwords.txt"},
	}
	wdict, err := fileloader.LoadDictFile(appConfig)
	if err != nil {
		t.Fatalf("main, could not load dict: %v", err)
	}
	corpusConfig := testCorpusConfig()
	corpusLoader := corpus.NewCorpusLoader(corpusConfig)
	r, err := os.Open("testdata/test-trad.html")
	if err != nil {
		t.Fatalf("main, could not open file: %v", err)
	}
	defer r.Close()
	text := corpusLoader.ReadText(r)
	tok := tokenizer.DictTokenizer{wdict}
	tokens, results := analysis.ParseText(text, "", corpus.NewCorpusEntry(),
			tok, corpusConfig, wdict)
	tokenText := listToString(tokens)
	if len(text) != len(tokenText) {
		t.Error("Expected to string length ", len(text), ", got ",
			len(tokenText))
		printList(t, tokens)
	}
	if results.CCount != 4 {
		t.Fatalf("Expected to get char count 4, got %d", results.CCount)
	}
	if len(results.Vocab) != 2 {
		t.Errorf("Expected to get Vocab 2, got %d: %v", len(results.Vocab),
			results.Vocab)
	}
}
