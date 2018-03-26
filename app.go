package main

import (
	"flag"
)

func main() {
	flag.Parse()

	nur := NewNexusUserReplocator("http://localhost:8081", "admin", "admin123")
	err := nur.CreateScript()
	if err != nil {
		panic(err)
	}

	configuration := Configuration{
		CasUrl:              "https://192.168.56.2/cas",
		ServiceUrl:          "http://192.168.56.1:8080",
		Target:              "http://localhost:8081",
		SkipSSLVerification: true,
		Port:                9090,
		PrincipalHeader:     "X-CARP-Authentication",
		UserReplicator:      nur.Replicate,
	}

	server, err := NewCarpServer(configuration)
	if err != nil {
		panic(err)
	}

	server.ListenAndServe()
}
