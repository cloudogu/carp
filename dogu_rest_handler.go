package carp

import (
	"fmt"
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
	// wroteHeader returns true if any write method from ResponseWriter was called previously.
	wroteHeader bool
}

func (s *statusResponseWriter) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	s.statusCode = statusCode
	s.wroteHeader = true
}
func (w *statusResponseWriter) Write(b []byte) (int, error) {
	// comply with the default behaviour of net.ResponseWriter to default to HTTP 200 on the first write call without
	// previously setting a header
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	w.wroteHeader = true
	return w.ResponseWriter.Write(b)
}

type restHandler struct {
	conf           Configuration
	handlerFactory handlerFactory
}

func NewDoguRestHandler(configuration Configuration, factory handlerFactory) (http.HandlerFunc, error) {
	doguRestHandler := &restHandler{conf: configuration, handlerFactory: factory}

	return doguRestHandler.handleRestRequest, nil
}

func (rh *restHandler) handleRestRequest(writer http.ResponseWriter, request *http.Request) {
	log.Infof("doguRestHandler: receiving request: %s", request.URL.String())
	statusWriter := &statusResponseWriter{
		ResponseWriter: writer,
		statusCode:     http.StatusOK,
	}

	if IsBrowserRequest(request) || !rh.conf.ForwardUnauthenticatedRESTRequests {
		log.Infof("doguRestHandler: browser request %s identified: cannot handle this, falling back to CAS handling", request.URL.String())
		cas, err := rh.handlerFactory.get(handlerFactoryCasHandler)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		cas.ServeHTTP(writer, request)
		return
	}

	username, _, _ := request.BasicAuth()

	logDebugRequestHeaders(request)

	// go reverse proxy may add additional IP addresses from localhost. We need to take the right one.
	forwardedIpAddrRaw := request.Header.Get(httpHeaderXForwardedFor)
	forwardedIpAddresses := strings.Split(forwardedIpAddrRaw, ", ")
	initialForwardedIpAddress := ""
	if len(forwardedIpAddresses) > 0 {
		initialForwardedIpAddress = forwardedIpAddresses[0]
	}
	log.Infof("%s: found REST user %s and IP address %s", request.RequestURI, username, initialForwardedIpAddress)

	var limiter *rate.Limiter
	if initialForwardedIpAddress == "" {
		log.Infof("doguRestHandler: X-Forwarded-For header is not set (might be an internal account): %s", request.URL.String())
		if username == "" {
			log.Warningf("doguRestHandler: username not found in X-Forwarded-For header")
		}
	}

	ipUsernameId := fmt.Sprintf("%s:%s", initialForwardedIpAddress, username)
	limiter = getOrCreateLimiter(ipUsernameId)

	log.Infof("doguRestHandler: user %s and IP address %s has %.1f tokens left", username, initialForwardedIpAddress, limiter.Tokens())

	if !limiter.Allow() {
		log.Infof("too many requests of user %s and IP address %s", username, initialForwardedIpAddress)
		http.Error(writer, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	rh.forwardReqToDogu()(statusWriter, request)

	log.Infof("%s: request of %s responded with status code %d", request.RequestURI, username, statusWriter.statusCode)
	if statusWriter.statusCode < 200 || statusWriter.statusCode >= 300 {
		logCurrentToken(initialForwardedIpAddress, username, limiter.Tokens())
		log.Infof("doguRestHandler: statusWriter found HTTP %d: serving request with CAS handler: %s", statusWriter.statusCode, request.URL.String())
		casHandler.ServeHTTP(writer, request)
		// TODO this introduces a memory leak because some IPs never receive cleanClient() calls
		return
	}

	cleanClient(initialForwardedIpAddress, username)
}

func (rh *restHandler) forwardReqToDogu() func(writer http.ResponseWriter, request *http.Request) {
	target, err := url.Parse(rh.conf.Target)
	if err != nil {
		return nil
	}

	fwd, err := forward.New(forward.PassHostHeader(true), forward.ResponseModifier(rh.conf.ResponseModifier))
	if err != nil {
		return nil
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		username, _, _ := request.BasicAuth()
		log.Infof("forwardToDoguHandler: forwarding request of user %s to dogu: %s", username, request.URL.String())
		request.Header.Del(rh.conf.PrincipalHeader)
		request.URL = target
		fwd.ServeHTTP(writer, request)
	}
}

// logDebugRequestHeaders prints all request headers except a basic auth password during a log level of debug.
func logDebugRequestHeaders(request *http.Request) {
	headers := strings.Builder{}
	for key, value := range request.Header {
		headers.WriteString(" ")
		headers.WriteString(key)
		headers.WriteString(": ")
		if key == httpHeaderAuthorization {
			splitUser, redactedPassword, ok := getRedactedCredentialsFromAuthHeader(value)
			if !ok {
				headers.WriteString("...")
				log.Debug("Splitting the basic auth user was unsuccessful... continuing")
				continue
			}
			headers.WriteString(splitUser)
			headers.WriteString(":")
			headers.WriteString(redactedPassword)
			continue
		}

		headers.WriteString(strings.Join(value, ", "))
	}

	log.Debugf("Headers:%s:", headers.String())
}

func getRedactedCredentialsFromAuthHeader(value []string) (string, string, bool) {
	if !strings.HasPrefix("Basic ", value[0]) {
		return "", "", false
	}

	// creating a request is kind of a hack but is more reliable than decoding basic auth creds manually
	req := http.Request{Header: map[string][]string{httpHeaderAuthorization: value}}
	user, _, ok := req.BasicAuth()

	redactedPassword := "***"
	return user, redactedPassword, ok
}

func getOrCreateLimiter(ip string) *rate.Limiter {
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
	logCurrentToken(ip, username, getOrCreateLimiter(ip).Tokens())

	mu.Lock()
	defer mu.Unlock()

	delete(clients, ip)
}
