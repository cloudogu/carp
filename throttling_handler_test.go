package carp

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestThrottlingHandler(t *testing.T) {
	limiterConfig := Configuration{LimiterTokenRate: 1, LimiterBurstSize: 2}

	cleanUp := func(server *httptest.Server) {
		server.Close()
		clients = make(map[string]*rate.Limiter)
	}

	t.Run("Throttle too many requests in short time", func(t *testing.T) {
		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
		}

		throttlingHandler := NewThrottlingHandler(limiterConfig, handler)

		var ctxHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(context.WithValue(request.Context(), _ServiceAccountAuthContextKey, true))
			throttlingHandler.ServeHTTP(writer, request)
		}

		server := httptest.NewServer(ctxHandler)
		defer cleanUp(server)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		req.Header.Set(_HttpHeaderXForwardedFor, "testIP")
		req.SetBasicAuth("test", "test")

		var found bool

		for i := 0; i < 5; i++ {
			resp, lErr := server.Client().Do(req)
			assert.NoError(t, lErr)

			t.Log(i, resp.StatusCode)
			if resp.StatusCode == http.StatusTooManyRequests {
				found = true
				break
			}
		}

		assert.True(t, found)
	})

	t.Run("Only throttle service accounts", func(t *testing.T) {
		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}

		throttlingHandler := NewThrottlingHandler(limiterConfig, handler)

		server := httptest.NewServer(throttlingHandler)
		defer cleanUp(server)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		req.Header.Set(_HttpHeaderXForwardedFor, "testIP")
		req.SetBasicAuth("test", "test")

		for i := 0; i < 5; i++ {
			resp, lErr := server.Client().Do(req)
			assert.NoError(t, lErr)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

		}
	})

	t.Run("Return error when invalid BasicAuth is provided", func(t *testing.T) {
		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
		}

		throttlingHandler := NewThrottlingHandler(limiterConfig, handler)

		var ctxHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(context.WithValue(request.Context(), _ServiceAccountAuthContextKey, true))
			throttlingHandler.ServeHTTP(writer, request)
		}

		server := httptest.NewServer(ctxHandler)
		defer cleanUp(server)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		req.Header.Set(_HttpHeaderXForwardedFor, "testIP")

		resp, lErr := server.Client().Do(req)
		assert.NoError(t, lErr)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Refresh Tokens after throttling", func(t *testing.T) {
		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
		}

		throttlingHandler := NewThrottlingHandler(limiterConfig, handler)

		var ctxHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(context.WithValue(request.Context(), _ServiceAccountAuthContextKey, true))
			throttlingHandler.ServeHTTP(writer, request)
		}

		server := httptest.NewServer(ctxHandler)
		defer cleanUp(server)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		req.Header.Set(_HttpHeaderXForwardedFor, "testIP")
		req.SetBasicAuth("test", "test")

		clientCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		// Using the same limiter config for client means, server can refresh tokens
		clientLimiter := rate.NewLimiter(rate.Limit(limiterConfig.LimiterTokenRate), limiterConfig.LimiterBurstSize)

		for i := 0; i < 5; i++ {
			lErr := clientLimiter.Wait(clientCtx)
			assert.NoError(t, lErr)

			resp, lErr := server.Client().Do(req)
			assert.NoError(t, lErr)

			t.Log(i, resp.StatusCode)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("Reset throttling after successful try", func(t *testing.T) {
		reqCounter := 0

		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			if reqCounter == (limiterConfig.LimiterBurstSize - 1) {
				writer.WriteHeader(http.StatusOK)
				reqCounter = 0
				return
			}

			writer.WriteHeader(http.StatusUnauthorized)
			reqCounter++
		}

		throttlingHandler := NewThrottlingHandler(limiterConfig, handler)

		var ctxHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(context.WithValue(request.Context(), _ServiceAccountAuthContextKey, true))
			throttlingHandler.ServeHTTP(writer, request)
		}

		server := httptest.NewServer(ctxHandler)
		defer cleanUp(server)

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		req.Header.Set(_HttpHeaderXForwardedFor, "testIP")
		req.SetBasicAuth("test", "test")

		var found bool

		for i := 0; i < 10; i++ {
			resp, lErr := server.Client().Do(req)
			assert.NoError(t, lErr)

			t.Log(i, resp.StatusCode)
			if resp.StatusCode == http.StatusTooManyRequests {
				found = true
				break
			}
		}

		assert.False(t, found)
	})
}
