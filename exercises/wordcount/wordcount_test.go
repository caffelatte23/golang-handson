package wordcount

import (
	"maps"
	"testing"
)

func TestWordCount(t *testing.T) {
	tests := []struct {
		in   string
		want map[string]int
	}{
		{"I am learning Go!", map[string]int{"I": 1, "am": 1, "learning": 1, "Go!": 1}},
		{"The quick brown fox jumped over the lazy dog.", map[string]int{
			"The": 1, "quick": 1, "brown": 1, "fox": 1, "jumped": 1, "over": 1, "the": 1, "lazy": 1, "dog.": 1}},
		{"I ate a donut. Then I ate another donut.", map[string]int{
			"I": 2, "ate": 2, "a": 1, "donut.": 2, "Then": 1, "another": 1}},
		{"  go  go\tgo\n", map[string]int{"go": 3}},
		{"", map[string]int{}},
	}
	for _, tt := range tests {
		got := WordCount(tt.in)
		if got == nil || !maps.Equal(got, tt.want) {
			t.Errorf("WordCount(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
