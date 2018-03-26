package main

import "net/http"

func NewCasRequestHandler(configuration Configuration, app http.Handler) (http.Handler, error) {
	casClientFactory, err := NewCasClientFactory(configuration)
	if err != nil {
		return nil, err
	}

	return &CasRequestHandler{
		CasBrowserHandler: casClientFactory.CreateClient().Handle(app),
		CasRestHandler:    casClientFactory.CreateRestClient().Handle(app),
	}, nil
}

type CasRequestHandler struct {
	CasBrowserHandler http.Handler
	CasRestHandler    http.Handler
}

func (h *CasRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := h.CasRestHandler
	if h.IsBrowserRequest(r) {
		handler = h.CasBrowserHandler
	}
	handler.ServeHTTP(w, r)
}

func (h *CasRequestHandler) IsBrowserRequest(r *http.Request) bool {
	return IsBrowserRequest(r.Header.Get("User-Agent"))
}
