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
