package main

import (
	"net/url"
	"strings"
)

func originAllowed(origin string, allowedOrigins []string) bool {
	requestURL, ok := parseOrigin(origin)
	if !ok {
		return false
	}

	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if strings.Contains(allowed, "://*.") {
			scheme, suffix, found := strings.Cut(allowed, "://*.")
			if !found || (scheme != "https" && scheme != "http") ||
				strings.ContainsAny(suffix, "/:?#@") || !strings.EqualFold(requestURL.Scheme, scheme) ||
				requestURL.Port() != "" {
				continue
			}

			host := strings.ToLower(requestURL.Hostname())
			suffix = strings.ToLower(suffix)
			subdomain, found := strings.CutSuffix(host, "."+suffix)
			if found && subdomain != "" && !strings.Contains(subdomain, ".") {
				return true
			}
			continue
		}

		allowedURL, valid := parseOrigin(allowed)
		if valid && originString(requestURL) == originString(allowedURL) {
			return true
		}
	}

	return false
}

func parseOrigin(origin string) (*url.URL, bool) {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Opaque != "" || parsed.User != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, false
	}
	return parsed, true
}

func originString(origin *url.URL) string {
	return strings.ToLower(origin.Scheme + "://" + origin.Host)
}
