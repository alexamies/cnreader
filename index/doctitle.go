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

// Functions for building and accessing the document title index.

package index

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/alexamies/chinesenotes-go/tokenizer"
	"github.com/alexamies/cnreader/library"
)

// Builds a flat index of document titles from the hierarchical corpus.
// This is suitable for loading into the database or loading from for the web
// app when running without a database.
func BuildDocTitleIndex(libLoader library.LibraryLoader, w io.Writer) error {
	corpLoader := libLoader.GetCorpusLoader()
	collections, err := corpLoader.LoadCollections()
	if err != nil {
		return fmt.Errorf("BuildDocTitleIndex could not load corpus: %v", err)
	}
	log.Printf("BuildDocTitleIndex loaded %d collections", len(*collections))
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = '\t'
	headers := []string{"# plain_text_file", "gloss_file", "title_cn", "title_en",
			"col_gloss_file", "col_title", "col_plus_doc_title"}
	if err := csvWriter.Write(headers); err != nil {
			return fmt.Errorf("BuildDocTitleIndex error writing headers: %v", err)
	}
	for _, c := range *collections {
		log.Printf("BuildDocTitleIndex, collection: %s\n", c.Title)
		docs, err := corpLoader.LoadCollection(c.CollectionFile, c.Title)
		if err != nil {
			return fmt.Errorf("BuildDocTitleIndex error loading %s: %v",
					c.CollectionFile, err)
		}
		for _, d := range *docs {
			title_cn, title_en := splitTitle(d.Title)
			title := c.Title + ": " + d.Title
			record := []string{d.RawFile, d.GlossFile, title_cn, title_en,
					c.GlossFile, c.Title, title}
			if err := csvWriter.Write(record); err != nil {
					return fmt.Errorf("BuildDocTitleIndex error writing record: %v", err)
			}
		}
	}
	csvWriter.Flush()
	return nil
}

// Split title with both Chinese and English mixed into separate parts
// Args
//   title A mixed Chinese and English title
// Return
//   The Chinese part of the title and (second) the English part
func splitTitle(title string) (string, string) {
	titleCn := ""
	titleEn := ""
	segments := tokenizer.Segment(title)
	if len(segments) == 0 {
		log.Printf("splitTitle, no title segments: %s\n", title)
	} else {
		if len(segments[0].Text) > 0 {
			if segments[0].Chinese {
				titleCn = segments[0].Text
			} else {
				titleEn = strings.Trim(segments[0].Text, " ")
			}
		}
		if len(segments) > 1 {
			if len(segments[1].Text) > 0 {
				if segments[1].Chinese {
					titleCn = segments[1].Text
				} else {
					titleEn = strings.Trim(segments[1].Text, " ")
				}
			}
		}
	}
	return titleCn, titleEn
}