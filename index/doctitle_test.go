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

package index

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/alexamies/cnreader/corpus"
	"github.com/alexamies/cnreader/library"
)

// Implements the CorpusLoader interface with no data
type mockCorpusLoader struct {
	collections []corpus.CollectionEntry
}

func (loader mockCorpusLoader) GetCollectionEntry(fName string) (*corpus.CollectionEntry, error) {
	return &corpus.CollectionEntry{}, nil
}

func (loader mockCorpusLoader) GetConfig() corpus.CorpusConfig {
	return corpus.CorpusConfig{}
}

func (loader mockCorpusLoader) LoadAll(r io.Reader) (*map[string]corpus.CorpusEntry, error) {
	data := map[string]corpus.CorpusEntry{}
	return &data, nil
}

func (loader mockCorpusLoader) LoadCollection(fName, colTitle string) (*[]corpus.CorpusEntry, error) {
	for _, col := range loader.collections {
		if col.CollectionFile == fName {
			return &col.CorpusEntries, nil
		}
	}
	return &[]corpus.CorpusEntry{}, nil
}

func (loader mockCorpusLoader) LoadCollections() (*[]corpus.CollectionEntry, error) {
	return &loader.collections, nil
}

func (loader mockCorpusLoader) LoadCorpus(r io.Reader) (*[]corpus.CollectionEntry, error) {
	return &loader.collections, nil
}

func (loader mockCorpusLoader) ReadText(src string) (string, error) {
	return "你好 Hello!", nil
}

// A mock LibraryLoader 
type mockLibraryLoader struct {
	corpLoader corpus.CorpusLoader
}

func (loader mockLibraryLoader) GetCorpusLoader() corpus.CorpusLoader {
	return loader.corpLoader
}

func (loader mockLibraryLoader) LoadLibrary(r io.Reader) (*[]library.CorpusData, error) {
	data := []library.CorpusData{}
	return &data, nil
}

// Table tests for BuildDocTitleIndex
func TestBuildDocTitleIndex(t *testing.T) {
	emptyCorpLoader := mockCorpusLoader{}
	emptyLibLoader := mockLibraryLoader{emptyCorpLoader}
	entry := corpus.CorpusEntry{
		RawFile: "raw_file.txt",
		GlossFile: "gloss_file.html",
		Title: "標題 Entry Title",
		ColTitle: "Collection Title",
	}
	collection := corpus.CollectionEntry{
		CollectionFile: "collection.txt",
		GlossFile: "collection.html",
		Title: "My collection",
		Summary: "xyz",
		Intro: "abc",
		CorpusEntries: []corpus.CorpusEntry{entry},
	}
	collections := []corpus.CollectionEntry{collection}
	smallCorpLoader := mockCorpusLoader{collections}
	smallLibLoader := mockLibraryLoader{smallCorpLoader}

	tests := []struct {
		name string
		libLoader library.LibraryLoader
		want string
	}{
		{
			name: "Empty",
			libLoader: emptyLibLoader,
			want: "# plain_text_file",
		},
		{
			name: "Small collection",
			libLoader: smallLibLoader,
			want: "\t標題\t",
		},
	}
	for _, tc := range tests {
		var buf bytes.Buffer
		err := BuildDocTitleIndex(tc.libLoader, &buf)
		if err != nil {
			t.Fatalf("TestBuildDocTitleIndex %s, error: %v", tc.name, err)
		}
		got := buf.String()
		if len(got) == 0 {
			t.Fatalf("%s, no data written", tc.name)
		}
		if !strings.Contains(got, tc.want) {
			t.Errorf("%s, got %s\n but want %s", tc.name, got, tc.want)
		}
	}
}

// Table tests for splitTitle
func TestSplitTitle(t *testing.T) {
	tests := []struct {
		name string
		input string
		wantCn string
		wantEn string
	}{
		{
			name: "Empty",
			input: "",
			wantCn: "",
			wantEn: "",
		},
		{
			name: "Chinese first",
			input: "標題 Title",
			wantCn: "標題",
			wantEn: "Title",
		},
	}
	for _, tc := range tests {
		titleCn, titleEn := splitTitle(tc.input)
		if titleCn != tc.wantCn {
			t.Fatalf("TestSplitTitle %s, got %s, wantCn %v", tc.name, titleCn, tc.wantCn)
		}
		if titleEn != tc.wantEn {
			t.Fatalf("TestSplitTitle %s, got %s, wanwantEnt: %v", tc.name, titleEn, tc.wantEn)
		}
	}
}
