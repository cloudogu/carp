package carp

import (
	"github.com/golang/glog"
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
	return &LogoutRedirectionHandler{
		logoutUrl:    configuration.CasUrl + "/logout",
		delegate:     delegateHandler,
		logoutMethod: configuration.LogoutMethod,
		logoutPath:   configuration.LogoutPath,
	}
}

func (h *LogoutRedirectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.isLogoutRequest(r) {
		glog.Infof("Detected logout request; redirecting to %s", h.logoutUrl)
		http.Redirect(w, r, h.logoutUrl, http.StatusSeeOther)
		return
	}
	h.delegate.ServeHTTP(w, r)
}

func (h *LogoutRedirectionHandler) isLogoutRequest(r *http.Request) bool {
	glog.Infof("Inspecting request %s url %s", r.Method, r.URL)
	return (h.logoutMethod != "" || h.logoutPath != "") &&
		(h.logoutMethod == "" || r.Method == h.logoutMethod) &&
		(h.logoutPath == "" || h.logoutPath != "" && strings.HasSuffix(r.URL.Path, h.logoutPath))
}
