package carp

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProxyHandler(t *testing.T) {
	t.Run("should create proxy-handler", func(t *testing.T) {

		conf := Configuration{
			Target: "https://foo.bar/test",
		}
		ph, err := NewProxyHandler(conf)

		require.NoError(t, err)
		require.NotNil(t, ph)
		require.NotNil(t, ph.fwd)
		assert.Equal(t, conf, ph.config)
		assert.Equal(t, conf.Target, ph.target.String())
	})

	t.Run("should fail to create proxy-handler for error in target-url", func(t *testing.T) {

		conf := Configuration{
			Target: "http://example.com/%ZZ",
		}
		_, err := NewProxyHandler(conf)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to parse target-url:")
	})
}

func TestProxyHandler_ServeHTTP(t *testing.T) {
	t.Run("should handle unauthenticated rest request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		r.Header.Set("MY-PRINCIPAL", "MyUser")

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/foo/bar", r.URL.String())
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "", r.Header.Get("MY-PRINCIPAL"))

			w.WriteHeader(312)
		}))
		defer srv.Close()

		ph, err := NewProxyHandler(Configuration{
			ForwardUnauthenticatedRESTRequests: true,
			Target:                             srv.URL,
			PrincipalHeader:                    "MY-PRINCIPAL",
		})
		require.NoError(t, err)

		ph.ServeHTTP(w, r)

		assert.Equal(t, 312, w.Code)
	})

	t.Run("should handle unauthenticated browser-request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		r.Header.Set("User-Agent", "mozilla")

		logBuf := new(bytes.Buffer)
		logging.SetBackend(logging.NewLogBackend(logBuf, "", 0))

		rCounter := 1
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rCounter == 1 {
				assert.Equal(t, "/test/foo/bar", r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(200)
			} else if rCounter == 2 {
				assert.Equal(t, "/foo/bar", r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(204)
			} else {
				t.Errorf("unexpect request for url: %s", r.URL.String())
			}

			rCounter++
		}))
		defer srv.Close()

		ph, err := NewProxyHandler(Configuration{
			ForwardUnauthenticatedRESTRequests: true,
			Target:                             srv.URL,
			ResourcePath:                       "/foo/bar",
			BaseUrl:                            srv.URL + "/test",
		})
		require.NoError(t, err)

		ph.ServeHTTP(w, r)

		assert.Equal(t, 204, w.Code)
		assert.Contains(t, logBuf.String(), "Delivering resource /foo/bar on anonymous request...")
	})

	t.Run("should handle unauthenticated browser-request with redirect to login", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		r.Header.Set("User-Agent", "mozilla")

		logBuf := new(bytes.Buffer)
		logging.SetBackend(logging.NewLogBackend(logBuf, "", 0))

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test/foo/bar", r.URL.String())
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(401)
		}))
		defer srv.Close()

		ph, err := NewProxyHandler(Configuration{
			ForwardUnauthenticatedRESTRequests: true,
			Target:                             srv.URL,
			ResourcePath:                       "/foo/bar",
			BaseUrl:                            srv.URL + "/test",
		})
		require.NoError(t, err)

		ph.ServeHTTP(w, r)

		// 500 because there is no cas-client, but for this test it is ok
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, logBuf.String(), "Redirect resource-request /foo/bar to CAS...")
	})

	t.Run("should handle unauthenticated non-browser-request with redirect to login", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/test/foo/bar", r.URL.String())
			assert.Equal(t, http.MethodGet, r.Method)

			w.WriteHeader(401)
		}))
		defer srv.Close()

		logBuf := new(bytes.Buffer)
		logging.SetBackend(logging.NewLogBackend(logBuf, "", 0))

		ph, err := NewProxyHandler(Configuration{
			ForwardUnauthenticatedRESTRequests: false,
			Target:                             srv.URL,
			ResourcePath:                       "/foo/bar",
			BaseUrl:                            srv.URL + "/test",
		})
		require.NoError(t, err)

		ph.ServeHTTP(w, r)

		// 500 because there is no cas-client, but for this test it is ok
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, logBuf.String(), "Redirect request /foo/bar to CAS...")
	})
}

func TestProxyHandler_replicateUser(t *testing.T) {
	t.Run("should call user replicator if configured", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

		ph, err := NewProxyHandler(Configuration{
			UserReplicator: func(username string, attributes UserAttibutes) error {
				assert.Equal(t, "myUser", username)

				return nil
			},
		})
		require.NoError(t, err)

		err = ph.replicateUser(r, "myUser")

		require.NoError(t, err)
	})

	t.Run("should fail to call user replicator if error in replicator", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

		ph, err := NewProxyHandler(Configuration{
			UserReplicator: func(username string, attributes UserAttibutes) error {
				assert.Equal(t, "myUser", username)

				return assert.AnError
			},
		})
		require.NoError(t, err)

		err = ph.replicateUser(r, "myUser")

		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to replicate user: assert.AnError general error for testing")
	})

	t.Run("should not call user replicator if not configured", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

		ph, err := NewProxyHandler(Configuration{})
		require.NoError(t, err)

		err = ph.replicateUser(r, "myUser")

		require.NoError(t, err)
	})
}

func TestProxyHandler_handleAuthenticatedBrowserRequest(t *testing.T) {
	t.Run("should handle authenticated browser-request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/foo/bar", r.URL.String())
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "", r.Header.Get("MyUserHeader"))

			w.WriteHeader(200)
		}))
		defer srv.Close()

		logBuf := new(bytes.Buffer)
		logging.SetBackend(logging.NewLogBackend(logBuf, "", 0))

		ph, err := NewProxyHandler(Configuration{
			PrincipalHeader:                    "MyUserHeader",
			ForwardUnauthenticatedRESTRequests: false,
			Target:                             srv.URL,
		})
		require.NoError(t, err)

		ph.handleAuthenticatedBrowserRequest(w, r)

		// 500 because there is no cas-client, but for this test it is ok
		fmt.Printf("body: %s", w.Body.String())
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, logBuf.String(), "Forwarding request")
	})
}
