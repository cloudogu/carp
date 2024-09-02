package carp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/op/go-logging"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDoguRestHandler(t *testing.T) {
	t.Run("should set context value in request if username matches", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isServiceAccount, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			require.True(t, ok)
			assert.True(t, isServiceAccount)
		})

		req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("service_account_BASELINE_aBcDeF:myPassword"))))
		w := httptest.NewRecorder()

		config := Configuration{
			ServiceAccountNameRegex:            "^service_account_([A-Za-z0-9]+)_([A-Za-z0-9]+)$",
			ForwardUnauthenticatedRESTRequests: true,
		}
		doguRestHandler, err := NewDoguRestHandler(config, nextHandler)
		require.NoError(t, err)

		doguRestHandler.ServeHTTP(w, req)
	})

	t.Run("should not set context value in request if username does not match", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			assert.False(t, ok)
		})

		req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("myUser:myPassword"))))
		w := httptest.NewRecorder()

		config := Configuration{
			ServiceAccountNameRegex:            "^service_account_([A-Za-z0-9]+)_([A-Za-z0-9]+)$",
			ForwardUnauthenticatedRESTRequests: true,
		}
		doguRestHandler, err := NewDoguRestHandler(config, nextHandler)
		require.NoError(t, err)

		doguRestHandler.ServeHTTP(w, req)
	})

	t.Run("should not set context value in request if request is browser-request", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			assert.False(t, ok)
		})

		req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("service_account_BASELINE_aBcDeF:myPassword"))))
		req.Header.Set("User-Agent", "mozilla")
		w := httptest.NewRecorder()

		config := Configuration{
			ServiceAccountNameRegex:            "^service_account_([A-Za-z0-9]+)_([A-Za-z0-9]+)$",
			ForwardUnauthenticatedRESTRequests: true,
		}
		doguRestHandler, err := NewDoguRestHandler(config, nextHandler)
		require.NoError(t, err)

		doguRestHandler.ServeHTTP(w, req)
	})

	t.Run("should not set context value in request if request is not forwarding unauthenticated rest-request", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			assert.False(t, ok)
		})

		req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("service_account_BASELINE_aBcDeF:myPassword"))))
		w := httptest.NewRecorder()

		config := Configuration{
			ServiceAccountNameRegex:            "^service_account_([A-Za-z0-9]+)_([A-Za-z0-9]+)$",
			ForwardUnauthenticatedRESTRequests: false,
		}
		doguRestHandler, err := NewDoguRestHandler(config, nextHandler)
		require.NoError(t, err)

		doguRestHandler.ServeHTTP(w, req)
	})

	t.Run("should not set context value in request if no basic-auth provided", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			assert.False(t, ok)
		})

		req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
		w := httptest.NewRecorder()

		config := Configuration{
			ServiceAccountNameRegex:            "^service_account_([A-Za-z0-9]+)_([A-Za-z0-9]+)$",
			ForwardUnauthenticatedRESTRequests: true,
		}
		doguRestHandler, err := NewDoguRestHandler(config, nextHandler)
		require.NoError(t, err)

		doguRestHandler.ServeHTTP(w, req)
	})

	t.Run("should fail to create handler with error in regex", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			assert.False(t, ok)
		})

		config := Configuration{
			ServiceAccountNameRegex:            "[",
			ForwardUnauthenticatedRESTRequests: true,
		}
		_, err := NewDoguRestHandler(config, nextHandler)

		require.Error(t, err)
		assert.ErrorContains(t, err, "error compiling serviceAccountNameRegex: error parsing regexp: missing closing ]")
	})

	t.Run("should not use doguRestHandler if no regex is configured", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := r.Context().Value(_ServiceAccountAuthContextKey).(bool)
			assert.False(t, ok)
		})

		logBuf := new(bytes.Buffer)
		logging.SetBackend(logging.NewLogBackend(logBuf, "", 0))

		config := Configuration{}
		doguRestHandler, err := NewDoguRestHandler(config, nextHandler)
		require.NoError(t, err)

		assert.NotNil(t, doguRestHandler)
		assert.Contains(t, logBuf.String(), "no ServiceAccountNameRegex configured. Not using doguRestHandler.")
	})
}

func TestIsServiceAccountAuthentication(t *testing.T) {
	tests := []struct {
		name         string
		contextKey   string
		contextValue interface{}
		want         bool
	}{
		{
			"should return true",
			_ServiceAccountAuthContextKey,
			true,
			true,
		},
		{
			"should return false",
			_ServiceAccountAuthContextKey,
			false,
			false,
		},
		{
			"should return false for wrong type",
			_ServiceAccountAuthContextKey,
			"this no bool",
			false,
		},
		{
			"should return false for missing key",
			"FooKey",
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), tt.contextKey, tt.contextValue)
			req := httptest.NewRequest(http.MethodGet, "/foo", nil).WithContext(ctx)
			assert.Equalf(t, tt.want, IsServiceAccountAuthentication(req), "IsServiceAccountAuthentication(%v)", req)
		})
	}
}
