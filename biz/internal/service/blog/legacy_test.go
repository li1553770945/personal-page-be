package blog

import (
	"encoding/json"
	"testing"
)

func TestEmbeddedLegacyPosts(t *testing.T) {
	var seeds []legacyPostSeed
	if err := json.Unmarshal(legacyPostsJSON, &seeds); err != nil {
		t.Fatalf("decode legacy posts: %v", err)
	}
	if len(seeds) != 96 {
		t.Fatalf("expected 96 legacy posts, got %d", len(seeds))
	}
	seen := make(map[string]bool, len(seeds))
	for _, seed := range seeds {
		if seed.Slug == "" || seed.Title == "" || seed.PublishedAt == "" {
			t.Fatalf("legacy post has missing required data: %#v", seed)
		}
		if seen[seed.Slug] {
			t.Fatalf("duplicate legacy slug: %s", seed.Slug)
		}
		seen[seed.Slug] = true
	}
}
