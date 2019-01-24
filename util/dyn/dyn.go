// Package dyn consumes the Dyn Managed DNS API (https://help.dyn.com/dns-api-knowledge-base/).
package dyn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/frankbraun/codechain/util/log"
)

// URI for the Dyn Managed DNS API.
const URI = "https://api.dynect.net"

const session = "/REST/Session/"

// Session for the Dyn Managed DNS API.
type Session struct {
	authHeader http.Header
}

func parseReturnedData(data []byte) (map[string]interface{}, error) {
	var ret map[string]interface{}
	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	if ret["status"] != "success" {
		return nil, fmt.Errorf("dyn: API call failed: %s", data)
	}
	return ret, nil
}

// New opens a new Dyn Managed DNS session.
func New(customerName, userName, password string) (*Session, error) {
	var s Session
	log.Printf("Start new session with customer_name=%s, user_name=%s, and password=%s",
		customerName, userName, password)

	jsn := map[string]string{
		"customer_name": customerName,
		"user_name":     userName,
		"password":      password,
	}
	enc, err := json.Marshal(jsn)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(URI+session, "application/json", bytes.NewBuffer(enc))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Println(string(data))

	// parse returned data
	ret, err := parseReturnedData(data)
	if err != nil {
		return nil, err
	}

	// save token
	token := ret["data"].(map[string]interface{})["token"].(string)
	log.Printf("Session started (token=%s)", token)
	s.authHeader = make(http.Header)
	s.authHeader.Set("Content-Type", "application/json")
	s.authHeader.Add("Auth-Token", token)

	return &s, nil
}

// Close a new Dyn Managed DNS session.
func (s *Session) Close() {
	log.Println("Closing session")
	var c http.Client
	req, err := http.NewRequest(http.MethodDelete, URI+session, nil)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	req.Header = s.authHeader
	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	log.Println(string(data))

	// parse returned data
	_, err = parseReturnedData(data)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
}
