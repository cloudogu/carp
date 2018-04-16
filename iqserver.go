package main

import "net/url"

func NewIQServerUserReplocator(url string, username string, password string) *NexusUserReplicator {
    return &NexusUserReplicator{
        URL:      url,
        Username: username,
        Password: password,
    }
}

type IQServerUserReplocator struct {
    URL      url.URL
    Username string
    Password string
}

func (repl *IQServerUserReplocator) Replicate(username string, attributes UserAttibutes) error {

    // CREATE:
    // POST /iqserver/rest/user
    // {"id":null,"username":"ssdorra","password":"hallo123","firstName":"Sebastian","lastName":"Sdorra","email":"sebastian.sdorra@cloudogu.com"}
    // curl --header "X-CSRF-TOKEN: api" --cookie "CLM-CSRF-TOKEN=api" ...
    // https://help.sonatype.com/iqserver/rest-apis/accessing-rest-apis-via-reverse-proxy-authentication

    // UPDATE:

    // {"id":"ad90e39eb98f457fb792b753c59bbec2","username":"ssdorra","usernameLowercase":"ssdorra","password":"#~FAKE~PASSWORD~#","firstName":"Sebastian","lastName":"Sdorra","email":"sebastian.sdorra@cloudogu.com"}


    return nil

}
