package carp

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDoguRestHandler(t *testing.T) {
	var casHandler http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
	}

	doguServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusForbidden)
	}))

	restDoguHander, err := NewDoguRestHandler(Configuration{ForwardUnauthenticatedRESTRequests: true, Target: doguServer.URL}, casHandler)
	require.NoError(t, err)

	server := httptest.NewServer(restDoguHander)
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	req.Header.Set("X-Forwarded-For", "testIP")

	var found bool

	for i := 0; i < 200; i++ {
		resp, lErr := server.Client().Do(req)
		t.Log(i, resp.StatusCode)
		assert.NoError(t, lErr)
		if resp.StatusCode == http.StatusTooManyRequests {
			found = true
			break
		}
	}

	assert.True(t, found)
}
