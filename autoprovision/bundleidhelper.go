package autoprovision

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// BundleID ...
type BundleID struct {
	Attributes   appstoreconnect.BundleIDAttributes
	Capabilities []appstoreconnect.BundleIDCapability
	ID           string
}

func appIDName(bundleID string) string {
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return "Bitrise " + r.Replace(bundleID)
}

// EnsureApp search for AppID on the developer portal for the provided bundleID.
// If the AppID is available on the developer portal, it will ne returned.
// If it's not, it will be generated.
func EnsureApp(client *appstoreconnect.Client, platform Platform, bundleID string) (*BundleID, error) {
	var bundleIDPlatform appstoreconnect.BundleIDPlatform
	switch platform {
	case IOS, TVOS:
		bundleIDPlatform = appstoreconnect.IOS
	case MacOS:
		bundleIDPlatform = appstoreconnect.MacOS
	default:
		return nil, fmt.Errorf("unkown platform: %s", platform)
	}

	log.Printf("Search for AppID for the %s bundleID", bundleID)

	b, err := fetchBundleID(client, bundleID)
	if err != nil {
		return nil, err
	}
	if b != nil {
		return b, nil
	}
	log.Warnf("No AppID was found with bundleID: %s", bundleID)

	appIDName := appIDName(bundleID)

	log.Printf("Registering AppID: %s with bundle id: %s", appIDName, bundleID)

	r, err := client.Provisioning.CreateBundleID(
		appstoreconnect.BundleIDCreateRequest{
			Data: appstoreconnect.BundleIDCreateRequestData{
				Attributes: appstoreconnect.BundleIDCreateRequestDataAttributes{
					Identifier: bundleID,
					Name:       appIDName,
					Platform:   bundleIDPlatform,
				},
				Type: "bundleIds",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register AppID for bundleID %s, error: %s", bundleID, err)
	}

	capabilities, err := fetchBundleIDCapabilities(client, r.Data)
	if err != nil {
		return nil, err
	}

	return &BundleID{
		Attributes: appstoreconnect.BundleIDAttributes{
			Identifier: r.Data.Attributes.Identifier,
			Name:       r.Data.Attributes.Name,
			Platform:   r.Data.Attributes.Platform,
		},
		Capabilities: capabilities,
		ID:           r.Data.ID,
	}, nil
}

func fetchBundleID(client *appstoreconnect.Client, bundleID string) (*BundleID, error) {
	r, err := client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{
		FilterIdentifier: bundleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bundleID: %s, error: %s", bundleID, err)
	}
	if len(r.Data) == 0 {
		return nil, nil
	}

	// The FilterIdentifier works as a Like command. It will not search for the exact match,
	// this is why we need to find the exact match in the list.
	var b *BundleID
	for i, d := range r.Data {
		if d.Attributes.Identifier == bundleID {
			capabilities, err := fetchBundleIDCapabilities(client, d)
			if err != nil {
				return nil, err
			}

			b = &BundleID{
				Attributes: appstoreconnect.BundleIDAttributes{
					Identifier: r.Data[i].Attributes.Identifier,
					Name:       r.Data[i].Attributes.Name,
					Platform:   r.Data[i].Attributes.Platform,
				},
				Capabilities: capabilities,
				ID:           r.Data[i].ID,
			}
			break
		}
	}
	return b, nil

}

func fetchBundleIDCapabilities(client *appstoreconnect.Client, bundleID appstoreconnect.BundleID) ([]appstoreconnect.BundleIDCapability, error) {
	c, err := client.Provisioning.CapabilitiesOf(bundleID)
	if err != nil {
		return nil, err
	}

	var bundleIDCapabilities []appstoreconnect.BundleIDCapability
	for _, cap := range c.Data {
		bundleIDCapabilities = append(bundleIDCapabilities, cap)
	}
	return bundleIDCapabilities, nil
}

// syncAppServices compares the target's capabilities one-by-one with the AppID's capability list on the developer portal,
// If the capability is not enabled, enables it.
func syncAppServices(client *appstoreconnect.Client, entitlements serialized.Object, bundleID string, capabilities []appstoreconnect.BundleIDCapability) error {
	for targetEntKey := range entitlements {
		var capabilityEnabled bool
		for _, bundleIDCap := range capabilities {
			if string(bundleIDCap.Attributes.CapabilityType) == string(appstoreconnect.ServiceTypeByKey[targetEntKey]) {
				log.Printf("Capability %s, is already enabled", string(appstoreconnect.ServiceTypeByKey[targetEntKey]))
				capabilityEnabled = true
			}
		}

		if !capabilityEnabled && string(appstoreconnect.ServiceTypeByKey[targetEntKey]) != "" {
			log.Warnf("Capability %s is not enabled", string(appstoreconnect.ServiceTypeByKey[targetEntKey]))
			log.Printf("Enabling capability")

			if err := enableAppService(client, appstoreconnect.ServiceTypeByKey[targetEntKey], bundleID, nil); err != nil {
				return fmt.Errorf("failed to enable capability %v for target: %s, error: %s", appstoreconnect.ServiceTypeByKey[targetEntKey], bundleID, err)
			}
			log.Donef("Capability enabled")
		}
	}

	// Data Protection
	if targetDataProtectionValue, err := entitlements.String("com.apple.developer.default-data-protection"); err != nil && !serialized.IsKeyNotFoundError(err) {
		return fmt.Errorf("failed to get target's data procetion entitlement, error: %s", err)
	} else if targetDataProtectionValue != "" {
		if err := updateDataProtection(client, bundleID, capabilities, targetDataProtectionValue); err != nil {
			return err
		}
	}

	// iCloud
	usesICloudDocuments, usesICloudKit, usesICloudKeyValueStorage, err := usesICloudServices(entitlements)
	if err != nil {
		return fmt.Errorf("failed to check if iCloud capability is enabled for bundleID: %s, error: %s", bundleID, err)
	}

	if usesICloudKeyValueStorage || usesICloudDocuments || usesICloudKit {
		if err := updateICloud(client, bundleID, capabilities, string(appstoreconnect.Xcode6)); err != nil {
			return err
		}
	}
	return nil
}

func usesICloudServices(targetEntitlements serialized.Object) (usesICloudDocuments, usesICloudKit, usesICloudKeyValueStorage bool, ferr error) {
	var err error
	usesICloudKeyValueStorage, err = func() (bool, error) {
		iCloudKeyValueStorage, err := targetEntitlements.String("com.apple.developer.ubiquity-kvstore-identifier")
		if err != nil {
			return false, err
		}
		return iCloudKeyValueStorage != "", nil
	}()
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		ferr = fmt.Errorf("failed to get target's iCLoud key value storage entitlement, error: %s", err)
		return
	}

	iCloudServices, err := targetEntitlements.StringSlice("com.apple.developer.icloud-services")
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		ferr = fmt.Errorf("failed to get target's iCLoud services entitlement, error: %s", err)
		return
	}

	if len(iCloudServices) > 0 {
		usesICloudDocuments = sliceutil.IsStringInSlice("CloudDocuments", iCloudServices)
		usesICloudKit = sliceutil.IsStringInSlice("CloudKit", iCloudServices)
	}
	return
}

func enableAppService(client *appstoreconnect.Client, capabilityType appstoreconnect.CapabilityType, bundleID string, settings []appstoreconnect.CapabilitySetting) error {
	_, err := client.Provisioning.EnableCapability(appstoreconnect.BundleIDCapabilityCreateRequest{
		Data: appstoreconnect.BundleIDCapabilityCreateRequestData{
			Attributes: appstoreconnect.BundleIDCapabilityCreateRequestDataAttributes{
				CapabilityType: capabilityType,
				Settings:       settings,
			},
			Relationships: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationships{
				BundleID: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleID{
					Data: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData{
						ID:   bundleID,
						Type: "bundleIds",
					},
				},
			},
			Type: "bundleIdCapabilities",
		},
	})
	return err
}

func updateAppService(client *appstoreconnect.Client, capabilityID string, capabilityType appstoreconnect.CapabilityType, capabilitySettings []appstoreconnect.CapabilitySetting) error {
	_, err := client.Provisioning.UpdateCapability(capabilityID, appstoreconnect.BundleIDCapabilityUpdateRequest{
		Data: appstoreconnect.BundleIDCapabilityUpdateRequestData{
			Attributes: appstoreconnect.BundleIDCapabilityUpdateRequestDataAttributes{
				CapabilityType: capabilityType,
				Settings:       capabilitySettings,
			},
			Type: "bundleIdCapabilities",
			ID:   capabilityID,
		},
	})
	return err
}

func updateICloud(client *appstoreconnect.Client, bundleID string, capabilities []appstoreconnect.BundleIDCapability, targetICloudVersion string) error {
	var iCloudCapabilityID string
	var iCLoudVersion appstoreconnect.CapabilityOptionKey
	for _, bundleIDCap := range capabilities {
		if bundleIDCap.Attributes.CapabilityType == appstoreconnect.ICloud {
			for _, settings := range bundleIDCap.Attributes.Settings {
				if settings.Key == appstoreconnect.IcloudVersion {
					iCLoudVersion, iCloudCapabilityID = settings.Options[0].Key, bundleIDCap.ID
				}
			}
		}
	}

	if iCloudCapabilityID == "" {
		log.Successf("Set iCloud: on")

		capabilitySettingOption := appstoreconnect.CapabilityOption{Key: appstoreconnect.Xcode6}
		capabilitySetting := appstoreconnect.CapabilitySetting{
			Options: []appstoreconnect.CapabilityOption{capabilitySettingOption},
			Key:     appstoreconnect.IcloudVersion,
		}
		return enableAppService(client, appstoreconnect.ICloud, bundleID, []appstoreconnect.CapabilitySetting{capabilitySetting})
	}

	log.Printf("iCloud: already set")
	if iCLoudVersion == appstoreconnect.Xcode6 {
		log.Printf("CloudKit: already set")
	} else {
		log.Successf("Set CloudKit: on")
		if err := updateICloudVersion(client, iCloudCapabilityID, appstoreconnect.Xcode6); err != nil {
			return fmt.Errorf("failed to update iCloud version, error: %s", err)
		}
	}

	return nil
}

func updateICloudVersion(client *appstoreconnect.Client, capabilityID string, xcodeVersion appstoreconnect.CapabilityOptionKey) error {
	capabilitySettingOption := appstoreconnect.CapabilityOption{Key: xcodeVersion}
	capabilitySetting := appstoreconnect.CapabilitySetting{
		Options: []appstoreconnect.CapabilityOption{capabilitySettingOption},
		Key:     appstoreconnect.IcloudVersion,
	}
	return updateAppService(client, capabilityID, appstoreconnect.ICloud, []appstoreconnect.CapabilitySetting{capabilitySetting})
}

// DataProtections ...
var DataProtections = map[string]appstoreconnect.CapabilityOptionKey{
	"NSFileProtectionComplete":                             appstoreconnect.CompleteProtection,
	"NSFileProtectionCompleteUnlessOpen":                   appstoreconnect.ProtectedUnlessOpen,
	"NSFileProtectionCompleteUntilFirstUserAuthentication": appstoreconnect.ProtectedUntilFirstUserAuth,
}

func updateDataProtection(client *appstoreconnect.Client, bundleID string, capabilities []appstoreconnect.BundleIDCapability, targetProtectionValue string) error {
	var protectionCapabilityID string
	var protectionCapabilityValue appstoreconnect.CapabilityOptionKey

	for _, bundleIDCap := range capabilities {
		if bundleIDCap.Attributes.CapabilityType == appstoreconnect.DataProtection {
			for _, settings := range bundleIDCap.Attributes.Settings {
				if settings.Key == appstoreconnect.DataProtectionPermissionLevel {
					protectionCapabilityValue, protectionCapabilityID = settings.Options[0].Key, bundleIDCap.ID
				}
			}
		}
	}

	value, ok := DataProtections[targetProtectionValue]
	if !ok {
		return errors.New("unknown data protection value: " + targetProtectionValue)
	}

	if protectionCapabilityValue == value {
		log.Printf("Data Protection: until_first_auth already set")
	} else {
		log.Successf("Set Data Protection: until_first_auth")
		if err := updateDataProtectionLVL(client, protectionCapabilityID, appstoreconnect.ProtectedUntilFirstUserAuth); err != nil {
			return fmt.Errorf("failed to update Data Protection Cabability, error: %s", err)
		}
	}

	return nil
}

func updateDataProtectionLVL(client *appstoreconnect.Client, capabilityID string, protectionLVL appstoreconnect.CapabilityOptionKey) error {
	if protectionLVL != appstoreconnect.CompleteProtection && protectionLVL != appstoreconnect.ProtectedUnlessOpen && protectionLVL != appstoreconnect.ProtectedUntilFirstUserAuth {
		return fmt.Errorf("the provided app protection level is invalid: %s. Valid app protection levels: %s", protectionLVL, strings.Join([]string{string(appstoreconnect.CompleteProtection), string(appstoreconnect.ProtectedUnlessOpen), string(appstoreconnect.ProtectedUntilFirstUserAuth)}, ", "))
	}

	capabilitySettingOption := appstoreconnect.CapabilityOption{Key: protectionLVL}
	capabilitySetting := appstoreconnect.CapabilitySetting{
		Options: []appstoreconnect.CapabilityOption{capabilitySettingOption},
		Key:     appstoreconnect.DataProtectionPermissionLevel,
	}
	return updateAppService(client, capabilityID, appstoreconnect.DataProtection, []appstoreconnect.CapabilitySetting{capabilitySetting})
}
