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
	log.Debug("Entering Method 'NewServer'")
	defer func() {
		log.Debug("End of Function 'NewServer'")
	}()

	log.Debugf("Param '%s'", configuration)
	handler, err := createRequestHandler(configuration)
	log.Debugf("Variable: %s", handler)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, err
	}

	casRequestHandler, err := NewCasRequestHandler(configuration, handler)
	log.Debugf("Variable: %s", casRequestHandler)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, err
	}

	addr := ":" + strconv.Itoa(configuration.Port)
	log.Debugf("Variable: %s", addr)
	log.Debugf("Variable: %s", casRequestHandler)
	return &http.Server{
		Addr:    addr,
		Handler: casRequestHandler,
	}, nil
}

func createRequestHandler(configuration Configuration) (http.HandlerFunc, error) {
	log.Debug("Entering Method 'createRequestHandler'")
	defer func() {
		log.Debug("End of Function 'createRequestHandler'")
	}()

	log.Debugf("Param '%s'", configuration)
	target, err := url.Parse(configuration.Target)
	log.Debugf("Variable: %s", target)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, errors.Wrapf(err, "failed to parse url: %s", configuration.Target)
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(configuration.ResponseModifier))
	log.Debugf("Variable: %s", fwd)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, errors.Wrap(err, "failed to create forward")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		log.Debug("Entering Method 'func(w http.ResponseWriter, req *http.Request)'")
		defer func() {
			log.Debug("End of Function 'func(w http.ResponseWriter, req *http.Request)'")
		}()

		if !cas.IsAuthenticated(req) {
			log.Debugf("Condition true: '!cas.IsAuthenticated(req)'")
			resourcePath := configuration.ResourcePath
			log.Debugf("Variable: %s", resourcePath)
			baseUrl := configuration.BaseUrl
			log.Debugf("Variable: %s", baseUrl)
			if configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
				log.Debugf("Condition true: 'configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req)'")
				// forward REST request for potential local user authentication
				// remove rut auth header to prevent unwanted access if set
				log.Debugf("Variable: %s", req.Header)
				req.Header.Del(configuration.PrincipalHeader)
				log.Debugf("Variable: %s", req.Header)

				req.URL = target
				log.Debugf("Variable: %s", req.URL)
				log.Debugf("Variable: %s", fwd)
				fwd.ServeHTTP(w, req)
			} else if IsBrowserRequest(req) && resourcePath != "" && baseUrl != "" && isRequestToResource(req, resourcePath) {
				log.Debugf("Condition true: 'IsBrowserRequest(req) && resourcePath != \"\" && baseUrl != \"\" && isRequestToResource(req, resourcePath)'")
				response, err := http.Get(baseUrl + req.URL.String())
				log.Debugf("Variable: %s", response)
				if err != nil {
					log.Debugf("Error: %s", err.Error())
					log.Errorf("failed to request resource: %v", err)
				}
				if response.StatusCode >= 400 {
					log.Debugf("Condition true: 'response.StatusCode >= 400'")
					// resource is unavailable
					// redirect not authenticated browser request to cas login page
					log.Debugf("REDIRECTING 2")
					cas.RedirectToLogin(w, req)
				} else {
					log.Debugf("Condition true: 'else'")
					log.Infof("Delivering resource %s on anonymous request...", req.URL.String())
					req.URL = target
					log.Debugf("Variable: %s", req.URL)
					fwd.ServeHTTP(w, req)
				}
			} else {
				log.Debugf("Condition true: 'else'")
				// redirect not authenticated browser request to cas login page
				log.Debugf("REDIRECTING 1")
				cas.RedirectToLogin(w, req)
			}
			return
		}
		username := cas.Username(req)
		log.Debugf("Variable: %s", username)
		if cas.IsFirstAuthenticatedRequest(req) {
			log.Debugf("Condition true: 'cas.IsFirstAuthenticatedRequest(req)'")
			if configuration.UserReplicator != nil {
				log.Debugf("Condition true: 'configuration.UserReplicator != nil'")
				attributes := cas.Attributes(req)
				log.Debugf("Variable: %s", attributes)
				err := configuration.UserReplicator(username, UserAttibutes(attributes))
				if err != nil {
					log.Debugf("Error: %s", err.Error())
					log.Errorf("failed to replicate user: %v", err)
				}
			}
		}
		req.Header.Set(configuration.PrincipalHeader, username)
		log.Debugf("Variable: %s", req.Header)
		req.URL = target
		log.Debugf("Variable: %s", req.URL)
		fwd.ServeHTTP(w, req)
	}, nil
}

func isRequestToResource(req *http.Request, resourcePath string) bool {
	log.Debug("Entering Method 'isRequestToResource'")
	defer func() {
		log.Debug("End of Function 'isRequestToResource'")
	}()

	log.Debugf("Param '%s'", req)
	log.Debugf("Param '%s'", resourcePath)
	return strings.Contains(req.URL.Path, resourcePath)
}
