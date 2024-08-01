package carp

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTP header values
const (
	httpValueBasicAuthNexusLocalUser = "bG9jYWxVc2VybmFtZTpwYXNzd29yZA=="
	httpValueBasicAuthBruteForceUser = "YXR0YWNrZXI6aDR4eDByNQ=="
)

const (
	someExternalClientIp = "10.20.30.40"
	someTargetFile       = "/nexus/repository/supersecret/file"
)

func TestNewDoguRestHandler(t *testing.T) {
	const requestCount = 3

	t.Run("should successfully GET 3x a target file first with target-internal user, but never call casHandler", func(t *testing.T) {
		requestCallCount := 0
		nexusCallCount := 0
		nexusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			// then
			expectedURI := fmt.Sprintf("%s/%d", someTargetFile, requestCallCount)
			nexusCallCount++

			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "Basic "+httpValueBasicAuthNexusLocalUser, r.Header.Get(httpHeaderAuthorization))
			assert.Equal(t, someExternalClientIp+", 127.0.0.1", r.Header.Get(httpHeaderXForwardedFor))
			assert.Equal(t, expectedURI, r.RequestURI)
			w.WriteHeader(http.StatusOK)
		})

		casHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect request against CAS: %#v", r)
		})

		carpCasHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect req against CAS handler: %#v", r)
		})

		requestUrl := prepareServers(t, nexusHandler, casHandler, carpCasHandler)

		// when
		for requestCallCount = 0; requestCallCount < requestCount; requestCallCount++ {
			t.Run("req#"+strconv.Itoa(requestCallCount), func(t *testing.T) {
				req := &http.Request{
					Method: http.MethodGet,
					URL:    requestUrl.JoinPath(strconv.Itoa(requestCallCount)),
					Header: map[string][]string{
						httpHeaderXForwardedFor: {someExternalClientIp},
						httpHeaderAuthorization: {"Basic " + httpValueBasicAuthNexusLocalUser},
					},
				}

				resp, err := (&http.Client{}).Do(req)

				// then cont'd
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.InDelta(t, 150.0, getLimiter(someExternalClientIp).Tokens(), 0.5)
				assert.Equal(t, requestCallCount, nexusCallCount-1, "unexpected target request count; did some requests went AWOL?")
			})
		}
	})
	t.Run("attack with unknown user and throttle by 3 tokens, fail target handler, fail casHandler", func(t *testing.T) {
		requestCallCount := 0
		nexusCallCount := 0
		nexusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			// then
			expectedURI := fmt.Sprintf("%s/%d", someTargetFile, requestCallCount)
			nexusCallCount++

			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "Basic "+httpValueBasicAuthBruteForceUser, r.Header.Get(httpHeaderAuthorization))
			assert.Equal(t, someExternalClientIp+", 127.0.0.1", r.Header.Get(httpHeaderXForwardedFor))
			assert.Equal(t, expectedURI, r.RequestURI)
			w.WriteHeader(http.StatusUnauthorized)
		})

		casCallCount := 0
		casHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, r *http.Request) {
			casCallCount++
			defer func() { _ = r.Body.Close() }()
			respWriter.WriteHeader(http.StatusUnauthorized)
		})

		mockCarpCasHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect req against CAS handler: %#v", r)
		})

		requestUrl := prepareServers(t, nexusHandler, casHandler, mockCarpCasHandler)

		// when
		for requestCallCount = 0; requestCallCount < requestCount; requestCallCount++ {
			t.Run("req#"+strconv.Itoa(requestCallCount), func(t *testing.T) {

				reqUrl := requestUrl.JoinPath(strconv.Itoa(requestCallCount)).String()
				req, _ := http.NewRequest(http.MethodGet, reqUrl, nil)
				req.Header = map[string][]string{
					httpHeaderXForwardedFor: {someExternalClientIp},
					httpHeaderAuthorization: {"Basic " + httpValueBasicAuthBruteForceUser},
				}

				httpCli := &http.Client{}
				resp, err := httpCli.Do(req)

				// then cont'd
				require.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
				assert.InDelta(t, 150-requestCallCount-1, getLimiter(someExternalClientIp).Tokens(), 0.9)
				assert.Equal(t, requestCallCount, nexusCallCount-1, "unexpected target request count; did some requests went AWOL?")
				assert.Equal(t, requestCallCount, casCallCount-1, "unexpected cas request count; did some requests went AWOL?")

			})
		}
	})
	t.Run("should successfully GET 3x a target file first with target-internal user, but never call casHandler", func(t *testing.T) {
		nexusCallCount := 0
		nexusHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			// then
			expectedURI := fmt.Sprintf("%s/%d", someTargetFile, nexusCallCount)
			nexusCallCount++

			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "Basic "+httpValueBasicAuthNexusLocalUser, r.Header.Get(httpHeaderAuthorization))
			assert.Equal(t, someExternalClientIp+", 127.0.0.1", r.Header.Get(httpHeaderXForwardedFor))
			assert.Equal(t, expectedURI, r.RequestURI)
			w.WriteHeader(http.StatusOK)

		})

		casHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect request against CAS: %#v", r)
		})

		carpCasHandler := http.HandlerFunc(func(respWriter http.ResponseWriter, r *http.Request) {
			defer func() { _ = r.Body.Close() }()
			respWriter.WriteHeader(http.StatusTeapot)
			t.Errorf("did not expect req against CAS handler: %#v", r)
		})

		requestUrl := prepareServers(t, nexusHandler, casHandler, carpCasHandler)

		// when
		for requestCallCount := 0; requestCallCount < requestCount; requestCallCount++ {
			t.Run("req#"+strconv.Itoa(requestCallCount), func(t *testing.T) {
				req := &http.Request{
					Method: http.MethodGet,
					URL:    requestUrl.JoinPath(strconv.Itoa(requestCallCount)),
					Header: map[string][]string{
						httpHeaderXForwardedFor: {someExternalClientIp},
						httpHeaderAuthorization: {"Basic " + httpValueBasicAuthNexusLocalUser},
					},
				}

				resp, err := (&http.Client{}).Do(req)

				// then cont'd
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.InDelta(t, 150.0, getLimiter(someExternalClientIp).Tokens(), 0.5)
			})
		}
	})
}

func prepareServers(t *testing.T, nexusHandler http.HandlerFunc, casHandler http.HandlerFunc, carpCasHandler http.HandlerFunc) *url.URL {
	t.Helper()

	nexusMock := httptest.NewServer(nexusHandler)
	t.Cleanup(func() { nexusMock.Close() })

	nexusPortSplit := strings.Split(nexusMock.URL, ":")
	nexusPort, _ := strconv.Atoi(nexusPortSplit[2])
	fmt.Println("Nexus identifies as", nexusMock.URL)

	casMock := httptest.NewServer(casHandler)
	t.Cleanup(func() { casMock.Close() })
	fmt.Println("CAS identifies as", casMock.URL)

	// build a listener to avoid the chicken-and-egg problem between server start and parsing conf.Target
	carpListener, err := net.Listen("tcp", "127.0.0.1:0")
	// do not defer listener.Close() here. The closing will be done during server.Close()
	require.NoError(t, err)
	carpServerUrl := "http://" + carpListener.Addr().String()
	fmt.Println("carp identifies as", carpServerUrl)

	conf := Configuration{
		BaseUrl:                            "http://127.0.0.1",
		CasUrl:                             casMock.URL,
		ServiceUrl:                         "https://ces.org/nexus",
		Target:                             nexusMock.URL,
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

	sut, err := NewDoguRestHandler(conf, carpCasHandler)
	require.NoError(t, err)

	carpServer := httptest.NewUnstartedServer(sut)
	t.Cleanup(func() { carpServer.Close() })

	// replace the default listener with our own
	_ = carpServer.Listener.Close()
	carpServer.Listener = carpListener
	carpServer.Start()

	requestUrlRaw, err := url.JoinPath(carpServer.URL, someTargetFile)
	require.NoError(t, err)
	requestUrl, err := url.Parse(requestUrlRaw)
	require.NoError(t, err)

	return requestUrl
}
