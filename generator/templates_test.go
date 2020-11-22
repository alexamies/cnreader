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

package generator

import (
	"testing"

	"github.com/alexamies/chinesenotes-go/config"
)

// TestNewTemplateMap building the template map
func TestNewTemplateMap(t *testing.T) {
	type test struct {
		name string
		config config.AppConfig
  }
  tests := []test{
		{
			name: "Empty config",
			config: config.AppConfig{},
		},
	}
  for _, tc := range tests {
		templates := NewTemplateMap(tc.config)
		_, ok := templates["corpus-template.html"]
		if !ok {
			t.Errorf("%s, template not found", tc.name)
		}
	}
}
