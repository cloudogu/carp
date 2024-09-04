package carp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// NewServer creates a new carp server. Start the server with ListenAndServe()
func NewServer(configuration Configuration) (*http.Server, error) {
	mainHandler, err := createHandlersForConfig(configuration)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    ":" + strconv.Itoa(configuration.Port),
		Handler: mainHandler,
	}, nil
}

func createHandlersForConfig(configuration Configuration) (http.HandlerFunc, error) {
	proxyHandler, err := NewProxyHandler(configuration)
	if err != nil {
		return nil, fmt.Errorf("error creating proxy-handler: %w", err)
	}

	casRequestHandler, err := NewCasRequestHandler(configuration, proxyHandler)
	if err != nil {
		return nil, fmt.Errorf("error creating cas-request-handler: %w", err)
	}

	throttlingHandler := NewThrottlingHandler(context.TODO(), configuration, casRequestHandler)

	doguRestHandler, err := NewDoguRestHandler(configuration, throttlingHandler)
	if err != nil {
		return nil, fmt.Errorf("error creating dogu-rest-handler: %w", err)
	}

	return doguRestHandler, nil
}
