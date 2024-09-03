package carp

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewServer(t *testing.T) {
	t.Run("should create new server", func(t *testing.T) {
		srv, err := NewServer(Configuration{})

		require.NoError(t, err)
		require.NotNil(t, srv)
	})

	t.Run("should fail to create new server for error in proxy-handler", func(t *testing.T) {
		_, err := NewServer(Configuration{
			Target: "http://example.com/%ZZ",
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "error creating proxy-handler: failed to parse target-url")
	})

	t.Run("should fail to create new server for error in cas-handler", func(t *testing.T) {
		_, err := NewServer(Configuration{
			CasUrl: "http://example.com/%ZZ",
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "error creating cas-request-handler: failed to parse cas url")
	})

	t.Run("should fail to create new server for error in dogu-rest-handler", func(t *testing.T) {
		_, err := NewServer(Configuration{
			ServiceAccountNameRegex: "[",
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "error creating dogu-rest-handler: error compiling serviceAccountNameRegex")
	})
}
