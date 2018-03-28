package main

import (
	"flag"
    "gopkg.in/yaml.v2"
    "os"
    "fmt"
    "io/ioutil"
    "github.com/golang/glog"
)

var Version = "x.y.z-dev"

func main() {
	flag.Parse()
	glog.Infof("start carp %s", Version)

    confPath := "carp.yml"
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
        fmt.Println("could not find carp.yml")
        os.Exit(2)
    }

    data, err := ioutil.ReadFile(confPath)
    if err != nil {
        panic(err)
    }

	configuration := Configuration{}
	err = yaml.Unmarshal(data, &configuration)
	if err != nil {
	    panic(err)
    }

	server, err := NewCarpServer(configuration)
	if err != nil {
		panic(err)
	}

	server.ListenAndServe()
}
