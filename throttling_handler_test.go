package carp

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestThrottlingHandler(t *testing.T) {
	t.Run("Throttle too many requests", func(t *testing.T) {
		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
		}

		throttlingHandler := NewThrottlingHandler(Configuration{LimiterTokenRate: 1, LimiterBurstSize: 2}, handler)

		var ctxHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			request = request.WithContext(context.WithValue(request.Context(), _ServiceAccountAuthContextKey, true))
			throttlingHandler.ServeHTTP(writer, request)
		}

		server := httptest.NewServer(ctxHandler)

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

}
