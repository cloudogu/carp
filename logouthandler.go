package carp

import (
	"net/http"
	"strings"
)

type LogoutRedirectionHandler struct {
	logoutUrl    string
	delegate     http.Handler
	logoutMethod string
	logoutPath   string
}

func NewLogoutRedirectionHandler(configuration Configuration, delegateHandler http.Handler) http.Handler {
	log.Debug("Entering Method 'NewLogoutRedirectionHandler'")
	defer func() {
		log.Debug("End of Function 'NewLogoutRedirectionHandler'")
	}()

	log.Debugf("Param '%s'", configuration)
	log.Debugf("Param '%s'", delegateHandler)
	url := configuration.CasUrl + "/logout"
	log.Debugf("Variable: %s", url)
	log.Debugf("Variable: %s", delegateHandler)
	log.Debugf("Variable: %s", configuration.LogoutMethod)
	log.Debugf("Variable: %s", configuration.LogoutPath)
	return &LogoutRedirectionHandler{
		logoutUrl:    url,
		delegate:     delegateHandler,
		logoutMethod: configuration.LogoutMethod,
		logoutPath:   configuration.LogoutPath,
	}
}

func (h *LogoutRedirectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("Entering Method 'ServeHTTP'")
	defer func() {
		log.Debug("End of Function 'ServeHTTP'")
	}()

	log.Debugf("Param '%s'", w)
	log.Debugf("Param '%s'", r)
	if h.isLogoutRequest(r) {
		log.Debugf("Condition true: 'h.isLogoutRequest(r)'")
		log.Infof("Detected logout request; redirecting to %s", h.logoutUrl)
		http.Redirect(w, r, h.logoutUrl, http.StatusSeeOther)
		return
	}

	h.delegate.ServeHTTP(w, r)
}

func (h *LogoutRedirectionHandler) isLogoutRequest(r *http.Request) bool {
	log.Debug("Entering Method 'isLogoutRequest'")
	defer func() {
		log.Debug("End of Function 'isLogoutRequest'")
	}()

	log.Debugf("Param '%s'", r)
	log.Infof("Inspecting request %s url %s", r.Method, r.URL)
	b := (h.logoutMethod != "" || h.logoutPath != "") &&
		(h.logoutMethod == "" || r.Method == h.logoutMethod) &&
		(h.logoutPath == "" || h.logoutPath != "" && strings.HasSuffix(r.URL.Path, h.logoutPath))
	log.Debugf("Variable: %s", b)
	return b
}
