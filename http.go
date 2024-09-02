package carp

import (
	"net/http"
	"strings"
)

func IsBrowserRequest(req *http.Request) bool {
	return isBrowserUserAgent(req.Header.Get("User-Agent")) || isSingleLogoutRequest(req)
}

func isBrowserUserAgent(userAgent string) bool {
	lowerUserAgent := strings.ToLower(userAgent)
	return strings.Contains(lowerUserAgent, "mozilla") || strings.Contains(lowerUserAgent, "opera")
}

func isSingleLogoutRequest(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		return false
	}

	if v := r.FormValue("logoutRequest"); v == "" {
		return false
	}

	return true
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusResponseWriter) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	s.statusCode = statusCode
}
