package appstoreconnect

import "net/http"

// BundleIDCapabilitiesURL ...
const BundleIDCapabilitiesURL = "bundleIdCapabilities"

// CapabilityType ...
type CapabilityType string

// CapabilityTypes ...
const (
	Icloud                         CapabilityType = "ICLOUD"
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
	Description      string              `json:"description"`
	Enabled          bool                `json:"enabled"`
	EnabledByDefault bool                `json:"enabledByDefault"`
	Key              CapabilityOptionKey `json:"key"`
	Name             string              `json:"name"`
	SupportsWildcard bool                `json:"supportsWildcard"`
}

// CapabilitySetting ...
type CapabilitySetting struct {
	AllowedInstances CapabilitySettingAllowedInstances `json:"allowedInstances"`
	Description      string                            `json:"description"`
	EnabledByDefault bool                              `json:"enabledByDefault"`
	Key              CapabilitySettingKey              `json:"key"`
	Name             string                            `json:"name"`
	Options          []CapabilityOption                `json:"options"`
	Visible          bool                              `json:"visible"`
	MinInstances     int                               `json:"minInstances"`
}

// BundleIDCapabilityCreateRequestDataAttributes ...
type BundleIDCapabilityCreateRequestDataAttributes struct {
	CapabilityType CapabilityType      `json:"capabilityType"`
	Settings       []CapabilitySetting `json:"settings"`
}

// BundleIDCapabilityCreateRequestDataRelationships ...
type BundleIDCapabilityCreateRequestDataRelationships struct {
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

// BundleIDCapability ...
type BundleIDCapability struct {
	Attributes struct {
		CapabilityType CapabilityType      `json:"capabilityType"`
		Settings       []CapabilitySetting `json:"settings"`
	}
	ID   string `json:"id"`
	Type string `json:"type"`
}

// BundleIDCapabilityResponse ...
type BundleIDCapabilityResponse struct {
	Data BundleIDCapability `json:"data"`
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
