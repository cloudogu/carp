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

const SCRIPT = `import groovy.json.JsonSlurper
import org.sonatype.nexus.security.user.UserNotFoundException

// parse json formatted carp user, which is send as argument for the script
def carpUser = new JsonSlurper().parseText(args)

// use undocumented getSecuritySystem to check and update existing users
def securitySystem = security.getSecuritySystem()

// every one should be an admin ;)
// TODO map cas groups to nexus roles?
def adminRole = ['nx-admin']

try {
    log.info('update user ' + carpUser.Username)

    def user = securitySystem.getUser(carpUser.Username)
    user.setFirstName(carpUser.FirstName)
    user.setLastName(carpUser.LastName)
    user.setEmailAddress(carpUser.Email)
    // set active? password?
    securitySystem.updateUser(user)
} catch (UserNotFoundException ex) {
    log.info('create user ' + carpUser.username)

    // user not found, create a new one
    // id, firstName, lastName, Email, active, password, arrayOfRoles
    // what about the password, null is not accepted ?
    security.addUser(carpUser.Username, carpUser.FirstName, carpUser.LastName, carpUser.Email, true, "secretPwd", adminRole)
}
`

func (repl *NexusUserReplicator) CreateScript() error {
	payload := make(map[string]string)

	name := "carp-user-replication"

	payload["name"] = "carp-user-replication"
	payload["type"] = "groovy"
	payload["content"] = SCRIPT

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
