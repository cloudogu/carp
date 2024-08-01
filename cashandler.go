package carp

import (
	"net/http"
)

func NewCasRequestHandler(configuration Configuration, app http.Handler) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	browserHandler := casClientFactory.CreateClient().Handle(app)

	return &CasRequestHandler{
		CasBrowserHandler: wrapWithLogoutRedirectionIfNeeded(configuration, browserHandler),
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
	CasBrowserHandler http.Handler
}

func (h *CasRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("casHandler: Serving request %s...", r.URL.String())
	h.CasBrowserHandler.ServeHTTP(w, r)
}
