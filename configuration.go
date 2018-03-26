package main

type Configuration struct {
	CasUrl              string
	ServiceUrl          string
	Target              string
	SkipSSLVerification bool
	Port                int
	PrincipalHeader     string
	UserReplicator      UserReplicator
}
