//go:generate go run scripts/generate.go nexus_scripts scripts
package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

func NewNexusUserReplocator(url string, username string, password string) *NexusUserReplicator {
	return &NexusUserReplicator{
		URL:      url,
		Username: username,
		Password: password,
	}
}

type NexusUserReplicator struct {
	URL      string
	Username string
	Password string
}

func (repl *NexusUserReplicator) CreateScript() error {
	payload := make(map[string]string)

	name := "carp-user-replication"

	payload["name"] = "carp-user-replication"
	payload["type"] = "groovy"
	payload["content"] = CARP_USER_REPLICATION

	data, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal payload")
	}

	client := new(http.Client)

	buffer := bytes.NewBuffer(data)

	scriptBaseUrl := repl.URL + "/service/rest/v1/script"

	req, err := http.NewRequest("GET", scriptBaseUrl+"/"+name, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}
	req.SetBasicAuth(repl.Username, repl.Password)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to check if script exists")
	}

	method := "POST"
	url := scriptBaseUrl
	if resp.StatusCode == http.StatusOK {
		method = "PUT"
		url = scriptBaseUrl + "/" + name
	}

	req, err = http.NewRequest(method, url, buffer)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(repl.Username, repl.Password)

	resp, err = client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to create script")
	}

	if resp.StatusCode != 204 {
		return errors.Errorf("creation of user replication script failed, server returned %v", resp.StatusCode)
	}

	return nil
}

func (repl *NexusUserReplicator) Replicate(username string, attributes UserAttibutes) error {
	client := new(http.Client)
	url := repl.URL + "/service/rest/v1/script/carp-user-replication/run"

	nexusUser := createNexusCarpUser(attributes)
	userData, err := json.Marshal(nexusUser)
	if err != nil {
		return errors.Wrap(err, "failed to marshal user attributes")
	}

	buffer := bytes.NewBuffer(userData)
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return errors.Wrapf(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "text/plain")
	req.SetBasicAuth(repl.Username, repl.Password)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to create script")
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("creation of user replication script failed, server returned %v", resp.StatusCode)
	}

	return nil
}

func createNexusCarpUser(attributes UserAttibutes) *NexusCarpUser {
	return &NexusCarpUser{
		Username:  firstOrEmpty(attributes["username"]),
		FirstName: firstOrEmpty(attributes["givenName"]),
		LastName:  firstOrEmpty(attributes["surname"]),
		Email:     firstOrEmpty(attributes["mail"]),
	}
}

func firstOrEmpty(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

type NexusCarpUser struct {
	Username  string
	FirstName string
	LastName  string
	Email     string
}
