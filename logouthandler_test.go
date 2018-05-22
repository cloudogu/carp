package carp

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
)

type MockDelegate struct { }

func (delegate MockDelegate) ServeHTTP(writer http.ResponseWriter, request *http.Request) {}

func TestShouldBypassNormalRequests(t *testing.T) {
	handler := MockDelegate{}

	redirectionHandler, e := NewLogoutRedirectionHandler(Configuration{}, handler)
	if e != nil {
		assert.Fail(t, "got unexpected error")
	}

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("GET", "/x", strings.NewReader("")))

	assert.Equal(t, 200, recorder.Code)
}

func TestShouldRedirectForRequestMatchingMethod(t *testing.T) {
	handler := MockDelegate{}

	redirectionHandler, e := NewLogoutRedirectionHandler(Configuration{LogoutMethod: "DELETE", CasUrl: "/cas"}, handler)
	if e != nil {
		assert.Fail(t, "got unexpected error")
	}

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("DELETE", "/x", strings.NewReader("")))

	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/cas/logout", recorder.HeaderMap.Get("Location"))
}

func TestShouldRedirectForRequestMatchingPath(t *testing.T) {
	handler := MockDelegate{}

	redirectionHandler, e := NewLogoutRedirectionHandler(Configuration{LogoutPath: "/quit", CasUrl: "/cas"}, handler)
	if e != nil {
		assert.Fail(t, "got unexpected error")
	}

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("POST", "/quit", strings.NewReader("")))

	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/cas/logout", recorder.HeaderMap.Get("Location"))
}

func TestShouldRedirectForRequestMatchingPathAndMethod(t *testing.T) {
	handler := MockDelegate{}

	redirectionHandler, e := NewLogoutRedirectionHandler(Configuration{LogoutMethod: "DELETE", LogoutPath: "/quit", CasUrl: "/cas"}, handler)
	if e != nil {
		assert.Fail(t, "got unexpected error")
	}

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("DELETE", "/quit", strings.NewReader("")))

	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/cas/logout", recorder.HeaderMap.Get("Location"))
}

func TestShouldNotRedirectForRequestNotMatchingPath(t *testing.T) {
	handler := MockDelegate{}

	redirectionHandler, e := NewLogoutRedirectionHandler(Configuration{LogoutMethod: "DELETE", LogoutPath: "/quit", CasUrl: "/cas"}, handler)
	if e != nil {
		assert.Fail(t, "got unexpected error")
	}

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("DELETE", "/x", strings.NewReader("")))

	assert.Equal(t, 200, recorder.Code)
}

func TestShouldNotRedirectForRequestNotMatchingMethod(t *testing.T) {
	handler := MockDelegate{}

	redirectionHandler, e := NewLogoutRedirectionHandler(Configuration{LogoutMethod: "DELETE", LogoutPath: "/quit", CasUrl: "/cas"}, handler)
	if e != nil {
		assert.Fail(t, "got unexpected error")
	}

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("POST", "/quit", strings.NewReader("")))

	assert.Equal(t, 200, recorder.Code)
}
