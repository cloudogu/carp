package carp

import (
	"net/http"
	"net/url"
	"strings"

	"strconv"

	"github.com/cloudogu/go-cas"
	"github.com/pkg/errors"
	"github.com/vulcand/oxy/forward"
)

func NewServer(configuration Configuration) (*http.Server, error) {
	handler, err := createRequestHandler(configuration)
	if err != nil {
		return nil, err
	}

	casRequestHandler, err := NewCasRequestHandler(configuration, handler)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    ":" + strconv.Itoa(configuration.Port),
		Handler: casRequestHandler,
	}, nil
}

func createRequestHandler(configuration Configuration) (http.HandlerFunc, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", configuration.Target)
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(configuration.ResponseModifier))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create forward")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		if !cas.IsAuthenticated(req) {
			if configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
				// forward REST request for potential local user authentication
				// remove rut auth header to prevent unwanted access if set
				req.Header.Del(configuration.PrincipalHeader)
				req.URL = target
				fwd.ServeHTTP(w, req)
			} else if IsBrowserRequest(req) && isRequestToResource(req, configuration.ResourcePath) {
				response, err := http.Get(configuration.BaseUrl + req.URL.String())
				if err != nil {
					log.Errorf("failed to request resource: %v", err)
				}
				if response.StatusCode >= 400 {
					// resource is unavailable
					// redirect not authenticated browser request to cas login page
					cas.RedirectToLogin(w, req)
				} else {
					log.Infof("Delivering resource %s on anonymous request...", req.URL.String())
					req.URL = target
					fwd.ServeHTTP(w, req)
				}
			} else {
				// redirect not authenticated browser request to cas login page
				cas.RedirectToLogin(w, req)
			}
			return
		}
		username := cas.Username(req)
		if cas.IsFirstAuthenticatedRequest(req) {
			if configuration.UserReplicator != nil {
				attributes := cas.Attributes(req)
				err := configuration.UserReplicator(username, UserAttibutes(attributes))
				if err != nil {
					log.Errorf("failed to replicate user: %v", err)
				}
			}
		}
		req.Header.Set(configuration.PrincipalHeader, username)
		req.URL = target
		fwd.ServeHTTP(w, req)
	}, nil
}

func isRequestToResource(req *http.Request, resourcePath string) bool {
	return strings.Contains(req.URL.Path, resourcePath)
}
