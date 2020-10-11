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

package corpus

import (
	"testing"
)

// Trivial test for excluded string
func TestIsExcluded0(t *testing.T) {
	t.Log("corpus.TestIsExcluded0: Begin unit tests")
	config := mockCorpusConfig()
	if IsExcluded(config.Excluded, "") {
		t.Error("corpus.TestIsExcluded0: Do not expect '' to be excluded")
	}
}

// Easy test for excluded string
func TestIsExcluded1(t *testing.T) {
	t.Log("corpus.TesIsExcluded: Begin unit tests")
	config := mockCorpusConfig()
	if !IsExcluded(config.Excluded, "如需引用文章") {
		t.Error("corpus.TestIsExcluded0: Expect '如需引用文章' to be excluded")
	}
}
