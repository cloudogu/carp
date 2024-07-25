package carp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nexusLocalUserAuthentication = "bG9jYWxVc2VybmFtZTpwYXNzd29yZA=="
)

func TestNewDoguRestHandler(t *testing.T) {
	// avoid parallel tests because of hardcoded HTTP ports
	t.Setenv("DoNotRunIn", "parallel")

	t.Run("should successfully call target first with target-internal user, but never casHandler", func(t *testing.T) {
		nexusMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "", r.Header.Get("asdf"))
			defer r.Body.Close()
		}))
		defer nexusMock.Close()
		nexusPortSplit := strings.Split(nexusMock.URL, ":")
		nexusPort, _ := strconv.Atoi(nexusPortSplit[2])
		println("Nexus identifies as", nexusMock.URL)

		casMock := httptest.NewServer(http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect request against CAS: %#v", req)
		}))
		defer casMock.Close()
		println("CAS identifies as", casMock.URL)

		mockCasHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect req against CAS handler: %#v", req)
		})

		conf := Configuration{
			BaseUrl:                            "http://127.0.0.1",
			CasUrl:                             casMock.URL,
			ServiceUrl:                         nexusMock.URL,
			Target:                             "", // will be set below
			ResourcePath:                       "/nexus/repository",
			SkipSSLVerification:                true,
			Port:                               nexusPort,
			PrincipalHeader:                    "X-CARP-Authentication",
			LogoutMethod:                       "DELETE",
			LogoutPath:                         "/rapture/session",
			ForwardUnauthenticatedRESTRequests: true,
			LoggingFormat:                      " %{level:.4s} [%{module}:%{shortfile}] %{message}",
			LogLevel:                           "INFO",
		}

		sut, err := NewDoguRestHandler(conf, mockCasHandler)
		require.NoError(t, err)

		carpServer := httptest.NewUnstartedServer(sut)
		defer carpServer.Close()
		println("carp identifies as", carpServer.Listener.Addr().String())
		//TODO does not work because it is too late: Target was already parsed. FOREVER!
		conf.Target = fmt.Sprintf("http://%s", carpServer.Listener.Addr().String())
		carpServer.Start()

		// when
		requestUrlRaw, err := url.JoinPath(carpServer.URL, "/nexus/repository/supersecret/file")
		require.NoError(t, err)
		requestUrl, err := url.Parse(requestUrlRaw)
		require.NoError(t, err)

		req := &http.Request{
			Method: http.MethodGet,
			URL:    requestUrl,
			Header: map[string][]string{
				"X-Forwarded-For": {"10.20.30.40"},
				"Authorization":   {"Basic " + nexusLocalUserAuthentication},
			},
		}

		resp, err := (&http.Client{}).Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
