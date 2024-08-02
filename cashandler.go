package carp

import (
	"net/http"
)

func NewCasRequestHandler(configuration Configuration) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	casHandler := createCasHandlerFunc(configuration)

	browserHandler := casClientFactory.CreateClient().Handle(casHandler)

	return &CasRequestHandler{
		CasBrowserHandler: wrapWithLogoutRedirectionIfNeeded(configuration, browserHandler),
	}, nil
}

func createCasHandlerFunc(configuration Configuration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO implement with stuff from the general handler
	})
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
