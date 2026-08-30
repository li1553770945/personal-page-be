package blog

import (
	"strings"
	"testing"
)

func TestValidateBlogCommentContent(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "trim", value: "  写得很好  ", want: "写得很好"},
		{name: "empty", value: " \n\t ", wantErr: true},
		{name: "too long", value: strings.Repeat("好", maxBlogCommentRunes+1), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateBlogCommentContent(test.value)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected validation error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestIsBlogCommentStatus(t *testing.T) {
	for _, status := range []string{"pending", "approved", "rejected"} {
		if !isBlogCommentStatus(status) {
			t.Fatalf("expected %q to be valid", status)
		}
	}
	if isBlogCommentStatus("deleted") {
		t.Fatal("unexpected valid status")
	}
}
