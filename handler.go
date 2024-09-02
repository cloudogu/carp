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
	mainHandler, err := createMainRequestHandler(configuration)
	if err != nil {
		return nil, err
	}

	casRequestHandler, err := NewCasRequestHandler(configuration, mainHandler)
	if err != nil {
		return nil, err
	}

	throttlingHandler := NewThrottlingHandler(configuration, casRequestHandler)
	if err != nil {
		return nil, err
	}

	doguRestHandler, err := NewDoguRestHandler(configuration, throttlingHandler)
	if err != nil {
		return nil, err
	}

	return doguRestHandler, nil
}

func createMainRequestHandler(configuration Configuration) (http.HandlerFunc, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to parse url: %s: %w", configuration.Target, err))
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(configuration.ResponseModifier))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create forward: %w", err))
	}

	return func(w http.ResponseWriter, req *http.Request) {
		username, _, _ := req.BasicAuth()

		if !cas.IsAuthenticated(req) {
			log.Infof("Found CAS-UNauthenticated request %s...", req.URL.String())

			if configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
				handleRestRequest(w, req, target, fwd, configuration.PrincipalHeader)
				return
			}

			handleUnauthenticatedBrowserRequest(configuration, req, w, target, fwd)
			return
		}

		handleAuthenticatedBrowerRequest(req, w, configuration, username, target, fwd)
	}, nil
}

func handleAuthenticatedBrowerRequest(req *http.Request, w http.ResponseWriter, configuration Configuration, username string, target *url.URL, fwd *forward.Forwarder) {
	log.Infof("Found CAS-authenticated request %s...", req.URL.String())

	if cas.IsFirstAuthenticatedRequest(req) {
		replicateUser(configuration, req, username)
	}
	req.Header.Set(configuration.PrincipalHeader, username)
	req.URL = target
	log.Infof("Forwarding request %s for user %s...", req.URL.String(), username)
	fwd.ServeHTTP(w, req)
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

func handleUnauthenticatedBrowserRequest(conf Configuration, req *http.Request, w http.ResponseWriter, target *url.URL, fwd *forward.Forwarder) {
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
			cas.RedirectToLogin(w, req)
			return
		}

		log.Infof("Delivering resource %s on anonymous request...", req.URL.String())
		req.URL = target
		fwd.ServeHTTP(w, req)
		return
	}

	// redirect the not-authenticated-browser-request to the CAS login page
	log.Infof("Redirect request %s to CAS...", req.URL.String())
	cas.RedirectToLogin(w, req)
	return
}

// forwards REST request for potential local user authentication
// remove rut auth header to prevent unwanted access if set
func handleRestRequest(w http.ResponseWriter, req *http.Request, target *url.URL, fwd *forward.Forwarder, principalHeader string) {
	req.Header.Del(principalHeader)
	req.URL = target
	fwd.ServeHTTP(w, req)
}

func isRequestToResource(req *http.Request, resourcePath string) bool {
	return strings.Contains(req.URL.Path, resourcePath)
}
