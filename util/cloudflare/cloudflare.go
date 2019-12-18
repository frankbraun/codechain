// Package cloudflare consumes the Cloudflare API (https://api.cloudflare.com/).
package cloudflare

import (
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/frankbraun/codechain/util/log"
)

// Session for the Cloudflare API.
type Session struct {
	api *cloudflare.API
}

// New opens a new Cloudflare API session.
func New(apiKey, email string) (*Session, error) {
	var s Session
	var err error
	log.Printf("Start new session with API Key=%s and Email=%s", apiKey, email)
	s.api, err = cloudflare.New(apiKey, email)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// NewWithConfig opens a new Cloudflare API session with parameters from config.
func NewWithConfig(c *Config) (*Session, error) {
	return New(c.APIKey, c.Email)
}

// TXTCreate creates a TXT record.
func (s *Session) TXTCreate(zone, fqdn, txtdata string, ttl int) (string, error) {
	zoneID, err := s.api.ZoneIDByName(zone)
	if err != nil {
		return "", err
	}
	rr := cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    fqdn,
		Content: txtdata,
		TTL:     ttl,
	}
	res, err := s.api.CreateDNSRecord(zoneID, rr)
	if err != nil {
		return "", err
	}
	jsn, err := json.MarshalIndent(res.Result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsn), nil
}

func (s *Session) getIDs(zone, fqdn string) (string, string, error) {
	zoneID, err := s.api.ZoneIDByName(zone)
	if err != nil {
		return "", "", err
	}
	rr := cloudflare.DNSRecord{
		Type: "TXT",
		Name: fqdn,
	}
	recs, err := s.api.DNSRecords(zoneID, rr)
	if err != nil {
		return "", "", err
	}
	if len(recs) == 0 {
		return "", "", fmt.Errorf("cloudflare: no TXT entry found for %s", fqdn)
	}
	if len(recs) > 1 {
		return "", "", fmt.Errorf("cloudflare: more than one TXT entry found for %s", fqdn)
	}
	recordID := recs[0].ID
	return zoneID, recordID, nil
}

// TXTUpdate updates a TXT record.
func (s *Session) TXTUpdate(zone, fqdn, txtdata string, ttl int) error {
	zoneID, recordID, err := s.getIDs(zone, fqdn)
	if err != nil {
		return err
	}
	rr := cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    fqdn,
		Content: txtdata,
		TTL:     ttl,
	}
	return s.api.UpdateDNSRecord(zoneID, recordID, rr)
}

// TXTDelete deletes all TXT records.
func (s *Session) TXTDelete(zone, fqdn string) error {
	zoneID, recordID, err := s.getIDs(zone, fqdn)
	if err != nil {
		return err
	}
	return s.api.DeleteDNSRecord(zoneID, recordID)
}
