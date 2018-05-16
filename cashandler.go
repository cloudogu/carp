package carp

import "net/http"

func NewCasRequestHandler(configuration Configuration, app http.Handler) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	logoutHandler := &RedirectingHandler{
		logoutUrl: configuration.CasUrl + "/logout",
		delegate: casClientFactory.CreateClient().Handle(app),
	}
	restHandler := casClientFactory.CreateRestClient().Handle(app)

	return &CasRequestHandler{
		CasBrowserHandler: logoutHandler,
		CasRestHandler:	restHandler,
	}, nil
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
