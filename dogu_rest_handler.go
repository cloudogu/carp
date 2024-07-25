package carp

import (
	"github.com/pkg/errors"
	"github.com/vulcand/oxy/forward"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"sync"
)

var (
	mu      sync.RWMutex
	clients = make(map[string]*rate.Limiter)
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusResponseWriter) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	s.statusCode = statusCode
}

func NewDoguRestHandler(configuration Configuration, casHandler http.Handler) (http.HandlerFunc, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", configuration.Target)
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(configuration.ResponseModifier))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create forward")
	}

	forwardReqToDogu := func(writer http.ResponseWriter, request *http.Request) {
		username, _, _ := request.BasicAuth()
		log.Infof("forwarding request of user %s to dogu: %s", username, request.RequestURI)
		request.Header.Del(configuration.PrincipalHeader)
		request.URL = target
		fwd.ServeHTTP(writer, request)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		statusWriter := &statusResponseWriter{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		if IsBrowserRequest(request) || !configuration.ForwardUnauthenticatedRESTRequests {
			casHandler.ServeHTTP(writer, request)
			return
		}

		username, _, _ := request.BasicAuth()

		ip := request.Header.Get("X-Forwarded-For")

		for _, header := range request.Header {
			log.Infof("Header: %s:", header)
		}

		log.Infof("%s: found REST user %s and IP address %s", request.RequestURI, username, ip)
		if ip == "" {
			log.Warning("X-Forwarded-For header is not set, let CAS handle request...")
			casHandler.ServeHTTP(writer, request)

			return
		}

		limiter := getLimiter(ip)

		log.Infof("%s: user %s and IP address %s has %.2f tokens left", request.RequestURI, username, ip, limiter.Tokens())

		if !limiter.Allow() {
			log.Infof("%s: to many requests of user %s and IP address %s", request.RequestURI, username, ip)
			http.Error(writer, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		forwardReqToDogu(statusWriter, request)

		log.Infof("%s: request of %s was responded with status code %d", request.RequestURI, username, statusWriter.statusCode)
		if statusWriter.statusCode < 200 || statusWriter.statusCode >= 300 {
			casHandler.ServeHTTP(writer, request)
			return
		}

		cleanClient(ip)

	}, nil
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	l, ok := clients[ip]
	if !ok {
		l = rate.NewLimiter(50, 150)
		clients[ip] = l
	}

	return l
}

func cleanClient(ip string) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, ip)
}
