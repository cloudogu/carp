package carp

import (
	"net/http"
)

// NewCasRequestHandler creates a CasRequestHandler that wraps the given http.Handler and adds CAS-Authentication to the request
func NewCasRequestHandler(configuration Configuration, handler http.Handler) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	browserHandler := casClientFactory.CreateClient().Handle(handler)

	return &CasRequestHandler{
		wrappedHandler:    handler,
		CasBrowserHandler: wrapWithLogoutRedirectionIfNeeded(configuration, browserHandler),
		CasRestHandler:    casClientFactory.CreateRestClient().Handle(handler),
	}, nil
}

func wrapWithLogoutRedirectionIfNeeded(configuration Configuration, handler http.Handler) http.Handler {
	if logoutRedirectionConfigured(configuration) {
		log.Info("Found configuration for logout redirection")
		logoutRedirectionHandler := NewLogoutRedirectionHandler(configuration, handler)
		return logoutRedirectionHandler
	} else {
		log.Info("No configuration for logout redirection found")
		return handler
	}
}

func logoutRedirectionConfigured(configuration Configuration) bool {
	return configuration.LogoutMethod != "" || configuration.LogoutPath != ""
}

type CasRequestHandler struct {
	wrappedHandler    http.Handler
	CasBrowserHandler http.Handler
	CasRestHandler    http.Handler
}

func (h *CasRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if IsServiceAccountAuthentication(r) {
		// no cas-authentication needed -> skip cas-handler
		h.wrappedHandler.ServeHTTP(w, r)
		return
	}

	handler := h.CasRestHandler
	if IsBrowserRequest(r) {
		handler = h.CasBrowserHandler
	}
	handler.ServeHTTP(w, r)
}
