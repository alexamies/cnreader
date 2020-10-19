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
	"bytes"
	"io"
	"testing"
)

// Trivial test for excluded string
func TestLoadExcluded(t *testing.T) {
	t.Log("corpus.TestIsExcluded0: Begin unit tests")
	var buf bytes.Buffer
	io.WriteString(&buf, "如需引用文章")
	excluded := LoadExcluded(&buf)
	tests := []struct {
		name string
		input string
		expectExcluded bool
	}{
		{
			name: "empty",
			input: "",
			expectExcluded: false,
		},
		{
			name: "has included",
			input: "如需引用文章",
			expectExcluded: true,
		},
	}
	for _, tc := range tests {
		if tc.expectExcluded && !IsExcluded(excluded, tc.input) {
			t.Errorf("%s, expected '%s' to be excluded", tc.name, tc.input)
		}
	}
}
