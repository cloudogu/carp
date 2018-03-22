package main

import (
	"net/http"
	"net/url"

	"crypto/tls"

	"strconv"

	"path"

	"fmt"

	"github.com/cloudogu/go-cas"
	"github.com/pkg/errors"
	"github.com/vulcand/oxy/forward"
)

type Configuration struct {
	CasUrl              string
	Target              string
	SkipSSLVerification bool
	Port                int
	PrincipalHeader     string
	UserReplicator      UserReplicator
}

func NewCarpServer(configuration Configuration) (*http.Server, error) {
	casClient, err := createCasClient(configuration)
	if err != nil {
		return nil, err
	}

	handler, err := createRequestHandler(configuration)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    ":" + strconv.Itoa(configuration.Port),
		Handler: casClient.Handle(handler),
	}, nil
}

func createRequestHandler(configuration Configuration) (http.HandlerFunc, error) {
	target, err := url.Parse(configuration.Target)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", configuration.Target)
	}

	fwd, err := forward.New(forward.PassHostHeader(true))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create forward")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		// TODO handle non browser clients

		if !cas.IsAuthenticated(req) {
			cas.RedirectToLogin(w, req)
			return
		}

		username := cas.Username(req)
		if cas.IsFirstAuthenticatedRequest(req) {
			if configuration.UserReplicator != nil {
				attributes := cas.Attributes(req)
				err := configuration.UserReplicator(username, UserAttibutes(attributes))
				if err != nil {
					fmt.Printf("failed to replicate user: %v", err)
				}
			}
		}

		req.Header.Set(configuration.PrincipalHeader, username)

		req.URL = target
		fwd.ServeHTTP(w, req)
	}, nil
}

func createCasClient(configuration Configuration) (*cas.Client, error) {
	casUrl, err := url.Parse(configuration.CasUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", configuration.CasUrl)
	}

	urlScheme := cas.NewDefaultURLScheme(casUrl)
	urlScheme.ServiceValidatePath = path.Join("p3", "serviceValidate")

	options := &cas.Options{
		URLScheme: urlScheme,
	}
	if configuration.SkipSSLVerification {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		options.Client = &http.Client{Transport: tr}
	}

	return cas.NewClient(options), nil
}
