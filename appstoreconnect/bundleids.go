package appstoreconnect

import (
	"net/http"
	"strings"
)

// BundleIDsURL ...
const BundleIDsURL = "bundleIds"

// ListBundleIDsOptions ...
type ListBundleIDsOptions struct {
	PagingOptions
	FilterIdentifier string           `url:"filter[identifier],omitempty"`
	FilterName       string           `url:"filter[name],omitempty"`
	FilterPlatform   BundleIDPlatform `url:"filter[platform],omitempty"`
	Include          string           `url:"include,omitempty"`
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
	if err := opt.UpdateCursor(); err != nil {
		return nil, err
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
