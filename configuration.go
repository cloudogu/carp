package main

type Configuration struct {
	CasUrl              string `yaml:"cas-url"`
	ServiceUrl          string `yaml:"service-url"`
	Target              string `yaml:"target-url"`
	SkipSSLVerification bool   `yaml:"skip-ssl-verification"`
	Port                int    `yaml:"port"`
	PrincipalHeader     string `yaml:"principal-header"`
	UserReplicator      UserReplicator
}
