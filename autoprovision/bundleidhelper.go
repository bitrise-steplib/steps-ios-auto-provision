package autoprovision

import (
	"fmt"
	"strings"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// BundleID ...
type BundleID struct {
	Attributes   appstoreconnect.BundleIDAttributes
	Capabilities []appstoreconnect.BundleIDCapability
	ID           string
}

// FindBundleID ...
func FindBundleID(client *appstoreconnect.Client, bundleIDIdentifier string) (*BundleID, error) {
	return fetchBundleID(client, bundleIDIdentifier)
}

// CheckBundleID ...
func CheckBundleID(bundleID BundleID, entitlements Entitlement) (bool, error) {
	for k, v := range entitlements {
		ent := Entitlement{k: v}

		found := false
		for _, cap := range bundleID.Capabilities {
			equal, err := ent.Equal(cap)
			if err != nil {
				return false, err
			}

			if equal {
				found = true
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

// SyncBundleID ...
func SyncBundleID(client *appstoreconnect.Client, bundleIDID string, entitlements Entitlement) error {
	_, err := setCapabilities(client, bundleIDID, entitlements)
	return err
}

// CreateBundleID ...
func CreateBundleID(client *appstoreconnect.Client, bundleIDIdentifier string, entitlements Entitlement) (*BundleID, error) {
	return createBundleID(client, IOS, bundleIDIdentifier)
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
	var bundleID *BundleID
	for _, d := range r.Data {
		if d.Attributes.Identifier == bundleIDIdentifier {
			bundleID = &BundleID{
				Attributes: appstoreconnect.BundleIDAttributes{
					Identifier: d.Attributes.Identifier,
					Name:       d.Attributes.Name,
					Platform:   d.Attributes.Platform,
				},
				ID: d.ID,
			}

			caps, err := fetchBundleIDCapabilities(client, d)
			if err != nil {
				return nil, err
			}

			bundleID.Capabilities = caps

			break
		}
	}
	return bundleID, nil

}

func fetchBundleIDCapabilities(client *appstoreconnect.Client, bundleID appstoreconnect.BundleID) ([]appstoreconnect.BundleIDCapability, error) {
	r, err := client.Provisioning.Capabilities(bundleID.Relationships.Capabilities.Links.Related)
	if err != nil {
		return nil, err
	}

	var caps []appstoreconnect.BundleIDCapability
	for _, d := range r.Data {
		caps = append(caps, appstoreconnect.BundleIDCapability{
			Attributes: appstoreconnect.BundleIDCapabilityAttributes{
				CapabilityType: d.Attributes.CapabilityType,
				Settings:       d.Attributes.Settings,
			},
			ID: d.ID,
		})
	}
	return caps, nil
}

func setCapabilities(client *appstoreconnect.Client, bundleIDID string, entitlements Entitlement) ([]appstoreconnect.BundleIDCapability, error) {
	var caps []appstoreconnect.BundleIDCapability

	for key, value := range entitlements {
		ent := Entitlement{key: value}
		cap, err := ent.Capability()
		if err != nil {
			return nil, err
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
							Type: "",
						},
					},
				},
			},
		}
		r, err := client.Provisioning.EnableCapability(body)
		if err != nil {
			return nil, err
		}

		caps = append(caps, appstoreconnect.BundleIDCapability{
			Attributes: appstoreconnect.BundleIDCapabilityAttributes{
				CapabilityType: r.Data.Attributes.CapabilityType,
				Settings:       r.Data.Attributes.Settings,
			},
		})

	}
	return caps, nil
}

func appIDName(bundleID string) string {
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return "Bitrise " + r.Replace(bundleID)
}

func createBundleID(client *appstoreconnect.Client, platform Platform,
	bundleIDIdentifier string) (*BundleID, error) {

	bundleIDPlatform, err := platform.BundleIDPlatform()
	if err != nil {
		return nil, err
	}

	appIDName := appIDName(bundleIDIdentifier)

	r, err := client.Provisioning.CreateBundleID(
		appstoreconnect.BundleIDCreateRequest{
			Data: appstoreconnect.BundleIDCreateRequestData{
				Attributes: appstoreconnect.BundleIDCreateRequestDataAttributes{
					Identifier: bundleIDIdentifier,
					Name:       appIDName,
					Platform:   *bundleIDPlatform,
				},
				Type: "bundleIds",
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register AppID for bundleID %s, error: %s", bundleIDIdentifier, err)
	}

	return &BundleID{
		Attributes: appstoreconnect.BundleIDAttributes{
			Identifier: r.Data.Attributes.Identifier,
			Name:       r.Data.Attributes.Name,
			Platform:   r.Data.Attributes.Platform,
		},
		ID: r.Data.ID,
	}, nil
}

func checkEntitlements(ents Entitlement, caps []appstoreconnect.BundleIDCapability) (bool, error) {
	for key, value := range ents {
		ent := Entitlement{key: value}

		found := false
		for _, cap := range caps {
			if equal, err := ent.Equal(cap); err != nil {
				return false, err
			} else if equal {
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

// EnsureBundleID ...
func EnsureBundleID(client *appstoreconnect.Client, platform Platform,
	bundleIDIdentifier string, entitlements Entitlement) (*BundleID, error) {

	bundleID, err := fetchBundleID(client, bundleIDIdentifier)
	if err != nil {
		return nil, err
	}
	if bundleID != nil {
		matching := true
		for key, value := range entitlements {
			ent := Entitlement{key: value}

			found := false
			for _, cap := range bundleID.Capabilities {
				if equal, err := ent.Equal(cap); err != nil {
					return nil, err
				} else if equal {
					found = true
					break
				}
			}

			if !found {
				matching = false
				break
			}
		}

		if matching {
			return bundleID, nil
		}

		// TODO: delete bundle id to reset the capabilities
	}

	bundleID, err = createBundleID(client, platform, bundleIDIdentifier)
	if err != nil {
		return nil, err
	}
	if entitlements != nil {
		caps, err := setCapabilities(client, bundleID.ID, entitlements)
		if err != nil {
			return nil, nil
		}
		bundleID.Capabilities = caps
	}

	return bundleID, nil
}
