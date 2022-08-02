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
	"reflect"
	"strings"
	"testing"
)

func TestTitleSubtrings(t *testing.T) {
	tests := []struct {
		name   string
		input1 string
		input2 string
		want   []string
	}{
		{
			name:   "Empty",
			input1: "",
			input2: "",
			want:   []string{},
		},
		{
			name:   "Three characters",
			input1: "看",
			input2: "世界",
			want:   []string{"看世界", "看", "世界", "世", "界"},
		},
	}
	for _, tc := range tests {
		got := titleSubtrings(tc.input1, tc.input2)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s, got %v\n but want %v", tc.name, got, tc.want)
		}
	}
}

func TestSubstrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Empty",
			input: "",
			want:  []string{},
		},
		{
			name:  "One character",
			input: "世",
			want:  []string{"世"},
		},
		{
			name:  "Two characters",
			input: "世界",
			want:  []string{"世界", "世", "界"},
		},
		{
			name:  "Three characters",
			input: "看世界",
			want:  []string{"看世界", "看", "世界", "世", "界"},
		},
	}
	for _, tc := range tests {
		chars := strings.Split(tc.input, "")
		got := substrings(chars)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s, got %v\n but want %v", tc.name, got, tc.want)
		}
	}
}
