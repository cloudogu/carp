package main

import (
	"flag"
    "gopkg.in/yaml.v2"
    "os"
    "fmt"
    "io/ioutil"
    "github.com/golang/glog"
    "strings"
)

var Version = "x.y.z-dev"

func main() {
	flag.Parse()

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
        fmt.Println("could not find configuration at", confPath)
        os.Exit(2)
    }

    glog.Infof("start carp %s", Version)

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
