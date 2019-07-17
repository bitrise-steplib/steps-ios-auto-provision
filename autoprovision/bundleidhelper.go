package autoprovision

import (
	"fmt"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// BundleID ...
type BundleID struct {
	Attributes   appstoreconnect.BundleIDAttributes
	Capabilities []appstoreconnect.BundleIDCapability
	// Profiles     []appstoreconnect.Profile
}

// EnsureApp ...
func EnsureApp(client *appstoreconnect.Client /*projectHelper ProjectHelper, devices, certificates */) {

	// TODO: For loop of projectHelper.targets
	fetchBundleID(client)

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

	// The FilterIdentifier works as a Like command. It will not search for the exact match, this is why we need to
	// find the exact match in the list.
	for _, d := range r.Data {
		if d.Attributes.Identifier == bundleID {
			return &BundleID{
				Attributes: appstoreconnect.BundleIDAttributes{
					Identifier: r.Data[0].Attributes.Identifier,
					Name:       r.Data[0].Attributes.Name,
					Platform:   r.Data[0].Attributes.Platform,
				},
			}, nil
		}
	}
	return nil, nil
}
