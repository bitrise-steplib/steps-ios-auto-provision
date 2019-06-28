package appstoreconnect

import (
	"net/http"
	"net/url"

	"github.com/bitrise-io/xcode-project/serialized"
)

// ProfilesURL ...
const ProfilesURL = "profiles"

// ListProfilesOptions ...
type ListProfilesOptions struct {
	FilterProfileState ProfileState `url:"filter[profileState],omitempty"`
	FilterProfileType  ProfileType  `url:"filter[profileType],omitempty"`
	FilterName         string       `url:"filter[name],omitempty"`
	Include            string       `url:"include,omitempty"`

	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
	Next   string `url:"-"`
}

// BundleIDPlatform ...
type BundleIDPlatform string

// BundleIDPlatforms ...
const (
	IOS   BundleIDPlatform = "IOS"
	MacOS BundleIDPlatform = "MAC_OS"
)

// ProfileState ...
type ProfileState string

// ProfileStates ...
const (
	Active  ProfileState = "ACTIVE"
	Invalid ProfileState = "INVALID"
)

// ProfileType ...
type ProfileType string

// ProfileTypes ...
const (
	IOSAppDevelopment  ProfileType = "IOS_APP_DEVELOPMENT"
	IOSAppStore        ProfileType = "IOS_APP_STORE"
	IOSAppAdHoc        ProfileType = "IOS_APP_ADHOC"
	IOSAppInHouse      ProfileType = "IOS_APP_INHOUSE"
	MacAppDevelopment  ProfileType = "MAC_APP_DEVELOPMENT"
	MacAppStore        ProfileType = "MAC_APP_STORE"
	MacAppDirect       ProfileType = "MAC_APP_DIRECT"
	TvOSAppDevelopment ProfileType = "TVOS_APP_DEVELOPMENT"
	TvOSAppStore       ProfileType = "TVOS_APP_STORE"
	TvOSAppAdHoc       ProfileType = "TVOS_APP_ADHOC"
	TvOSAppInHouse     ProfileType = "TVOS_APP_INHOUSE"
)

// ProfileAttributes ...
type ProfileAttributes struct {
	Name           string           `json:"name"`
	Platform       BundleIDPlatform `json:"platform"`
	ProfileContent string           `json:"profileContent"`
	UUID           string           `json:"uuid"`
	CreatedDate    string           `json:"createdDate"`
	ProfileState   ProfileState     `json:"profileState"`
	ProfileType    ProfileType      `json:"profileType"`
	ExpirationDate string           `json:"expirationDate"`
}

// Profile ...
type Profile struct {
	Attributes ProfileAttributes `json:"attributes"`

	Relationships struct {
		BundleID struct {
			Links struct {
				Related string `json:"related"`
				Self    string `json:"self"`
			} `json:"links"`
		} `json:"bundleId"`

		Certificates struct {
			Links struct {
				Related string `json:"related"`
				Self    string `json:"self"`
			} `json:"links"`
		} `json:"certificates"`

		Devices struct {
			Links struct {
				Related string `json:"related"`
				Self    string `json:"self"`
			} `json:"links"`
		} `json:"devices"`
	} `json:"relationships"`

	ID string `json:"id"`
}

// ProfilesResponse ...
type ProfilesResponse struct {
	Data     []Profile `json:"data"`
	Included []struct {
		Type       string            `json:"type"`
		ID         string            `json:"id"`
		Attributes serialized.Object `json:"attributes"`
	} `json:"included"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

// ListProfiles ...
func (s ProvisioningService) ListProfiles(opt *ListProfilesOptions) (*ProfilesResponse, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			return nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}

	u, err := addOptions(ProfilesURL, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	r := &ProfilesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

// ProfileCreateRequestDataAttributes ...
type ProfileCreateRequestDataAttributes struct {
	Name        string      `json:"name"`
	ProfileType ProfileType `json:"profileType"`
}

// ProfileCreateRequestDataRelationshipData ...
type ProfileCreateRequestDataRelationshipData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ProfileCreateRequestDataRelationshipsBundleID ...
type ProfileCreateRequestDataRelationshipsBundleID struct {
	Data ProfileCreateRequestDataRelationshipData `json:"data"`
}

// ProfileCreateRequestDataRelationshipsCertificates ...
type ProfileCreateRequestDataRelationshipsCertificates struct {
	Data []ProfileCreateRequestDataRelationshipData `json:"data"`
}

// ProfileCreateRequestDataRelationshipsDevices ...
type ProfileCreateRequestDataRelationshipsDevices struct {
	Data []ProfileCreateRequestDataRelationshipData `json:"data"`
}

// ProfileCreateRequestDataRelationships ...
type ProfileCreateRequestDataRelationships struct {
	BundleID     ProfileCreateRequestDataRelationshipsBundleID     `json:"bundleId"`
	Certificates ProfileCreateRequestDataRelationshipsCertificates `json:"certificates"`
	Devices      ProfileCreateRequestDataRelationshipsDevices      `json:"devices"`
}

// ProfileCreateRequestData ...
type ProfileCreateRequestData struct {
	Attributes    ProfileCreateRequestDataAttributes    `json:"attributes"`
	Relationships ProfileCreateRequestDataRelationships `json:"relationships"`
	Type          string                                `json:"type"`
}

// ProfileCreateRequest ...
type ProfileCreateRequest struct {
	Data ProfileCreateRequestData `json:"data"`
}

// ProfileResponse ...
type ProfileResponse struct {
	Data  Profile            `json:"data"`
	Links PagedDocumentLinks `json:"links,omitempty"`
}

// CreateProfile ...
func (s ProvisioningService) CreateProfile(body ProfileCreateRequest) (*ProfileResponse, error) {
	req, err := s.client.NewRequest(http.MethodPost, ProfilesURL, body)
	if err != nil {
		return nil, err
	}

	r := &ProfileResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}
