package carp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestThrottlingHandler(t *testing.T) {
	t.Run("Throttle too many requests", func(t *testing.T) {
		var handler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusUnauthorized)
		}

		throttlingHandler := NewThrottlingHandler(Configuration{LimiterTokenRate: 1, LimiterBurstSize: 1}, handler)
		server := httptest.NewServer(throttlingHandler)

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		req.Header.Set(_HttpHeaderXForwardedFor, "testIP")

		var found bool

		for i := 0; i < 5; i++ {
			resp, lErr := server.Client().Do(req)
			t.Log(i, resp.StatusCode)
			assert.NoError(t, lErr)
			if resp.StatusCode == http.StatusTooManyRequests {
				found = true
				break
			}
		}

		assert.True(t, found)
	})

}
