package appstoreconnect

import (
	"fmt"
	"net/http"
	"net/url"
)

// CertificatesURL ...
const CertificatesURL = "certificates"

// ListCertificatesOptions ...
type ListCertificatesOptions struct {
	FilterSerialNumber    string          `url:"filter[serialNumber],omitempty"`
	FilterCertificateType CertificateType `url:"filter[certificateType],omitempty"`

	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
	Next   string `url:"-"`
}

// CertificateType ...
type CertificateType string

// CertificateTypes ...
const (
	IOSDevelopment           CertificateType = "IOS_DEVELOPMENT"
	IOSDistribution          CertificateType = "IOS_DISTRIBUTION"
	MacDistribution          CertificateType = "MAC_APP_DISTRIBUTION"
	MacInstallerDistribution CertificateType = "MAC_INSTALLER_DISTRIBUTION"
	MacDevelopment           CertificateType = "MAC_APP_DEVELOPMENT"
	DeveloperIDKext          CertificateType = "DEVELOPER_ID_KEXT"
	DeveloperIDApplication   CertificateType = "DEVELOPER_ID_APPLICATION"
)

// CertificateAttributes ...
type CertificateAttributes struct {
	CertificateContent string           `json:"certificateContent"`
	DisplayName        string           `json:"displayName"`
	ExpirationDate     string           `json:"expirationDate"`
	Name               string           `json:"name"`
	Platform           BundleIDPlatform `json:"platform"`
	SerialNumber       string           `json:"serialNumber"`
	CertificateType    CertificateType  `json:"certificateType"`
}

// Certificate ...
type Certificate struct {
	Attributes CertificateAttributes `json:"attributes"`
	ID         string                `json:"id"`
	Type       string                `json:"type"`
}

// CertificatesResponse ...
type CertificatesResponse struct {
	Data  []Certificate      `json:"data"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

// ListCertificates ...
func (s ProvisioningService) ListCertificates(opt *ListCertificatesOptions) (*CertificatesResponse, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			return nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}

	u, err := addOptions(CertificatesURL, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &CertificatesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

// FetchCertificate fetch the certificate entity from the
func (s ProvisioningService) FetchCertificate(serialNumber string) (Certificate, error) {
	r, err := s.ListCertificates(&ListCertificatesOptions{
		FilterSerialNumber: serialNumber,
	})
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to fetch certificate %s from App Store Connect, error :%s", serialNumber, err)
	}

	if len(r.Data) == 0 {
		return Certificate{}, fmt.Errorf("no certificate entity found for %s", serialNumber)
	} else if len(r.Data) == 0 {
		return Certificate{}, fmt.Errorf("multiple certificate entity found for %s . Entyties: %s", serialNumber, r.Data)
	}
	return r.Data[0], nil
}
