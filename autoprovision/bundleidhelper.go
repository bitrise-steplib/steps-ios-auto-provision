package autoprovision

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/xcode-project/xcodeproj"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// BundleID ...
type BundleID struct {
	Attributes   appstoreconnect.BundleIDAttributes
	Capabilities []appstoreconnect.BundleIDCapability
	// Profiles     []appstoreconnect.Profile
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

func syncAppServices() {
	// targetEntitlements, err := projectHelper.targetEntitlements(t.Name, configurationName)
	// if err != nil && !serialized.IsKeyNotFoundError(err) {
	// 	return fmt.Errorf("failed to get target's  (%s), entitlement list, error: %s", t.Name, err)
	// }

	// // fmt.Printf("Target: %s entitlements: %v", t.Name, targetEntitlements)
	// for _, targetEnt := range targetEntitlements {

	// }
}
