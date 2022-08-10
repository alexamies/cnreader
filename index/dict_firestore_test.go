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
	"testing"
)

func TestMergeSubtrings(t *testing.T) {
	tests := []struct {
		name        string
		simplified  string
		traditional string
		want        []string
	}{
		{
			name:        "Simplifed same as traditional",
			simplified:  "看",
			traditional: "",
			want:        []string{"看"},
		},
		{
			name:        "Simplifed different to traditional",
			simplified:  "奥运",
			traditional: "奧運",
			want:        []string{"奥运", "奥", "运", "奧運", "奧", "運"},
		},
	}
	for _, tc := range tests {
		got := mergeSubtrings(tc.simplified, tc.traditional)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s, got %v\n but want %v.", tc.name, got, tc.want)
		}
	}
}
