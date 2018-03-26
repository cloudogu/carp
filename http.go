package main

import "strings"

func IsBrowserRequest(userAgent string) bool {
	lowerUserAgent := strings.ToLower(userAgent)
	return strings.Contains(lowerUserAgent, "mozilla") || strings.Contains(lowerUserAgent, "opera")
}
