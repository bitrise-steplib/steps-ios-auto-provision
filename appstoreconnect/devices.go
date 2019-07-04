package appstoreconnect

import (
	"net/http"
	"net/url"
)

// DevicesURL ...
const DevicesURL = "devices"

// ListDevicesOptions ...
type ListDevicesOptions struct {
	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
	Next   string `url:"-"`
}

// DeviceClass ...
type DeviceClass string

// DeviceClasses ...
const (
	AppleWatch DeviceClass = "APPLE_WATCH"
	Ipad       DeviceClass = "IPAD"
	Iphone     DeviceClass = "IPHONE"
	Ipod       DeviceClass = "IPOD"
	AppleTV    DeviceClass = "APPLE_TV"
	Mac        DeviceClass = "MAC"
)

// Status ...
type Status string

// Statuses ...
const (
	Enabled  Status = "ENABLED"
	Disabled Status = "DISABLED"
)

// DeviceAttributes ...
type DeviceAttributes struct {
	DeviceClass DeviceClass      `json:"deviceClass"`
	Model       string           `json:"model"`
	Name        string           `json:"name"`
	Platform    BundleIDPlatform `json:"platform"`
	Status      Status           `json:"status"`
	UDID        string           `json:"udid"`
	AddedDate   string           `json:"addedDate"`
}

// Device ...
type Device struct {
	Type       string           `json:"type"`
	ID         string           `json:"id"`
	Attributes DeviceAttributes `json:"attributes"`
}

// DevicesResponse ...
type DevicesResponse struct {
	Data  []Device           `json:"data"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

// ListDevices ...
func (s ProvisioningService) ListDevices(opt *ListDevicesOptions) (*DevicesResponse, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			return nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}

	u, err := addOptions(DevicesURL, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &DevicesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

// DeviceCreateRequestDataAttributes ...
type DeviceCreateRequestDataAttributes struct {
	Name     string           `json:"name"`
	Platform BundleIDPlatform `json:"platform"`
	UDID     string           `json:"udid"`
}

// DeviceCreateRequestData ...
type DeviceCreateRequestData struct {
	Attributes DeviceCreateRequestDataAttributes `json:"attributes"`
	Type       string                            `json:"type"`
}

// DeviceCreateRequest ...
type DeviceCreateRequest struct {
	Data DeviceCreateRequestData `json:"data"`
}

// RegisterNewDevice ...
func (s ProvisioningService) RegisterNewDevice(body DeviceCreateRequest) (*DevicesResponse, error) {
	req, err := s.client.NewRequest(http.MethodPost, DevicesURL, body)
	if err != nil {
		return nil, err
	}

	r := &DevicesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (s ProvisioningService) DevicesOf(selfLink string) (*DevicesResponse, error) {
	req, err := s.client.NewRequest(http.MethodGet, selfLink, nil)
	if err != nil {
		return nil, err
	}

	r := &DevicesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}
