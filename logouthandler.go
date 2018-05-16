package carp

import (
	"net/http"
)

type RedirectingHandler struct {
	logoutUrl string
	delegate http.Handler
}

func (h *RedirectingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		http.Redirect(w, r, h.logoutUrl, 303)
		return
	}
	h.delegate.ServeHTTP(w, r)
}
