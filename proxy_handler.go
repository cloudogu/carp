package carp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cloudogu/go-cas"
	"github.com/vulcand/oxy/forward"
)

type ProxyHandler struct {
	target *url.URL
	fwd    *forward.Forwarder
	config Configuration
}

func NewProxyHandler(configuration Configuration) (*ProxyHandler, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to parse target-url: %s: %w", configuration.Target, err))
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(configuration.ResponseModifier))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to create forward: %w", err))
	}

	return &ProxyHandler{
		config: configuration,
		target: target,
		fwd:    fwd,
	}, nil
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if cas.IsAuthenticated(req) {
		ph.handleAuthenticatedBrowserRequest(w, req)
		return
	}

	log.Debugf("Found unauthenticated request %s...", req.URL.String())

	if ph.config.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
		ph.handleRestRequest(w, req)
		return
	}

	ph.handleUnauthenticatedBrowserRequest(w, req)
}

func (ph *ProxyHandler) handleAuthenticatedBrowserRequest(w http.ResponseWriter, req *http.Request) {
	log.Debugf("Found CAS-authenticated request %s...", req.URL.String())

	username := cas.Username(req)
	if cas.IsFirstAuthenticatedRequest(req) {
		if err := ph.replicateUser(req, username); err != nil {
			log.Error(err.Error())
		}
	}
	req.Header.Set(ph.config.PrincipalHeader, username)
	req.URL = ph.target
	log.Infof("Forwarding request %s for user %s...", req.URL.String(), username)
	ph.fwd.ServeHTTP(w, req)
}

func (ph *ProxyHandler) replicateUser(req *http.Request, username string) error {
	if ph.config.UserReplicator == nil {
		return nil
	}

	attributes := cas.Attributes(req)
	err := ph.config.UserReplicator(username, UserAttibutes(attributes))
	if err != nil {
		return fmt.Errorf("failed to replicate user: %w", err)
	}

	return nil
}

func (ph *ProxyHandler) handleUnauthenticatedBrowserRequest(w http.ResponseWriter, req *http.Request) {
	resourcePath := ph.config.ResourcePath
	baseUrl := ph.config.BaseUrl
	isResourceRequestWithoutAuth := IsBrowserRequest(req) && resourcePath != "" && baseUrl != "" && isRequestToResource(req, resourcePath)
	if isResourceRequestWithoutAuth {
		response, err := http.Get(baseUrl + req.URL.String())
		if err != nil {
			log.Errorf("failed to request resource to test for status code: %s", err.Error())
		}

		if response.StatusCode >= 400 {
			// resource is unavailable
			// redirect not authenticated browser request to cas login page
			log.Debugf("Redirect resource-request %s to CAS...", req.URL.String())
			cas.RedirectToLogin(w, req)
			return
		}

		log.Infof("Delivering resource %s on anonymous request...", req.URL.String())
		req.URL = ph.target
		ph.fwd.ServeHTTP(w, req)
		return
	}

	// redirect the not-authenticated-browser-request to the CAS login page
	log.Infof("Redirect request %s to CAS...", req.URL.String())
	cas.RedirectToLogin(w, req)
	return
}

// forwards REST request for potential local user authentication
// remove rut auth header to prevent unwanted access if set
func (ph *ProxyHandler) handleRestRequest(w http.ResponseWriter, req *http.Request) {
	req.Header.Del(ph.config.PrincipalHeader)
	req.URL = ph.target
	ph.fwd.ServeHTTP(w, req)
}

func isRequestToResource(req *http.Request, resourcePath string) bool {
	return strings.Contains(req.URL.Path, resourcePath)
}
