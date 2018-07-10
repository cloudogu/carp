package carp

import (
	"github.com/golang/glog"
	"net/http"
)

func NewCasRequestHandler(configuration Configuration, app http.Handler, forwardUnauthenticatedRequests bool) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	browserHandler := casClientFactory.CreateClient().Handle(app)

	return &CasRequestHandler{
		CasBrowserHandler: wrapWithLogoutRedirectionIfNeeded(configuration, browserHandler),
		CasRestHandler:    casClientFactory.CreateRestClient().Handle(app, forwardUnauthenticatedRequests),
	}, nil
}

func wrapWithLogoutRedirectionIfNeeded(configuration Configuration, handler http.Handler) http.Handler {
	if logoutRedirectionConfigured(configuration) {
		glog.Infoln("Found configuration for logout redirection")
		logoutRedirectionHandler := NewLogoutRedirectionHandler(configuration, handler)
		return logoutRedirectionHandler
	} else {
		glog.Infoln("No configuration for logout redirection found")
		return handler
	}
}

func logoutRedirectionConfigured(configuration Configuration) bool {
	return configuration.LogoutMethod != "" || configuration.LogoutPath != ""
}

type CasRequestHandler struct {
	CasBrowserHandler http.Handler
	CasRestHandler    http.Handler
}

func (h *CasRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := h.CasRestHandler
	if IsBrowserRequest(r) {
		handler = h.CasBrowserHandler
	}
	handler.ServeHTTP(w, r)
}
