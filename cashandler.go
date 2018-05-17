package carp

import "net/http"

func NewCasRequestHandler(configuration Configuration, app http.Handler) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	browserHandler := casClientFactory.CreateClient().Handle(app)

	effectiveBrowserHandler, err := wrapWithLogoutRedirectionIfNeeded(configuration, browserHandler)
	if err != nil {
		return nil, err
	}

	restHandler := casClientFactory.CreateRestClient().Handle(app)

	return &CasRequestHandler{
		CasBrowserHandler: effectiveBrowserHandler,
		CasRestHandler:	restHandler,
	}, nil
}

func wrapWithLogoutRedirectionIfNeeded(configuration Configuration, handler http.Handler) (http.Handler, error) {
	if logoutRedirectionConfigured(configuration) {
		logoutRedirectionHandler, err := NewLogoutRedirectionHandler(configuration, handler)
		if err != nil {
			return nil, err
		}
		return logoutRedirectionHandler, nil
	} else {
		return handler, nil
	}
}

func logoutRedirectionConfigured(configuration Configuration) bool {
	return configuration.LogoutMethod != "" || configuration.LogoutPath != ""
}

type CasRequestHandler struct {
	CasBrowserHandler http.Handler
	CasRestHandler	http.Handler
}

func (h *CasRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := h.CasRestHandler
	if IsBrowserRequest(r) {
		handler = h.CasBrowserHandler
	}
	handler.ServeHTTP(w, r)
}
