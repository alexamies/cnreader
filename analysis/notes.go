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
	"regexp"
	"strings"
)

// Processes notes with a regular expression
type notesProcessor struct {
	patterns []*regexp.Regexp
	replaces []string
}

// newNotesProcessor creates a new notesProcessor
// Param
//   patternList a list of patterns to match regular expressions, quoted and delimited by commas
//   replaceList a list of replacement regular expressions, same cardinality
func newNotesProcessor(patternList, replaceList string) notesProcessor{
	p := strings.Split(patternList, `","`)
	patterns := []*regexp.Regexp{}
	for _, t := range p {
		pattern := strings.Trim(t, ` "`)
		re := regexp.MustCompile(pattern)
		patterns =  append(patterns, re)
	}
	r := strings.Split(replaceList, ",")
	replaces := []string{}
	for _, t := range r {
		replace := strings.Trim(t, ` "`)
		replaces =  append(replaces, replace)
	}
	return notesProcessor {
		patterns: patterns,
		replaces: replaces,
	}
}

// processes the notes
func (p notesProcessor) process(notes string) string {
	s := notes
	for i, re := range p.patterns {
		if re.MatchString(notes) {
			s = re.ReplaceAllString(s, p.replaces[i])
		}
	}
	return s
}