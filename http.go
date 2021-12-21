package carp

import (
	"net/http"
	"strings"
)

func IsBrowserRequest(req *http.Request) bool {
	log.Debug("Entering Method 'IsBrowserRequest'")
	defer func() {
		log.Debug("End of Function 'IsBrowserRequest'")
	}()

	log.Debugf("Param '%s'", req)
	b := isBrowserUserAgent(req.Header.Get("User-Agent")) || isSingleLogoutRequest(req)
	log.Debugf("Variable: %s", b)
	return b
}

func isBrowserUserAgent(userAgent string) bool {
	log.Debug("Entering Method 'isBrowserUserAgent'")
	defer func() {
		log.Debug("End of Function 'isBrowserUserAgent'")
	}()

	log.Debugf("Param '%s'", userAgent)
	lowerUserAgent := strings.ToLower(userAgent)
	log.Debugf("Variable: %s", lowerUserAgent)
	b := strings.Contains(lowerUserAgent, "mozilla") || strings.Contains(lowerUserAgent, "opera")
	log.Debugf("Variable: %s", b)
	return b
}

func isSingleLogoutRequest(r *http.Request) bool {
	log.Debug("Entering Method 'isSingleLogoutRequest'")
	defer func() {
		log.Debug("End of Function 'isSingleLogoutRequest'")
	}()

	log.Debugf("Param '%s'", r)
	if r.Method != "POST" {
		log.Debugf("Condition true: 'r.Method != \"POST\"'")
		return false
	}

	contentType := r.Header.Get("Content-Type")
	log.Debugf("Variable: %s", contentType)
	if contentType != "application/x-www-form-urlencoded" {
		log.Debugf("Condition true: 'contentType != \"application/x-www-form-urlencoded\"'")
		return false
	}

	if v := r.FormValue("logoutRequest"); v == "" {
		log.Debugf("Condition true: 'v := r.FormValue(\"logoutRequest\"); v == \"\"'")
		return false
	}

	return true
}
