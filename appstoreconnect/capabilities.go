package appstoreconnect

import (
	"net/http"
	"strings"
)

// BundleIDCapabilitiesURL ...
const BundleIDCapabilitiesURL = "bundleIdCapabilities"

// CapabilityType ...
type CapabilityType string

// CapabilityTypes ...
const (
	ICloud                         CapabilityType = "ICLOUD"
	InAppPurchase                  CapabilityType = "IN_APP_PURCHASE"
	GameCenter                     CapabilityType = "GAME_CENTER"
	PushNotifications              CapabilityType = "PUSH_NOTIFICATIONS"
	Wallet                         CapabilityType = "WALLET"
	InterAppAudio                  CapabilityType = "INTER_APP_AUDIO"
	Maps                           CapabilityType = "MAPS"
	AssociatedDomains              CapabilityType = "ASSOCIATED_DOMAINS"
	PersonalVPN                    CapabilityType = "PERSONAL_VPN"
	AppGroups                      CapabilityType = "APP_GROUPS"
	Healthkit                      CapabilityType = "HEALTHKIT"
	Homekit                        CapabilityType = "HOMEKIT"
	WirelessAccessoryConfiguration CapabilityType = "WIRELESS_ACCESSORY_CONFIGURATION"
	ApplePay                       CapabilityType = "APPLE_PAY"
	DataProtection                 CapabilityType = "DATA_PROTECTION"
	Sirikit                        CapabilityType = "SIRIKIT"
	NetworkExtensions              CapabilityType = "NETWORK_EXTENSIONS"
	Multipath                      CapabilityType = "MULTIPATH"
	HotSpot                        CapabilityType = "HOT_SPOT"
	NFCTagReading                  CapabilityType = "NFC_TAG_READING"
	Classkit                       CapabilityType = "CLASSKIT"
	AutofillCredentialProvider     CapabilityType = "AUTOFILL_CREDENTIAL_PROVIDER"
	AccessWIFIInformation          CapabilityType = "ACCESS_WIFI_INFORMATION"
)

// ServiceNameByKey ...
var ServiceNameByKey = map[string]string{
	"com.apple.security.application-groups":               "App Groups",
	"com.apple.developer.in-app-payments":                 "Apple Pay",
	"com.apple.developer.associated-domains":              "Associated Domains",
	"com.apple.developer.healthkit":                       "HealthKit",
	"com.apple.developer.homekit":                         "HomeKit",
	"com.apple.developer.networking.HotspotConfiguration": "Hotspot",
	"com.apple.InAppPurchase":                             "In-App Purchase",
	"inter-app-audio":                                     "Inter-App Audio",
	"com.apple.developer.networking.multipath":            "Multipath",
	"com.apple.developer.networking.networkextension":     "Network Extensions",
	"com.apple.developer.nfc.readersession.formats":       "NFC Tag Reading",
	"com.apple.developer.networking.vpn.api":              "Personal VPN",
	"aps-environment":                                     "Push Notifications",
	"com.apple.developer.siri":                            "SiriKit",
	"com.apple.developer.pass-type-identifiers":           "Wallet",
	"com.apple.external-accessory.wireless-configuration": "Wireless Accessory Configuration",
}

// ServiceTypeByKey ...
var ServiceTypeByKey = map[string]CapabilityType{
	"com.apple.security.application-groups":               AppGroups,
	"com.apple.developer.in-app-payments":                 ApplePay,
	"com.apple.developer.associated-domains":              AssociatedDomains,
	"com.apple.developer.healthkit":                       Healthkit,
	"com.apple.developer.homekit":                         Homekit,
	"com.apple.developer.networking.HotspotConfiguration": HotSpot,
	"com.apple.InAppPurchase":                             InAppPurchase,
	"inter-app-audio":                                     InterAppAudio,
	"com.apple.developer.networking.multipath":            Multipath,
	"com.apple.developer.networking.networkextension":     NetworkExtensions,
	"com.apple.developer.nfc.readersession.formats":       NFCTagReading,
	"com.apple.developer.networking.vpn.api":              PersonalVPN,
	"aps-environment":                                     PushNotifications,
	"com.apple.developer.siri":                            Sirikit,
	"com.apple.developer.pass-type-identifiers":           Wallet,
	"com.apple.external-accessory.wireless-configuration": WirelessAccessoryConfiguration,
}

// CapabilitySettingAllowedInstances ...
type CapabilitySettingAllowedInstances string

// AllowedInstances ...
const (
	Entry    CapabilitySettingAllowedInstances = "ENTRY"
	Single   CapabilitySettingAllowedInstances = "SINGLE"
	Multiple CapabilitySettingAllowedInstances = "MULTIPLE"
)

// CapabilitySettingKey ...
type CapabilitySettingKey string

// CapabilitySettingKeys
const (
	IcloudVersion                 CapabilitySettingKey = "ICLOUD_VERSION"
	DataProtectionPermissionLevel CapabilitySettingKey = "DATA_PROTECTION_PERMISSION_LEVEL"
)

// CapabilityOptionKey ...
type CapabilityOptionKey string

// CapabilityOptionKeys ...
const (
	Xcode5                      CapabilityOptionKey = "XCODE_5"
	Xcode6                      CapabilityOptionKey = "XCODE_6"
	CompleteProtection          CapabilityOptionKey = "COMPLETE_PROTECTION"
	ProtectedUnlessOpen         CapabilityOptionKey = "PROTECTED_UNLESS_OPEN"
	ProtectedUntilFirstUserAuth CapabilityOptionKey = "PROTECTED_UNTIL_FIRST_USER_AUTH"
)

// CapabilityOption ...
type CapabilityOption struct {
	Description      string              `json:"description,omitempty"`
	Enabled          bool                `json:"enabled,omitempty"`
	EnabledByDefault bool                `json:"enabledByDefault,omitempty"`
	Key              CapabilityOptionKey `json:"key,omitempty"`
	Name             string              `json:"name,omitempty"`
	SupportsWildcard bool                `json:"supportsWildcard,omitempty"`
}

// CapabilitySetting ...
type CapabilitySetting struct {
	AllowedInstances CapabilitySettingAllowedInstances `json:"allowedInstances,omitempty"`
	Description      string                            `json:"description,omitempty"`
	EnabledByDefault bool                              `json:"enabledByDefault,omitempty"`
	Key              CapabilitySettingKey              `json:"key,omitempty"`
	Name             string                            `json:"name,omitempty"`
	Options          []CapabilityOption                `json:"options,omitempty"`
	Visible          bool                              `json:"visible,omitempty"`
	MinInstances     int                               `json:"minInstances,omitempty"`
}

//
// BundleIDCapabilityCreateRequest

// BundleIDCapabilityCreateRequestDataAttributes ...
type BundleIDCapabilityCreateRequestDataAttributes struct {
	CapabilityType CapabilityType      `json:"capabilityType"`
	Settings       []CapabilitySetting `json:"settings"`
}

// BundleIDCapabilityCreateRequestDataRelationships ...
type BundleIDCapabilityCreateRequestDataRelationships struct {
	BundleID BundleIDCapabilityCreateRequestDataRelationshipsBundleID `json:"bundleId"`
}

// BundleIDCapabilityCreateRequestDataRelationshipsBundleID ...
type BundleIDCapabilityCreateRequestDataRelationshipsBundleID struct {
	Data BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData `json:"data"`
}

// BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData ...
type BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// BundleIDCapabilityCreateRequestData ...
type BundleIDCapabilityCreateRequestData struct {
	Attributes    BundleIDCapabilityCreateRequestDataAttributes    `json:"attributes"`
	Relationships BundleIDCapabilityCreateRequestDataRelationships `json:"relationships"`
	Type          string                                           `json:"type"`
}

// BundleIDCapabilityCreateRequest ...
type BundleIDCapabilityCreateRequest struct {
	Data BundleIDCapabilityCreateRequestData `json:"data"`
}

//
// BundleIDCapabilityUpdateRequest

// BundleIDCapabilityUpdateRequestDataAttributes ...
type BundleIDCapabilityUpdateRequestDataAttributes struct {
	CapabilityType CapabilityType      `json:"capabilityType"`
	Settings       []CapabilitySetting `json:"settings"`
}

// BundleIDCapabilityUpdateRequestData ...
type BundleIDCapabilityUpdateRequestData struct {
	Attributes BundleIDCapabilityUpdateRequestDataAttributes `json:"attributes"`
	ID         string                                        `json:"id"`
	Type       string                                        `json:"type"`
}

// BundleIDCapabilityUpdateRequest ...
type BundleIDCapabilityUpdateRequest struct {
	Data BundleIDCapabilityUpdateRequestData `json:"data"`
}

// BundleIDCapabilityAttributes ...
type BundleIDCapabilityAttributes struct {
	CapabilityType CapabilityType      `json:"capabilityType"`
	Settings       []CapabilitySetting `json:"settings"`
}

// BundleIDCapability ...
type BundleIDCapability struct {
	Attributes BundleIDCapabilityAttributes
	ID         string         `json:"id"`
	Type       CapabilityType `json:"type"`
}

// BundleIDCapabilityResponse ...
type BundleIDCapabilityResponse struct {
	Data BundleIDCapability `json:"data"`
}

// BundleIDCapabilitesResponse ...
type BundleIDCapabilitesResponse struct {
	Data []BundleIDCapability `json:"data"`
}

// EnableCapability ...
func (s ProvisioningService) EnableCapability(body BundleIDCapabilityCreateRequest) (*BundleIDCapabilityResponse, error) {
	req, err := s.client.NewRequest(http.MethodPost, BundleIDCapabilitiesURL, body)
	if err != nil {
		return nil, err
	}

	r := &BundleIDCapabilityResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}

// UpdateCapability ...
func (s ProvisioningService) UpdateCapability(id string, body BundleIDCapabilityUpdateRequest) (*BundleIDCapabilityResponse, error) {
	req, err := s.client.NewRequest(http.MethodPatch, BundleIDCapabilitiesURL+"/"+id, body)
	if err != nil {
		return nil, err
	}

	r := &BundleIDCapabilityResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Capabilities ...
func (s ProvisioningService) Capabilities(relationshipLink string) (*BundleIDCapabilitesResponse, error) {
	url := strings.TrimPrefix(relationshipLink, baseURL+"v1")
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	r := &BundleIDCapabilitesResponse{}
	if _, err := s.client.Do(req, r); err != nil {
		return nil, err
	}

	return r, nil
}
