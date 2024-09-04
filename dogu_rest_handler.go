package carp

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

const _ServiceAccountAuthContextKey = "ServiceAccountAuth"

func IsServiceAccountAuthentication(r *http.Request) bool {
	isServiceAccountAuth, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
	if !ok {
		return false
	}
	return isServiceAccountAuth
}

func NewDoguRestHandler(configuration Configuration, handler http.Handler) (http.HandlerFunc, error) {
	if configuration.ServiceAccountNameRegex == "" {
		log.Info("no ServiceAccountNameRegex configured. Not using doguRestHandler.")
		return handler.ServeHTTP, nil
	}

	usernameRegex, err := regexp.Compile(configuration.ServiceAccountNameRegex)
	if err != nil {
		return nil, fmt.Errorf("error compiling serviceAccountNameRegex: %w", err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		log.Debugf("doguRestHandler: receiving request: %s", request.URL.String())

		ctx := request.Context()

		username, _, ok := request.BasicAuth()

		if ok && configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(request) && usernameRegex.MatchString(username) {
			// This is a Rest-Request with a service-account-user -> set in context
			log.Debugf("doguRestHandler: request with username '%s' is a service-account-rest-request", username)
			ctx = context.WithValue(ctx, _ServiceAccountAuthContextKey, true)
		}

		// forward request to next handler with new context
		handler.ServeHTTP(writer, request.WithContext(ctx))
	}, nil
}
