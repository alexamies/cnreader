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

package analysis

import (
	"testing"
)

// TestNotesProcessor tests processing of notes
func TestNotesProcessor(t *testing.T) {
	testCases := []struct {
		name string
		match  string
		replace  string
		notes  string
		expect string
	}{
		{
			name: "Empty",
			match: "", 
			replace: "", 
			notes: "hello", 
			expect: "hello", 
		},
		{
			name: "Basic capture",
			match: `(T 1)`, 
			replace: "<a>${1}</a>", 
			notes: "T 1", 
			expect: "<a>T 1</a>", 
		},
		{
			name: "Single digit",
			match: `(T ([0-9]))(\)|,|;)`, 
			replace: `<a href="/taisho/t000${2}.html">${1}</a>${3}`, 
			notes: "; T 2)", 
			expect: `; <a href="/taisho/t0002.html">T 2</a>)`, 
		},
		{
			name: "Two digits",
			match: `(T ([0-9]{2}))(\)|,|;)`, 
			replace: `<a href="/taisho/t00${2}.html">${1}</a>${3}`, 
			notes: " T 12,", 
			expect: ` <a href="/taisho/t0012.html">T 12</a>,`, 
		},
		{
			name: "No match",
			match: `(T ([0-9]{2}))(\)|,|;)`, 
			replace: `<a href="/taisho/t00${2}.html">${1}</a>${3}`, 
			notes: "Testing 123", 
			expect: `Testing 123`, 
		},
		{
			name: "Granddaddy",
			match: `"(T ([0-9]))(\)|,|;)","(T ([0-9]{2}))(\)|,|;)"`, 
			replace: `"<a href="/taisho/t000${2}.html">${1}</a>${3}","<a href="/taisho/t00${2}.html">${1}</a>${3}"`, 
			notes: " T 23;", 
			expect: ` <a href="/taisho/t0023.html">T 23</a>;`, 
		},
	}
	for _, tc := range testCases {
		processor := newNotesProcessor(tc.match, tc.replace)
		got := processor.process(tc.notes)
		if got != tc.expect {
			t.Errorf("TestNotesProcessor %s: got %s, want %s", tc.name, got,
					tc.expect)
		}
	}
}