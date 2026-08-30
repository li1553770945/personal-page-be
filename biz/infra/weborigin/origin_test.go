package weborigin

import "testing"

func TestAllowed(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "missing origin", origin: "", want: true},
		{name: "production apex", origin: "https://peacesheep.xyz", want: true},
		{name: "production www", origin: "https://www.peacesheep.xyz", want: true},
		{name: "pages production", origin: "https://personal-page-fe-new.pages.dev", want: true},
		{name: "pages preview", origin: "https://96d8b7d4.personal-page-fe-new.pages.dev", want: true},
		{name: "localhost dev", origin: "http://localhost:3000", want: true},
		{name: "pages http rejected", origin: "http://personal-page-fe-new.pages.dev", want: false},
		{name: "lookalike suffix rejected", origin: "https://evilpersonal-page-fe-new.pages.dev", want: false},
		{name: "foreign pages project rejected", origin: "https://example.pages.dev", want: false},
		{name: "path rejected", origin: "https://personal-page-fe-new.pages.dev/attack", want: false},
		{name: "query rejected", origin: "https://personal-page-fe-new.pages.dev?attack=1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Allowed(tt.origin); got != tt.want {
				t.Fatalf("Allowed(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}
