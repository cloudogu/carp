package carp

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/stretchr/testify/assert"
)

func TestNoRedirectionConfigured(t *testing.T) {
    req, _ := http.NewRequest(http.MethodGet, "/x", nil)
    req.Header.Set("User-Agent", "mozilla")

    recorder := httptest.NewRecorder()
	requestHandler, _ := NewCasRequestHandler(Configuration{}, MockDelegate{})

    requestHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRedirectionConfiguredWithMethod(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, "/x", nil)
	req.Header.Set("User-Agent", "mozilla")

	recorder := httptest.NewRecorder()
	requestHandler, _ := NewCasRequestHandler(
		Configuration{LogoutMethod: http.MethodDelete, CasUrl: "/cas"}, MockDelegate{})

	requestHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
}

func TestRedirectionConfiguredWithPath(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
	req.Header.Set("User-Agent", "mozilla")

	recorder := httptest.NewRecorder()
	requestHandler, _ := NewCasRequestHandler(
		Configuration{LogoutPath: "/logout", CasUrl: "/cas"}, MockDelegate{})

	requestHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
}

func TestNoRedirectionForRestThoughRedirectionConfigured(t *testing.T) {
	req, _ := http.NewRequest(http.MethodDelete, "/x", nil)

	recorder := httptest.NewRecorder()
	requestHandler, _ := NewCasRequestHandler(
		Configuration{LogoutMethod: "DELETE", CasUrl: "/cas"}, MockDelegate{})

	requestHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
