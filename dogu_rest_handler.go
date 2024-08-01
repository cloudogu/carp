package carp

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/vulcand/oxy/forward"
	"golang.org/x/time/rate"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	// TODO make token number customizable
	limiterDefaultTokensRate        = 50
	limiterDefaultTokensDuringBurst = 150
)

const httpHeaderXForwardedFor = "X-Forwarded-For"
const httpHeaderAuthorization = "Authorization"

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
		log.Infof("forwardToDoguHandler: forwarding request of user %s to dogu: %s", username, request.URL.String())
		request.Header.Del(configuration.PrincipalHeader)
		request.URL = target
		fwd.ServeHTTP(writer, request)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		log.Infof("doguRestHandler: receiving request: %s", request.URL.String())
		statusWriter := &statusResponseWriter{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		if IsBrowserRequest(request) || !configuration.ForwardUnauthenticatedRESTRequests {
			log.Infof("doguRestHandler: browser request identified: serving request with CAS handler: %s", request.URL.String())
			casHandler.ServeHTTP(writer, request)
			return
		}

		username, _, _ := request.BasicAuth()

		headers := strings.Builder{}
		for key, value := range request.Header {
			headers.WriteString(" ")
			headers.WriteString(key)
			headers.WriteString(": ")
			if key == httpHeaderAuthorization {
				splitUser, redactedPassword, ok := getUsernameFromAuthorizationHeader(value)
				if !ok {
					headers.WriteString("...")
					log.Debug("Splitting the basic auth user was unsuccessful... continuing")
					continue
				}
				headers.WriteString(splitUser)
				headers.WriteString(redactedPassword)
			} else {
				headers.WriteString(strings.Join(value, ", "))
			}
		}
		log.Debugf("Headers:%s:", headers.String())

		// go reverse proxy may add additional IP addresses from localhost. We need to take the right one.
		forwardedIpAddrRaw := request.Header.Get(httpHeaderXForwardedFor)
		forwardedIpAddresses := strings.Split(forwardedIpAddrRaw, ", ")
		initialForwardedIpAddress := ""
		if len(forwardedIpAddresses) > 0 {
			initialForwardedIpAddress = forwardedIpAddresses[0]
		}
		log.Infof("%s: found REST user %s and IP address %s", request.RequestURI, username, initialForwardedIpAddress)

		if initialForwardedIpAddress == "" {
			log.Infof("doguRestHandler: X-Forwarded-For header is not set: serving request with CAS handler: %s", request.URL.String())
			casHandler.ServeHTTP(writer, request)

			return
		}

		limiter := getLimiter(initialForwardedIpAddress)

		log.Infof("user %s and IP address %s has %.1f tokens left", username, initialForwardedIpAddress, limiter.Tokens())

		if !limiter.Allow() {
			log.Infof("too many requests of user %s and IP address %s", username, initialForwardedIpAddress)
			http.Error(writer, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		forwardReqToDogu(statusWriter, request)

		log.Infof("%s: request of %s responded with status code %d", request.RequestURI, username, statusWriter.statusCode)
		if statusWriter.statusCode < 200 || statusWriter.statusCode >= 300 {
			logCurrentToken(initialForwardedIpAddress, username, limiter.Tokens())
			log.Infof("doguRestHandler: statusWriter found HTTP %d: serving request with CAS handler: %s", statusWriter.statusCode, request.URL.String())
			casHandler.ServeHTTP(writer, request)
			// TODO this introduces a memory leak because some IPs never receive cleanClient() calls
			return
		}

		cleanClient(initialForwardedIpAddress, username)

	}, nil
}

func getUsernameFromAuthorizationHeader(value []string) (string, string, bool) {
	_, splitCreds, ok := strings.Cut(value[0], " ")
	if !ok {
		return "", "", false
	}
	decodedCreds, _ := base64.URLEncoding.DecodeString(splitCreds)
	splitUser, _, ok := strings.Cut(string(decodedCreds), ":")
	if !ok {
		return "", "", false
	}
	redactedPassword := ":***"
	return splitUser, redactedPassword, true
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	l, ok := clients[ip]
	if !ok {
		l = rate.NewLimiter(limiterDefaultTokensRate, limiterDefaultTokensDuringBurst)
		clients[ip] = l
	}

	return l
}

func logCurrentToken(ip, username string, token float64) {
	log.Infof("carp throttle: user %s and IP address %s with previously %.1f tokens left", username, ip, token)
}

func cleanClient(ip, username string) {
	logCurrentToken(ip, username, getLimiter(ip).Tokens())

	mu.Lock()
	defer mu.Unlock()

	delete(clients, ip)
}
