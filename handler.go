package carp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"strconv"

	"github.com/cloudogu/go-cas"
	"github.com/vulcand/oxy/forward"
)

const (
	handlerFactoryCasHandler  = "cas"
	handlerFactoryRestHandler = "rest"
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
	hdlFactory := createHandlerFactory()

	mainHandler, err := createMainRequestHandler(configuration, hdlFactory)
	if err != nil {
		return nil, err
	}
	casRequestHandler, err := NewCasRequestHandler(configuration, hdlFactory)
	if err != nil {
		return nil, err
	}

	doguRestHandler, err := NewDoguRestHandler(configuration, hdlFactory)
	if err != nil {
		return nil, err
	}

	hdlFactory.add(handlerFactoryCasHandler, casRequestHandler)
	hdlFactory.add(handlerFactoryRestHandler, doguRestHandler)

	return mainHandler, nil
}

func createMainRequestHandler(configuration Configuration, factory handlerFactory) (http.HandlerFunc, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to parse url: %s: %w", configuration.Target, err))
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(configuration.ResponseModifier))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create forward: %w", err))
	}

	return func(w http.ResponseWriter, req *http.Request) {
		statusWriter := &statusResponseWriter{ResponseWriter: w}
		username, _, _ := req.BasicAuth()

		if !cas.IsAuthenticated(req) {
			log.Infof("Found CAS-UNauthenticated request %s...", req.URL.String())

			if configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
				handleRestRequest(factory, statusWriter, req)
				return
			}

			handleUnauthenticatedBrowserRequest(configuration, req, statusWriter, target, fwd)
			return
		}

		handleAuthenticatedBrowerRequest(req, statusWriter, configuration, username, target, fwd)
	}, nil
}

func handleAuthenticatedBrowerRequest(req *http.Request, statusWriter *statusResponseWriter, configuration Configuration, username string, target *url.URL, fwd *forward.Forwarder) {
	log.Infof("Found CAS-authenticated request %s...", req.URL.String())

	if cas.IsFirstAuthenticatedRequest(req) {
		replicateUser(configuration, req, username)
	}
	req.Header.Set(configuration.PrincipalHeader, username)
	req.URL = target
	log.Infof("Forwarding request %s for user %s...", req.URL.String(), username)
	fwd.ServeHTTP(statusWriter, req)
}

func replicateUser(configuration Configuration, req *http.Request, username string) {
	if configuration.UserReplicator == nil {
		return
	}

	attributes := cas.Attributes(req)
	err := configuration.UserReplicator(username, UserAttibutes(attributes))
	if err != nil {
		log.Errorf("failed to replicate user: %s", err.Error())
		// try to continue with the request anyway...
	}
}

func handleUnauthenticatedBrowserRequest(conf Configuration, req *http.Request, statusWriter *statusResponseWriter, target *url.URL, fwd *forward.Forwarder) {
	resourcePath := conf.ResourcePath
	baseUrl := conf.BaseUrl
	isResourceRequestWithoutAuth := IsBrowserRequest(req) && resourcePath != "" && baseUrl != "" && isRequestToResource(req, resourcePath)
	if isResourceRequestWithoutAuth {
		response, err := http.Get(baseUrl + req.URL.String())
		if err != nil {
			log.Errorf("failed to request resource to test for status code: %s", err.Error())
		}

		if response.StatusCode >= 400 {
			// resource is unavailable
			// redirect not authenticated browser request to cas login page
			cas.RedirectToLogin(statusWriter, req)
			return
		}

		log.Infof("Delivering resource %s on anonymous request...", req.URL.String())
		req.URL = target
		fwd.ServeHTTP(statusWriter, req)
		return
	}

	// redirect the not-authenticated-browser-request to the CAS login page
	log.Infof("Redirect request %s to CAS...", req.URL.String())
	cas.RedirectToLogin(statusWriter, req)
	return
}

func handleRestRequest(factory handlerFactory, statusWriter *statusResponseWriter, req *http.Request) {
	restH, err := factory.get(handlerFactoryRestHandler)
	if err != nil {
		log.Errorf("failed to handle rest request: %s", err.Error())
		statusWriter.WriteHeader(http.StatusInternalServerError)
		_ = req.Body.Close()
		return
	}
	restH.ServeHTTP(statusWriter, req)
	return
}

func isRequestToResource(req *http.Request, resourcePath string) bool {
	return strings.Contains(req.URL.Path, resourcePath)
}
