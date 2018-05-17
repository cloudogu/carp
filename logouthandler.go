package carp

import (
	"net/http"
	"strings"
)

type LogoutRedirectionHandler struct {
	logoutUrl string
	delegate http.Handler
	logoutMethod string
	logoutPath string
}

func NewLogoutRedirectionHandler(configuration Configuration, delegateHandler http.Handler) (http.Handler, error) {
	return &LogoutRedirectionHandler{
		logoutUrl: configuration.CasUrl + "/logout",
		delegate: delegateHandler,
		logoutMethod: configuration.LogoutMethod,
		logoutPath: configuration.LogoutPath,
	}, nil
}

func (h *LogoutRedirectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.isLogoutRequest(r) {
		http.Redirect(w, r, h.logoutUrl, 303)
		return
	}
	h.delegate.ServeHTTP(w, r)
}

func (h *LogoutRedirectionHandler) isLogoutRequest(r *http.Request) bool {
	return (r.Method == "" || r.Method == h.logoutMethod) &&
		(h.logoutPath == "" || strings.HasSuffix(r.RequestURI, h.logoutPath))
}
