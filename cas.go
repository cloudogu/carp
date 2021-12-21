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
	log.Debug("Entering Method 'NewCasClientFactory'")

	log.Debugf("Param '%s'", configuration)

	casUrl, err := url.Parse(configuration.CasUrl)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, errors.Wrapf(err, "failed to parse cas url: %s", configuration.CasUrl)
	}

	serviceUrl, err := url.Parse(configuration.ServiceUrl)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return nil, errors.Wrapf(err, "failed to parse service url: %s", configuration.ServiceUrl)
	}

	urlScheme := cas.NewDefaultURLScheme(casUrl)
	log.Debugf("Variable: %s", urlScheme)
	urlScheme.ServiceValidatePath = path.Join("p3", "serviceValidate")
	log.Debugf("Variable: %s", urlScheme.ServiceValidatePath)

	httpClient := &http.Client{}
	if configuration.SkipSSLVerification {
		log.Debugf("Condition true: 'configuration.SkipSSLVerification'")
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		log.Debugf("Variable: %s", transport)
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
	log.Debug("Entering Method 'CreateClient'")

	log.Debug("Entering Method 'CreateClient'")
	log.Debugf("Variable: %s", factory.urlScheme)
	log.Debugf("Variable: %s", factory.httpClient)
	return cas.NewClient(&cas.Options{
		URLScheme: factory.urlScheme,
		Client:    factory.httpClient,
	})
}

func (factory *CasClientFactory) CreateRestClient() *cas.RestClient {
	log.Debug("Entering Method 'CreateRestClient'")
	log.Debugf("Variable: %s", factory.serviceUrl)
	log.Debugf("Variable: %s", factory.urlScheme)
	log.Debugf("Variable: %s", factory.httpClient)
	log.Debugf("Variable: %s", factory.forwardUnauthenticatedRESTRequests)
	return cas.NewRestClient(&cas.RestOptions{
		ServiceURL:                         factory.serviceUrl,
		URLScheme:                          factory.urlScheme,
		Client:                             factory.httpClient,
		ForwardUnauthenticatedRESTRequests: factory.forwardUnauthenticatedRESTRequests,
	})
}
