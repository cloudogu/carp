package main

import (
	"flag"
)

func main() {
	flag.Parse()

	configuration := Configuration{
		CasUrl:              "https://192.168.56.2/cas",
		Target:              "http://localhost:8081",
		SkipSSLVerification: true,
		Port:                9090,
		PrincipalHeader:     "X-CARP-Authentication",
	}

	server, err := NewCarpServer(configuration)
	if err != nil {
		panic(err)
	}

	server.ListenAndServe()
}
