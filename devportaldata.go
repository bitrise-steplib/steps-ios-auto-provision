package main

import (
	"fmt"
	"strings"
)

// DeviceData ...
type DeviceData struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	DeviceID   string `json:"device_identifier"`
	Title      string `json:"title"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	DeviceType string `json:"device_type"`
}

// DevPortalData ...
type DevPortalData struct {
	KeyID       string       `json:"key_id"`
	IssuerID    string       `json:"issuer_id"`
	PrivateKey  string       `json:"private_key"`
	TestDevices []DeviceData `json:"test_devices"`
}

// PrivateKeyWithHeader adds header and footer if needed
func (d DevPortalData) PrivateKeyWithHeader() string {
	if strings.HasPrefix(d.PrivateKey, "-----BEGIN PRIVATE KEY----") {
		return d.PrivateKey
	}

	return fmt.Sprint(
		"-----BEGIN PRIVATE KEY-----\n",
		d.PrivateKey,
		"\n-----END PRIVATE KEY-----",
	)
}
