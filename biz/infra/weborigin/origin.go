package weborigin

import (
	"net/url"
	"strings"
)

const pagesHostname = "personal-page-fe-new.pages.dev"

var productionOrigins = map[string]struct{}{
	"https://peacesheep.xyz":                 {},
	"https://www.peacesheep.xyz":             {},
	"https://api.peacesheep.xyz":             {},
	"https://personal-page-fe-new.pages.dev": {},
}

// Allowed accepts the production site, its Cloudflare Pages deployments, and
// local HTTP development origins. It deliberately validates the parsed host so
// lookalike domains cannot pass with a suffix-only string check.
func Allowed(rawOrigin string) bool {
	origin := strings.TrimSpace(rawOrigin)
	if origin == "" {
		return true
	}
	if _, ok := productionOrigins[origin]; ok {
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return false
	}
	hostname := strings.ToLower(parsed.Hostname())
	if parsed.Scheme == "http" && hostname == "localhost" {
		return true
	}
	return parsed.Scheme == "https" && parsed.Port() == "" && strings.HasSuffix(hostname, "."+pagesHostname)
}
