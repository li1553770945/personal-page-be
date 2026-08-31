package blog

import "testing"

func TestNormalizeBlogSort(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "defaults to pinned", value: "", want: "pinned"},
		{name: "keeps pinned", value: "pinned", want: "pinned"},
		{name: "accepts time", value: "time", want: "time"},
		{name: "normalizes case", value: " TIME ", want: "time"},
		{name: "rejects unknown mode", value: "importance", want: "pinned"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeBlogSort(test.value); got != test.want {
				t.Fatalf("normalizeBlogSort(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}
