// Package dyn consumes the Dyn Managed DNS API (https://help.dyn.com/dns-api-knowledge-base/).
package dyn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

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

	jsn := map[string]interface{}{
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

// NewWithConfig opens a new Dyn Managed DNS session with parameters from config.
func NewWithConfig(c *Config) (*Session, error) {
	return New(c.CustomerName, c.UserName, c.Password)
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

// TXTCreate creates a TXT record.
func (s *Session) TXTCreate(zone, fqdn, txtdata string, ttl int) error {
	log.Printf("Create new TXT record with zone=%s, fqdn=%s, txtdata=%s, and ttl=%d",
		zone, fqdn, txtdata, ttl)

	jsn := map[string]interface{}{
		"rdata": map[string]string{
			"txtdata": txtdata,
		},
		"ttl": strconv.Itoa(ttl),
	}
	enc, err := json.Marshal(jsn)
	if err != nil {
		return err
	}

	var c http.Client
	req, err := http.NewRequest(http.MethodPost,
		URI+"/REST/TXTRecord/"+zone+"/"+fqdn+"/", bytes.NewBuffer(enc))
	if err != nil {
		return err
	}
	req.Header = s.authHeader
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(data))

	// parse returned data
	_, err = parseReturnedData(data)
	if err != nil {
		return err
	}
	return nil
}

// TXTUpdate updates a TXT record (replaces all existing TXT records).
func (s *Session) TXTUpdate(zone, fqdn, txtdata string, ttl int) error {
	log.Printf("Update TXT record with zone=%s, fqdn=%s, txtdata=%s, and ttl=%d",
		zone, fqdn, txtdata, ttl)

	jsn := map[string]interface{}{
		"rdata": map[string]string{
			"txtdata": txtdata,
		},
		"ttl": strconv.Itoa(ttl),
	}
	enc, err := json.Marshal(jsn)
	if err != nil {
		return err
	}

	var c http.Client
	req, err := http.NewRequest(http.MethodPut,
		URI+"/REST/TXTRecord/"+zone+"/"+fqdn+"/", bytes.NewBuffer(enc))
	if err != nil {
		return err
	}
	req.Header = s.authHeader
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(data))

	// parse returned data
	_, err = parseReturnedData(data)
	if err != nil {
		return err
	}
	return nil
}

// TXTDelete deletes all TXT records.
func (s *Session) TXTDelete(zone, fqdn string) error {
	log.Printf("Delete TXT records with zone=%s and fqdn=%s", zone, fqdn)

	var c http.Client
	req, err := http.NewRequest(http.MethodDelete,
		URI+"/REST/TXTRecord/"+zone+"/"+fqdn+"/", nil)
	if err != nil {
		return err
	}
	req.Header = s.authHeader
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(data))

	// parse returned data
	_, err = parseReturnedData(data)
	if err != nil {
		return err
	}
	return nil
}

// ZoneChangeset returns the pending changset for zone.
func (s *Session) ZoneChangeset(zone string) (map[string]interface{}, error) {
	log.Printf("Get Zone Changeset for zone=%s", zone)

	var c http.Client
	req, err := http.NewRequest(http.MethodGet, URI+"/REST/ZoneChanges/"+zone, nil)
	if err != nil {
		return nil, err
	}
	req.Header = s.authHeader
	resp, err := c.Do(req)
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
	return ret, nil
}

// ZoneUpdate publishes the pending update for zone.
func (s *Session) ZoneUpdate(zone string) error {
	log.Printf("Publish pending changes for zone=%s", zone)

	jsn := map[string]bool{
		"publish": true,
	}
	enc, err := json.Marshal(jsn)
	if err != nil {
		return err
	}

	var c http.Client
	req, err := http.NewRequest(http.MethodPut, URI+"/REST/Zone/"+zone+"/",
		bytes.NewBuffer(enc))
	if err != nil {
		return err
	}
	req.Header = s.authHeader
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(data))

	// parse returned data
	_, err = parseReturnedData(data)
	if err != nil {
		return err
	}
	return nil
}
