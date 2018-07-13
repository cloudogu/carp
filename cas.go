package carp

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"path"

	"github.com/cloudogu/go-cas"
	"github.com/pkg/errors"
)

func NewCasClientFactory(configuration Configuration) (*CasClientFactory, error) {
	casUrl, err := url.Parse(configuration.CasUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse cas url: %s", configuration.CasUrl)
	}

	serviceUrl, err := url.Parse(configuration.ServiceUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse service url: %s", configuration.ServiceUrl)
	}

	urlScheme := cas.NewDefaultURLScheme(casUrl)
	urlScheme.ServiceValidatePath = path.Join("p3", "serviceValidate")

	httpClient := &http.Client{}
	if configuration.SkipSSLVerification {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.Transport = transport
	}

	return &CasClientFactory{
		serviceUrl:                         serviceUrl,
		urlScheme:                          urlScheme,
		httpClient:                         httpClient,
		forwardUnauthenticatedRESTRequests: configuration.ForwardUnauthenticatedRESTRequests,
	}, nil
}

type CasClientFactory struct {
	urlScheme                          cas.URLScheme
	httpClient                         *http.Client
	serviceUrl                         *url.URL
	forwardUnauthenticatedRESTRequests bool
}

func (factory *CasClientFactory) CreateClient() *cas.Client {
	return cas.NewClient(&cas.Options{
		URLScheme: factory.urlScheme,
		Client:    factory.httpClient,
	})
}

func (factory *CasClientFactory) CreateRestClient() *cas.RestClient {
	return cas.NewRestClient(&cas.RestOptions{
		ServiceURL: factory.serviceUrl,
		URLScheme:  factory.urlScheme,
		Client:     factory.httpClient,
		ForwardUnauthenticatedRESTRequests: factory.forwardUnauthenticatedRESTRequests,
	})
}
