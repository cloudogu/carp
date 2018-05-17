package carp

import (
    "os"
    "strings"
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "github.com/pkg/errors"
)

type Configuration struct {
	CasUrl              string `yaml:"cas-url"`
	ServiceUrl          string `yaml:"service-url"`
	Target              string `yaml:"target-url"`
	SkipSSLVerification bool   `yaml:"skip-ssl-verification"`
	Port                int    `yaml:"port"`
	PrincipalHeader     string `yaml:"principal-header"`
	LogoutMethod        string `yaml:"logout-method"`
	LogoutPath          string `yaml:"logout-path"`
	UserReplicator      UserReplicator
}

func ReadConfiguration() (Configuration, error) {
    configuration := Configuration{}

    confPath := "carp.yml"
    if len(os.Args) > 1 {
        for _, arg := range os.Args[1:] {
            if ! strings.HasPrefix(arg, "-") {
                confPath = arg
                break
            }
        }
    }

    if _, err := os.Stat(confPath); os.IsNotExist(err) {
        return configuration, errors.Errorf("could not find configuration at %s", confPath)
    }

    data, err := ioutil.ReadFile(confPath)
    if err != nil {
        return configuration, errors.Wrapf(err, "failed to read configuration %s", confPath)
    }

    err = yaml.Unmarshal(data, &configuration)
    if err != nil {
        return configuration, errors.Wrapf(err, "failed to unmarshal configuration %s", confPath)
    }

    return configuration, nil
}
