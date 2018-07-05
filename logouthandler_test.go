package carp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockDelegate struct{}

func (delegate MockDelegate) ServeHTTP(writer http.ResponseWriter, request *http.Request) {}

func TestShouldBypassNormalRequests(t *testing.T) {
	configuration := Configuration{}
	redirectionHandler := createSut(configuration, t)

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/x", strings.NewReader("")))

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestShouldRedirectForRequestMatchingMethod(t *testing.T) {
	configuration := Configuration{LogoutMethod: http.MethodDelete, CasUrl: "/cas"}
	redirectionHandler := createSut(configuration, t)

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodDelete, "/x", strings.NewReader("")))

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/cas/logout", recorder.HeaderMap.Get("Location"))
}

func TestShouldRedirectForRequestMatchingPath(t *testing.T) {
	configuration := Configuration{LogoutPath: "/quit", CasUrl: "/cas"}
	redirectionHandler := createSut(configuration, t)

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodPost, "/quit", strings.NewReader("")))

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/cas/logout", recorder.HeaderMap.Get("Location"))
}

func TestShouldRedirectForRequestMatchingPathAndMethod(t *testing.T) {
	configuration := Configuration{LogoutMethod: http.MethodDelete, LogoutPath: "/quit", CasUrl: "/cas"}
	redirectionHandler := createSut(configuration, t)

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodDelete, "/quit", strings.NewReader("")))

	assert.Equal(t, http.StatusSeeOther, recorder.Code)
	assert.Equal(t, "/cas/logout", recorder.HeaderMap.Get("Location"))
}

func TestShouldNotRedirectForRequestNotMatchingPath(t *testing.T) {
	configuration := Configuration{LogoutMethod: http.MethodDelete, LogoutPath: "/quit", CasUrl: "/cas"}
	redirectionHandler := createSut(configuration, t)

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodDelete, "/x", strings.NewReader("")))

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestShouldNotRedirectForRequestNotMatchingMethod(t *testing.T) {
	configuration := Configuration{LogoutMethod: http.MethodDelete, LogoutPath: "/quit", CasUrl: "/cas"}
	redirectionHandler := createSut(configuration, t)

	recorder := httptest.NewRecorder()
	redirectionHandler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodPost, "/quit", strings.NewReader("")))

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func createSut(configuration Configuration, t *testing.T) http.Handler {
	handler := MockDelegate{}
	return NewLogoutRedirectionHandler(configuration, handler)
}
