package carp

import (
	"net/http"
	"net/url"

	"strconv"

	"github.com/cloudogu/go-cas"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/vulcand/oxy/forward"
)

func NewServer(configuration Configuration, forwardUnauthenticatedRESTRequests bool) (*http.Server, error) {
	handler, err := createRequestHandler(configuration, forwardUnauthenticatedRESTRequests)
	if err != nil {
		return nil, err
	}

	casRequestHandler, err := NewCasRequestHandler(configuration, handler, forwardUnauthenticatedRESTRequests)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    ":" + strconv.Itoa(configuration.Port),
		Handler: casRequestHandler,
	}, nil
}

func createRequestHandler(configuration Configuration, forwardUnauthenticatedRESTRequests bool) (http.HandlerFunc, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", configuration.Target)
	}

	fwd, err := forward.New(forward.PassHostHeader(true))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create forward")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		if !cas.IsAuthenticated(req) {
			if forwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
				// forward REST req for potential local user authentication
				req.URL = target
				fwd.ServeHTTP(w, req)
			} else {
				// redirect not authenticated browser req to cas login page
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
					glog.Errorf("failed to replicate user: %v", err)
				}
			}
		}
		req.Header.Set(configuration.PrincipalHeader, username)
		req.URL = target
		fwd.ServeHTTP(w, req)
	}, nil
}
