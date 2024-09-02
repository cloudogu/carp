package carp

import (
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
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

func NewThrottlingHandler(configuration Configuration, handler http.Handler) (http.HandlerFunc, error) {
	return func(writer http.ResponseWriter, request *http.Request) {
		username, _, _ := request.BasicAuth()

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

		// forward request to next handler with new context
		handler.ServeHTTP(writer, request)
	}, nil
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
