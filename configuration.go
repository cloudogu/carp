package carp

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	BaseUrl                            string `yaml:"base-url"`
	CasUrl                             string `yaml:"cas-url"`
	ServiceUrl                         string `yaml:"service-url"`
	Target                             string `yaml:"target-url"`
	ResourcePath                       string `yaml:"resource-path"`
	SkipSSLVerification                bool   `yaml:"skip-ssl-verification"`
	Port                               int    `yaml:"port"`
	PrincipalHeader                    string `yaml:"principal-header"`
	LogoutMethod                       string `yaml:"logout-method"`
	LogoutPath                         string `yaml:"logout-path"`
	ForwardUnauthenticatedRESTRequests bool   `yaml:"forward-unauthenticated-rest-requests"`
	LoggingFormat                      string `yaml:"log-format"`
	LogLevel                           string `yaml:"log-level"`
	UserReplicator                     UserReplicator
	ResponseModifier                   func(*http.Response) error
}

func InitializeAndReadConfiguration() (Configuration, error) {
	log.Debug("Entering Method 'InitializeAndReadConfiguration'")
	configuration, err := readConfiguration()
	log.Debugf("Variable: %s", configuration)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return configuration, errors.Wrap(err, "could not initialize")
	}

	err = prepareLogger(configuration)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return configuration, errors.Wrap(err, "could not initialize")
	}

	log.Debugf("Variable: %s", configuration)
	log.Debug("End of Function 'InitializeAndReadConfiguration'")
	return configuration, nil
}

func readConfiguration() (Configuration, error) {
	log.Debug("Entering Method 'readConfiguration'")
	configuration := Configuration{}

	confPath := "carp.yml"
	if len(os.Args) > 1 {
		log.Debugf("Condition true: 'len(os.Args) > 1'")
		for _, arg := range os.Args[1:] {
			if !strings.HasPrefix(arg, "-") {
				confPath = arg
				break
			}
		}
	}

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		log.Debugf("Error: %s", err.Error())
		return configuration, errors.Errorf("could not find configuration at %s", confPath)
	}

	data, err := ioutil.ReadFile(confPath)
	log.Debugf("Variable: %s", data)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return configuration, errors.Wrapf(err, "failed to read configuration %s", confPath)
	}

	err = yaml.Unmarshal(data, &configuration)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
		return configuration, errors.Wrapf(err, "failed to unmarshal configuration %s", confPath)
	}

	log.Debugf("Variable: %s", configuration)
	log.Debug("End of Function 'readConfiguration'")
	return configuration, nil
}

// Deprecated: ReadConfiguration exists for historical compatibility
func ReadConfiguration() (Configuration, error) {
	log.Debug("Entering Method 'ReadConfiguration'")
	configuration, err := readConfiguration()
	log.Debugf("Variable: %s", configuration)
	if err != nil {
		log.Debugf("Error: %s", err.Error())
	}
	log.Debug("End of Function 'ReadConfiguration'")
	return configuration, err
}
