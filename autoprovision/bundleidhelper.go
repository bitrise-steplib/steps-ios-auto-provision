package autoprovision

import (
	"fmt"
	"strings"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// FindBundleID ...
func FindBundleID(client *appstoreconnect.Client, bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
	var nextPageURL string
	var bundleIDs []appstoreconnect.BundleID
	for {
		response, err := client.Provisioning.ListBundleIDs(&appstoreconnect.ListBundleIDsOptions{
			PagingOptions: appstoreconnect.PagingOptions{
				Limit: 20,
				Next:  nextPageURL,
			},
			FilterIdentifier: bundleIDIdentifier,
		})
		if err != nil {
			return nil, err
		}

		bundleIDs = append(bundleIDs, response.Data...)

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			break
		}
	}

	if len(bundleIDs) == 0 {
		return nil, nil
	}

	// The FilterIdentifier works as a Like command. It will not search for the exact match,
	// this is why we need to find the exact match in the list.
	for _, d := range bundleIDs {
		if d.Attributes.Identifier == bundleIDIdentifier {
			return &d, nil
		}
	}
	return nil, nil
}

func checkBundleIDEntitlements(bundleIDEntitlements []appstoreconnect.BundleIDCapability, projectEntitlements Entitlement) (bool, error) {
	for k, v := range projectEntitlements {
		ent := Entitlement{k: v}

		if !ent.AppearsOnDeveloperPortal() {
			continue
		}

		found := false
		for _, cap := range bundleIDEntitlements {
			equal, err := ent.Equal(cap)
			if err != nil {
				return false, err
			}

			if equal {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

// CheckBundleIDEntitlements checks if a given Bundle ID has every capability enabled, required by the project.
func CheckBundleIDEntitlements(client *appstoreconnect.Client, bundleID appstoreconnect.BundleID, projectEntitlements Entitlement) (bool, error) {
	response, err := client.Provisioning.Capabilities(
		bundleID.Relationships.Capabilities.Links.Related)

	if err != nil {
		return false, err
	}

	return checkBundleIDEntitlements(response.Data, projectEntitlements)
}

// SyncBundleID ...
func SyncBundleID(client *appstoreconnect.Client, bundleIDID string, entitlements Entitlement) error {
	for key, value := range entitlements {
		ent := Entitlement{key: value}
		cap, err := ent.Capability()
		if err != nil {
			return err
		}
		if cap == nil {
			continue
		}

		body := appstoreconnect.BundleIDCapabilityCreateRequest{
			Data: appstoreconnect.BundleIDCapabilityCreateRequestData{
				Attributes: appstoreconnect.BundleIDCapabilityCreateRequestDataAttributes{
					CapabilityType: cap.Attributes.CapabilityType,
					Settings:       cap.Attributes.Settings,
				},
				Relationships: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationships{
					BundleID: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleID{
						Data: appstoreconnect.BundleIDCapabilityCreateRequestDataRelationshipsBundleIDData{
							ID:   bundleIDID,
							Type: "bundleIds",
						},
					},
				},
				Type: "bundleIdCapabilities",
			},
		}
		_, err = client.Provisioning.EnableCapability(body)
		if err != nil {
			return err
		}
	}

	return nil
}

func appIDName(bundleID string) string {
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return "Bitrise " + r.Replace(bundleID)
}

// CreateBundleID ...
func CreateBundleID(client *appstoreconnect.Client, bundleIDIdentifier string) (*appstoreconnect.BundleID, error) {
	appIDName := appIDName(bundleIDIdentifier)

	r, err := client.Provisioning.CreateBundleID(
		appstoreconnect.BundleIDCreateRequest{
			Data: appstoreconnect.BundleIDCreateRequestData{
				Attributes: appstoreconnect.BundleIDCreateRequestDataAttributes{
					Identifier: bundleIDIdentifier,
					Name:       appIDName,
					Platform:   appstoreconnect.IOS,
				},
				Type: "bundleIds",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register AppID for bundleID (%s): %s", bundleIDIdentifier, err)
	}

	return &r.Data, nil
}
