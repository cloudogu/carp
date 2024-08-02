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
	doguRestHandler, err := createHandlersForConfig(configuration)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    ":" + strconv.Itoa(configuration.Port),
		Handler: doguRestHandler,
	}, nil
}

func createHandlersForConfig(configuration Configuration) (http.HandlerFunc, error) {
	hdlFactory := createHandlerFactory(configuration)

	handler, err := createRequestHandler(configuration)
	if err != nil {
		return nil, err
	}
	casRequestHandler, err := NewCasRequestHandler(configuration)
	if err != nil {
		return nil, err
	}

	doguRestHandler, err := NewDoguRestHandler(configuration, casRequestHandler)
	if err != nil {
		return nil, err
	}

	hdlFactory.add(handlerFactoryCasHandler, casRequestHandler)
	hdlFactory.add(handlerFactoryRestHandler, doguRestHandler)

	return doguRestHandler, nil
}

type handlerFactory struct {
	conf     Configuration
	handlers map[string]http.Handler
}

func (f *handlerFactory) add(handlerId string, handler http.Handler) {
	switch handlerId {
	case handlerFactoryCasHandler:
		fallthrough
	case handlerFactoryRestHandler:
		f.handlers[handlerId] = handler
	default:
		panic("unknown request handler ID " + handlerId)
	}
}
func (f *handlerFactory) get(handlerId string) (http.Handler, error) {
	switch handlerId {
	case handlerFactoryCasHandler:
		fallthrough
	case handlerFactoryRestHandler:
		return f.handlers[handlerId], nil
	default:
		return nil, fmt.Errorf("unknown request handler ID " + handlerId)
	}
}

func createHandlerFactory(configuration Configuration) handlerFactory {
	return handlerFactory{
		conf:     configuration,
		handlers: make(map[string]http.Handler),
	}
}

func createRequestHandler(configuration Configuration) (http.HandlerFunc, error) {
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
		username := ""
		username, _, _ = req.BasicAuth()

		if !cas.IsAuthenticated(req) {
			log.Infof("Found CAS-UNauthenticated request %s...", req.URL.String())
			resourcePath := configuration.ResourcePath
			baseUrl := configuration.BaseUrl
			if configuration.ForwardUnauthenticatedRESTRequests && !IsBrowserRequest(req) {
				// forward REST request for potential local user authentication
				// remove rut auth header to prevent unwanted access if set
				req.Header.Del(configuration.PrincipalHeader)
				req.URL = target

				log.Infof("Forwarding rest request %s for user %s...", req.URL.String(), username)
				fwd.ServeHTTP(w, req)
			} else if IsBrowserRequest(req) && resourcePath != "" && baseUrl != "" && isRequestToResource(req, resourcePath) {
				response, err := http.Get(baseUrl + req.URL.String())
				if err != nil {
					log.Errorf("failed to request resource: %v", err)
				}
				if response.StatusCode >= 400 {
					// resource is unavailable
					// redirect not authenticated browser request to cas login page
					cas.RedirectToLogin(statusWriter, req)
				} else {
					log.Infof("Delivering resource %s on anonymous request...", req.URL.String())
					req.URL = target
					fwd.ServeHTTP(statusWriter, req)
				}
			} else {
				// redirect not authenticated browser request to cas login page
				log.Infof("Redirect request %s to CAS...", req.URL.String())
				cas.RedirectToLogin(statusWriter, req)
			}
			return
		}

		log.Infof("Found CAS-authenticated request %s...", req.URL.String())

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
		log.Infof("Forwarding request %s for user %s...", req.URL.String(), username)
		fwd.ServeHTTP(statusWriter, req)
	}, nil
}

func isRequestToResource(req *http.Request, resourcePath string) bool {
	return strings.Contains(req.URL.Path, resourcePath)
}
