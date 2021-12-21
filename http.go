package carp

import (
	"net/http"
	"strings"
)

func IsBrowserRequest(req *http.Request) bool {
	log.Debug("Entering Function 'IsBrowserRequest'")
	log.Debugf("Param '%s'", req)
	b := isBrowserUserAgent(req.Header.Get("User-Agent")) || isSingleLogoutRequest(req)
	log.Debugf("Variable: %s", b)
	log.Debug("End of Function 'IsBrowserRequest'")
	return b
}

func isBrowserUserAgent(userAgent string) bool {
	log.Debug("Entering Function 'isBrowserUserAgent'")
	log.Debugf("Param '%s'", userAgent)
	lowerUserAgent := strings.ToLower(userAgent)
	log.Debugf("Variable: %s", lowerUserAgent)
	b := strings.Contains(lowerUserAgent, "mozilla") || strings.Contains(lowerUserAgent, "opera")
	log.Debugf("Variable: %s", b)
	log.Debug("End of Function 'isBrowserUserAgent'")
	return b
}

func isSingleLogoutRequest(r *http.Request) bool {
	log.Debug("Entering Function 'isSingleLogoutRequest'")
	log.Debugf("Param '%s'", r)
	if r.Method != "POST" {
		log.Debugf("Condition true: 'r.Method != \"POST\"'")
		log.Debug("End of Function 'isSingleLogoutRequest'")
		return false
	}

	contentType := r.Header.Get("Content-Type")
	log.Debugf("Variable: %s", contentType)
	if contentType != "application/x-www-form-urlencoded" {
		log.Debugf("Condition true: 'contentType != \"application/x-www-form-urlencoded\"'")
		log.Debug("End of Function 'isSingleLogoutRequest'")
		return false
	}

	if v := r.FormValue("logoutRequest"); v == "" {
		log.Debugf("Condition true: 'v := r.FormValue(\"logoutRequest\"); v == \"\"'")
		log.Debug("End of Function 'isSingleLogoutRequest'")
		return false
	}

	log.Debug("End of Function 'isSingleLogoutRequest'")
	return true
}
