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
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/alexamies/chinesenotes-go/dictionary"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateDictIndex writes a list of dicitonary words with subtring array
func UpdateDictIndex(ctx context.Context, client FsClient, dict *dictionary.Dictionary, corpus string, generation int) error {
	fsCol := fmt.Sprintf("%s_dict_substring_%d", corpus, generation)
	log.Printf("UpdateDictSubstringIndex loaded %d headwords", len(dict.HeadwordIds))
	for _, hw := range dict.HeadwordIds {
		ref := client.Collection(fsCol).Doc(hw.Pinyin)
		_, err := ref.Get(ctx)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("UpdateDocTitleIndex, Failed getting tf for ref %v: %v", ref, err)
			}
		}
		if hw.Traditional == "\\N" {
			hw.Traditional = ""
		}
		ss := mergeSubtrings(hw.Simplified, hw.Traditional)
		hws := dictionary.HeadwordSubstrings{
			HeadwordId:  int64(hw.HeadwordId),
			Simplified:  hw.Simplified,
			Traditional: hw.Traditional,
			Substrings:  ss,
		}
		_, err = ref.Set(ctx, hws)
		if err != nil {
			return fmt.Errorf("failed setting entry for ref %v: %v", hws, err)
		}
	}
	return nil
}

func mergeSubtrings(simplified, traditional string) []string {
	s := strings.Split(simplified, "")
	ss := dictionary.Ngrams(s, 1)
	t := strings.Split(traditional, "")
	tt := dictionary.Ngrams(t, 1)
	m := map[string]bool{}
	for _, w := range ss {
		m[w] = true
	}
	for _, w := range tt {
		if !m[w] {
			ss = append(ss, w)
		}
	}
	return ss
}
