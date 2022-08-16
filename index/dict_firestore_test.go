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

	"github.com/alexamies/chinesenotes-go/dicttypes"
)

func TestIsDomain(t *testing.T) {
	s1 := "莲花"
	t1 := "蓮花"
	p1 := "liánhuā"
	e1 := "lotus"
	s2 := "域"
	t2 := "\\N"
	p2 := "yù"
	e2 := "district; region"
	hw1 := dicttypes.Word{
		HeadwordId:  1,
		Simplified:  s1,
		Traditional: t1,
		Pinyin:      p1,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  1,
				Simplified:  s1,
				Traditional: t1,
				Pinyin:      p1,
				English:     e1,
			},
		},
	}
	hw2 := dicttypes.Word{
		HeadwordId:  2,
		Simplified:  s2,
		Traditional: t2,
		Pinyin:      p2,
		Senses: []dicttypes.WordSense{
			{
				HeadwordId:  2,
				Simplified:  s2,
				Traditional: t2,
				Pinyin:      p2,
				English:     e2,
				Domain: "Idiom",
			},
		},
	}
	tests := []struct {
		name  	string
		w 			*dicttypes.Word
		domain 	string
		want    bool
	}{
		{
			name:     "Is not Idiom",
			w:  			&hw1,
			domain: 	"Idiom",
			want:     false,
		},
		{
			name:     "Is an Idiom",
			w:  			&hw2,
			domain: 	"Idiom",
			want:     true,
		},
	}
	for _, tc := range tests {
		got := isDomain(tc.w, tc.domain)
		if got != tc.want {
			t.Errorf("%s, got %v\n but want %v.", tc.name, got, tc.want)
		}
	}
}

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
