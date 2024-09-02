package carp

import (
	"context"
	"net/http"
	"strings"
)

const _ServiceAccountAuthContextKey = "BypassCasAuth"

func IsServiceAccountAuthentication(r *http.Request) bool {
	isServiceAccountAuth, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
	if !ok {
		return false
	}
	return isServiceAccountAuth
}

func NewDoguRestHandler(configuration Configuration, handler http.Handler) (http.HandlerFunc, error) {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Infof("doguRestHandler: receiving request: %s", request.URL.String())

		ctx := request.Context()

		username, _, _ := request.BasicAuth()

		// TODO service-account-matching
		if !IsBrowserRequest(request) && configuration.ForwardUnauthenticatedRESTRequests && strings.HasPrefix("service_account_", username) {
			// This is a Rest-Request with a service-account-user -> it should bypass cas-authentication
			ctx = context.WithValue(ctx, _ServiceAccountAuthContextKey, true)
		}

		// forward request to next handler with new context
		handler.ServeHTTP(writer, request.WithContext(ctx))
	}, nil
}
