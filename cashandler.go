package carp

import (
	"net/http"
)

func NewCasRequestHandler(configuration Configuration, app http.Handler) (http.Handler, error) {
	log.Debug("Entering Method 'NewCasRequestHandler'")
	log.Debugf("Param '%s'", configuration)
	log.Debugf("Param '%s'", app)
	casClientFactory, err := NewCasClientFactory(configuration)
	log.Debugf("Variable: %s", casClientFactory)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, err
	}

	browserHandler := casClientFactory.CreateClient().Handle(app)
	log.Debugf("Variable: %s", browserHandler)

	needed := wrapWithLogoutRedirectionIfNeeded(configuration, browserHandler)
	log.Debugf("Variable: %s", needed)
	handle := casClientFactory.CreateRestClient().Handle(app)
	log.Debugf("Variable: %s", handle)
	log.Debug("End of Function 'NewCasRequestHandler'")
	return &CasRequestHandler{
		CasBrowserHandler: needed,
		CasRestHandler:    handle,
	}, nil
}

func wrapWithLogoutRedirectionIfNeeded(configuration Configuration, handler http.Handler) http.Handler {
	log.Debug("Entering Method 'wrapWithLogoutRedirectionIfNeeded'")
	log.Debugf("Param '%s'", configuration)
	log.Debugf("Param '%s'", handler)
	if logoutRedirectionConfigured(configuration) {
		log.Debugf("Condition true: 'logoutRedirectionConfigured(configuration)'")
		log.Info("Found configuration for logout redirection")
		logoutRedirectionHandler := NewLogoutRedirectionHandler(configuration, handler)
		log.Debugf("Variable: %s", logoutRedirectionHandler)
		log.Debug("End of Function 'wrapWithLogoutRedirectionIfNeeded'")
		return logoutRedirectionHandler
	} else {
		log.Info("No configuration for logout redirection found")
		log.Debugf("Variable: %s", handler)
		log.Debug("End of Function 'wrapWithLogoutRedirectionIfNeeded'")
		return handler
	}
}

func logoutRedirectionConfigured(configuration Configuration) bool {
	log.Debug("Entering Method 'logoutRedirectionConfigured'")
	log.Debugf("Param '%s'", configuration)
	log.Debugf("Variable: %s", configuration.LogoutMethod != "" || configuration.LogoutPath != "")
	log.Debug("End of Function 'logoutRedirectionConfigured'")
	return configuration.LogoutMethod != "" || configuration.LogoutPath != ""
}

type CasRequestHandler struct {
	CasBrowserHandler http.Handler
	CasRestHandler    http.Handler
}

func (h *CasRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("Entering Method 'ServeHTTP'")
	log.Debugf("Param '%s'", w)
	log.Debugf("Param '%s'", r)
	handler := h.CasRestHandler
	log.Debugf("Variable: %s", handler)
	if IsBrowserRequest(r) {
		log.Debugf("Condition true: 'IsBrowserRequest(r)'")
		handler = h.CasBrowserHandler
	}
	log.Debug("End of Function 'ServeHTTP'")
	handler.ServeHTTP(w, r)
}
