package autoprovision

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/xcode-project/serialized"

	"github.com/bitrise-io/xcode-project/xcodeproj"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// BundleID ...
type BundleID struct {
	Attributes   appstoreconnect.BundleIDAttributes
	Capabilities []appstoreconnect.BundleIDCapability
	// Profiles     []appstoreconnect.Profile
	ID string
}

// EnsureApp search for AppID on the developer portal for the provided target's bundleID.
// If the target is not executable (not app, extension or UITest), then it returns nil
// If the AppID is available in the developer portal, it will return it
// If it's not, it will generate a new one with a name of:
// Bitrise {bundleID} {targetID}. `Example: auto_provision.ios-simple-objc + bc7cd9d1cc241639c4457975fefd920f => Bitrise auto provision ios simple objc bc7cd9d1cc241639c4457975fefd920f`
func EnsureApp(client *appstoreconnect.Client, projectHelper ProjectHelper, target xcodeproj.Target, configurationName string) (*BundleID, error) {
	// Check only executable targets which need to be registered on the Dev. Portal
	if !target.IsExecutableProduct() {
		return nil, nil
	}

	platform := func(projectPlatform Platform) appstoreconnect.BundleIDPlatform {
		switch projectPlatform {
		case IOS, TVOS:
			return appstoreconnect.IOS
		case MacOS:
			return appstoreconnect.MacOS
		default:
			return appstoreconnect.IOS
		}
	}(projectHelper.Platform)

	targetBundleID, err := projectHelper.TargetBundleID(target.Name, configurationName)
	if err != nil {
		return nil, fmt.Errorf("failed to find target's (%s) bundleID, error: %s", target.Name, err)
	}
	log.Printf("Search for AppID for the %s bundleID", targetBundleID)
	// Search AppID
	b, err := fetchBundleID(client, targetBundleID)
	if err != nil {
		return nil, err
	}
	if b != nil {
		return b, nil
	}
	log.Warnf("No AppID was found with bundleID: %s", target.Name)

	// Generate AppID name from the target bundleID and from targetID
	// auto_provision.ios-simple-objc + bc7cd9d1cc241639c4457975fefd920f => Bitrise auto provision ios simple objc bc7cd9d1cc241639c4457975fefd920f
	appIDName := appIDNameFrom(targetBundleID, target.ID)
	log.Printf("Registering AppID: %s with bundle id: %s", appIDName, targetBundleID)

	// No AppID found with the target's bundleID
	// Register AppID
	r, err := client.Provisioning.CreateBundleID(
		appstoreconnect.BundleIDCreateRequest{
			Data: appstoreconnect.BundleIDCreateRequestData{
				Attributes: appstoreconnect.BundleIDCreateRequestDataAttributes{
					Identifier: targetBundleID,
					Name:       appIDName,
					Platform:   platform,
				},
				Type: "bundleIds",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register AppID for bundleID %s, error: %s", targetBundleID, err)
	}

	capabilities, err := fetchBundleIDCapabilities(client, r.Data)
	if err != nil {
		return nil, err
	}

	b = &BundleID{
		Attributes: appstoreconnect.BundleIDAttributes{
			Identifier: r.Data.Attributes.Identifier,
			Name:       r.Data.Attributes.Name,
			Platform:   r.Data.Attributes.Platform,
		},
		Capabilities: capabilities,
		ID:           b.ID,
	}
	return b, nil
}

func appIDNameFrom(bundleID, targetID string) string {
	replacer := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return fmt.Sprintf("Bitrise %s %s", replacer.Replace(bundleID), targetID)
}

func fetchBundleID(client *appstoreconnect.Client, bundleIDIdentifier string) (*BundleID, error) {
	r, err := client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{
		FilterIdentifier: bundleIDIdentifier,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bundleID: %s, error: %s", bundleIDIdentifier, err)
	}

	if len(r.Data) == 0 {
		return nil, nil
	}

	// The FilterIdentifier works as a Like command. It will not search for the exact match,
	// this is why we need to find the exact match in the list.
	var b *BundleID
	for i, d := range r.Data {
		if d.Attributes.Identifier == bundleIDIdentifier {
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
// If the capability is not ethen enables it.
func syncAppServices(client *appstoreconnect.Client, projectHelper ProjectHelper, target xcodeproj.Target, configurationName string, bundleID BundleID) error {
	targetEntitlements, err := projectHelper.targetEntitlements(target.Name, configurationName)
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		return fmt.Errorf("failed to get target's  (%s), entitlement list, error: %s", target.Name, err)
	}

	for targetEntKey := range targetEntitlements {
		var capabilityEnabled bool
		for _, bundleIDCap := range bundleID.Capabilities {
			if string(bundleIDCap.Attributes.CapabilityType) == string(appstoreconnect.ServiceTypeByKey[targetEntKey]) {
				log.Printf("Capability %s, is already enabled", string(appstoreconnect.ServiceTypeByKey[targetEntKey]))
				capabilityEnabled = true
			}
		}

		// Enable capability
		if !capabilityEnabled && string(appstoreconnect.ServiceTypeByKey[targetEntKey]) != "" {
			log.Warnf("Capability %s is not enabled", string(appstoreconnect.ServiceTypeByKey[targetEntKey]))
			log.Printf("Enabling capability")

			if err := enableAppService(client, appstoreconnect.ServiceTypeByKey[targetEntKey], bundleID, nil); err != nil {
				return fmt.Errorf("failed to enable capability %v for target: %s, error: %s", appstoreconnect.ServiceTypeByKey[targetEntKey], target.Name, err)
			}
			log.Donef("Capability enabled")
		}
	}

	//
	// Data Protection
	if targetDataProtectionValue, err := targetEntitlements.String("com.apple.developer.default-data-protection"); err != nil && !serialized.IsKeyNotFoundError(err) {
		return fmt.Errorf("failed to get target's data procetion entitlement, error: %s", err)
	} else if targetDataProtectionValue != "" {
		if err := updateDataProtection(client, bundleID, targetDataProtectionValue); err != nil {
			return err
		}
	}

	//
	// iCloud
	usesICloudDocuments, usesICloudKit, usesICloudKeyValueStorage, err := usesICloudServices(targetEntitlements)
	if err != nil {
		return fmt.Errorf("failed to check if iCloud capability is enabled for bundleID: %s, error: %s", bundleID.Attributes.Identifier, err)
	}

	if usesICloudKeyValueStorage || usesICloudDocuments || usesICloudKit {
		if err := updateICloud(client, bundleID, string(appstoreconnect.Xcode6)); err != nil {
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

func enableAppService(client *appstoreconnect.Client, capabilityType appstoreconnect.CapabilityType, bundleID BundleID, settings []appstoreconnect.CapabilitySetting) error {
	_, err := client.Provisioning.EnableCapability(appstoreconnect.BundleIDCapabilityCreateRequest{
		Data: appstoreconnect.BundleIDCapabilityCreateRequestData{
			Attributes: appstoreconnect.BundleIDCapabilityCreateRequestDataAttributes{
				CapabilityType: capabilityType,
				Settings:       settings,
			},
			Relationships: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationships{
				BundleID: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleID{
					Data: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData{
						ID:   bundleID.ID,
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

func updateICloud(client *appstoreconnect.Client, bundleID BundleID, targetICloudVersion string) error {
	iCLoudVersion, iCloudCapabilityID := func() (appstoreconnect.CapabilityOptionKey, string) {
		for _, bundleIDCap := range bundleID.Capabilities {
			if bundleIDCap.Attributes.CapabilityType == appstoreconnect.ICloud {
				for _, settings := range bundleIDCap.Attributes.Settings {
					if settings.Key == appstoreconnect.IcloudVersion {
						return settings.Options[0].Key, bundleIDCap.ID
					}
				}
			}
		}
		return "", ""
	}()

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

func updateDataProtection(client *appstoreconnect.Client, bundleID BundleID, targetProtectionValue string) error {
	protectionCapabilityValue, protectionCapabilityID := func() (appstoreconnect.CapabilityOptionKey, string) {
		for _, bundleIDCap := range bundleID.Capabilities {
			if bundleIDCap.Attributes.CapabilityType == appstoreconnect.DataProtection {
				for _, settings := range bundleIDCap.Attributes.Settings {
					if settings.Key == appstoreconnect.DataProtectionPermissionLevel {
						return settings.Options[0].Key, bundleIDCap.ID
					}
				}
			}
		}
		return "", ""
	}()

	switch targetProtectionValue {
	case "NSFileProtectionComplete":
		if protectionCapabilityValue == appstoreconnect.CompleteProtection {
			log.Printf("Data Protection: complete already set")
		} else {
			log.Successf("Set Data Protection: complete")
			if err := updateDataProtectionLVL(client, protectionCapabilityID, appstoreconnect.CompleteProtection); err != nil {
				return fmt.Errorf("failed to update Data Protection Cabability, error: %s", err)
			}
		}
	case "NSFileProtectionCompleteUnlessOpen":
		if protectionCapabilityValue == appstoreconnect.ProtectedUnlessOpen {
			log.Printf("Data Protection: unless_open already set")
		} else {
			log.Successf("Set Data Protection: unless_open")
			if err := updateDataProtectionLVL(client, protectionCapabilityID, appstoreconnect.ProtectedUnlessOpen); err != nil {
				return fmt.Errorf("failed to update Data Protection Cabability, error: %s", err)
			}
		}
	case "NSFileProtectionCompleteUntilFirstUserAuthentication":
		if protectionCapabilityValue == appstoreconnect.ProtectedUntilFirstUserAuth {
			log.Printf("Data Protection: until_first_auth already set")
		} else {
			log.Successf("Set Data Protection: until_first_auth")
			if err := updateDataProtectionLVL(client, protectionCapabilityID, appstoreconnect.ProtectedUntilFirstUserAuth); err != nil {
				return fmt.Errorf("failed to update Data Protection Cabability, error: %s", err)
			}
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
