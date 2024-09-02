package carp

import (
	"context"
	"net/http"
	"strings"
)

const BypassCasAuthContextKey = "BypassCasAuth"

func ShouldByPassCasAuthentication(r *http.Request) bool {
	shouldByPass, ok := r.Context().Value(BypassCasAuthContextKey).(bool)
	if !ok {
		return false
	}
	return shouldByPass
}

func NewDoguRestHandler(configuration Configuration, handler http.Handler) (http.HandlerFunc, error) {
	doguRestHandler := &restHandler{conf: configuration, wrappedHandler: handler}

	return doguRestHandler.handleRestRequest, nil
}

type restHandler struct {
	conf           Configuration
	wrappedHandler http.Handler
}

func (rh *restHandler) handleRestRequest(writer http.ResponseWriter, request *http.Request) {
	log.Infof("doguRestHandler: receiving request: %s", request.URL.String())

	ctx := request.Context()

	username, _, _ := request.BasicAuth()

	// TODO service-account-matching
	if !IsBrowserRequest(request) && rh.conf.ForwardUnauthenticatedRESTRequests && strings.HasPrefix("service_account_", username) {
		// This is a Rest-Request with a service-account-user -> it should bypass cas-authentication
		ctx = context.WithValue(request.Context(), BypassCasAuthContextKey, true)
	}

	// forward request to next handler with new context
	rh.wrappedHandler.ServeHTTP(writer, request.WithContext(ctx))
}
