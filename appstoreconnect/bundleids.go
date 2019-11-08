package appstoreconnect

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// BundleIDsURL ...
const BundleIDsURL = "bundleIds"

// ListBundleIDsOptions ...
type ListBundleIDsOptions struct {
	FilterIdentifier string           `url:"filter[identifier],omitempty"`
	FilterName       string           `url:"filter[name],omitempty"`
	FilterPlatform   BundleIDPlatform `url:"filter[platform],omitempty"`
	Include          string           `url:"include,omitempty"`

	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
	Next   string `url:"-"`
}

// PagedDocumentLinks ...
type PagedDocumentLinks struct {
	Next string `json:"next,omitempty"`
}

// BundleIDAttributes ...
type BundleIDAttributes struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Platform   string `json:"platform"`
}

// BundleIDRelationships ...
type BundleIDRelationships struct {
	Profiles struct {
		Links struct {
			Related string `json:"related"`
			Self    string `json:"next"`
		} `json:"links"`
	} `json:"profiles"`

	Capabilities struct {
		Links struct {
			Related string `json:"related"`
			Self    string `json:"next"`
		} `json:"links"`
	} `json:"bundleIdCapabilities"`
}

// BundleID ...
type BundleID struct {
	Attributes    BundleIDAttributes    `json:"attributes"`
	Relationships BundleIDRelationships `json:"relationships"`

	ID   string `json:"id"`
	Type string `json:"type"`
}

// BundleIdsResponse ...
type BundleIdsResponse struct {
	Data  []BundleID         `json:"data,omitempty"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

// ListBundleIDs ...
func (s ProvisioningService) ListBundleIDs(opt *ListBundleIDsOptions) (*BundleIdsResponse, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			return nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}

	u, err := addOptions(BundleIDsURL, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &BundleIdsResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, err
}

// FetchBundleID fetch the provided bundle ID from the App Store Connect
func (s ProvisioningService) FetchBundleID(bundleID string) (BundleID, error) {
	r, err := s.ListBundleIDs(&ListBundleIDsOptions{
		FilterIdentifier: bundleID,
	})
	if err != nil {
		return BundleID{}, fmt.Errorf("failed to fetch bundle ID %s from App Store Connect: %s", bundleID, err)
	}

	if len(r.Data) == 0 {
		return BundleID{}, fmt.Errorf("no bundle ID entity found for %s", bundleID)
	} else if len(r.Data) == 0 {
		return BundleID{}, fmt.Errorf("multiple bundle ID entity found for %s . Entyties: %s", bundleID, r.Data)
	}
	return r.Data[0], nil
}

// BundleIDResponse ...
type BundleIDResponse struct {
	Data BundleID `json:"data,omitempty"`
}

// BundleIDCreateRequestDataAttributes ...
type BundleIDCreateRequestDataAttributes struct {
	Identifier string           `json:"identifier"`
	Name       string           `json:"name"`
	Platform   BundleIDPlatform `json:"platform"`
}

// BundleIDCreateRequestData ...
type BundleIDCreateRequestData struct {
	Attributes BundleIDCreateRequestDataAttributes `json:"attributes"`
	Type       string                              `json:"type"`
}

// BundleIDCreateRequest ...
type BundleIDCreateRequest struct {
	Data BundleIDCreateRequestData `json:"data"`
}

// CreateBundleID ...
func (s ProvisioningService) CreateBundleID(body BundleIDCreateRequest) (*BundleIDResponse, error) {
	req, err := s.client.NewRequest(http.MethodPost, BundleIDsURL, body)
	if err != nil {
		return nil, err
	}

	r := &BundleIDResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

// BundleID ...
func (s ProvisioningService) BundleID(relationshipLink string) (*BundleIDResponse, error) {
	url := strings.TrimPrefix(relationshipLink, baseURL+apiVersion)
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	r := &BundleIDResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}
