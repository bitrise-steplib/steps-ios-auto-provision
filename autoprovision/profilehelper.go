package autoprovision

import (
	"fmt"
	"github.com/bitrise-io/xcode-project/pretty"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func profileName(profileType appstoreconnect.ProfileType, bundleID string) (string, error) {
	var distr string
	switch profileType {
	case appstoreconnect.IOSAppStore, appstoreconnect.TvOSAppStore:
		distr = "app-store"
	case appstoreconnect.IOSAppAdHoc, appstoreconnect.TvOSAppAdHoc:
		distr = "ad-hoc"
	case appstoreconnect.IOSAppInHouse, appstoreconnect.TvOSAppInHouse:
		distr = "enterprise"
	case appstoreconnect.IOSAppDevelopment, appstoreconnect.TvOSAppDevelopment:
		distr = "development"
	default:
		return "", fmt.Errorf("unsupported profileType: %s, supported: IOS_APP_*, TVOS_APP_*", profileType)
	}
	return fmt.Sprintf("Bitrise %s - (%s)", distr, bundleID), nil
}

// FindProfile ...
func FindProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleIDIdentifier string) (*appstoreconnect.Profile, error) {
	name, err := profileName(profileType, bundleIDIdentifier)
	if err != nil {
		return nil, err
	}

	opt := &appstoreconnect.ListProfilesOptions{
		FilterProfileState: appstoreconnect.Active,
		FilterProfileType:  profileType,
		FilterName:         name,
		Limit:              1,
	}

	r, err := client.Provisioning.ListProfiles(opt)
	if err != nil {
		return nil, err
	}
	if len(r.Data) == 0 {
		return nil, nil
	}

	return &r.Data[0], nil
}

func checkProfileEntitlements(client *appstoreconnect.Client, prof appstoreconnect.Profile, entitlements Entitlement) (bool, error) {
	bundleIDresp, err := client.Provisioning.BundleID(prof.Relationships.BundleID.Links.Related)
	if err != nil {
		return false, err
	}
	bundleID := bundleIDresp.Data

	capabilitiesResp, err := client.Provisioning.Capabilities(bundleID.Relationships.Capabilities.Links.Related)
	if err != nil {
		return false, err
	}

	for k, v := range entitlements {
		ent := Entitlement{k: v}

		found := false
		for _, cap := range capabilitiesResp.Data {
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

func checkProfileCertificates(client *appstoreconnect.Client, prof appstoreconnect.Profile, certificateIDs []string) (bool, error) {
	ceretificatesResp, err := client.Provisioning.Certificates(prof.Relationships.Certificates.Links.Related)
	if err != nil {
		return false, err
	}

	ids := map[string]bool{}
	for _, cert := range ceretificatesResp.Data {
		ids[cert.ID] = true
	}
	for _, id := range certificateIDs {
		if !ids[id] {
			return false, nil
		}
	}
	return true, nil
}

func checkProfileDevices(client *appstoreconnect.Client, prof appstoreconnect.Profile, deviceIDs []string) (bool, error) {
	devicesResp, err := client.Provisioning.Devices(prof.Relationships.Devices.Links.Related)
	if err != nil {
		return false, err
	}

	ids := map[string]bool{}
	for _, dev := range devicesResp.Data {
		ids[dev.ID] = true
	}
	for _, id := range deviceIDs {
		if !ids[id] {
			return false, nil
		}
	}
	return true, nil
}

// CheckProfile ...
func CheckProfile(client *appstoreconnect.Client, prof appstoreconnect.Profile, entitlements Entitlement, deviceIDs, certificateIDs []string) (bool, error) {
	fmt.Println(pretty.Object(prof))

	if ok, err := checkProfileEntitlements(client, prof, entitlements); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	if ok, err := checkProfileCertificates(client, prof, certificateIDs); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	return checkProfileDevices(client, prof, deviceIDs)
}

// DeleteProfile ...
func DeleteProfile(client *appstoreconnect.Client, id string) error {
	return client.Provisioning.DeleteProfile(id)
}

// CreateProfile ...
func CreateProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID BundleID, certificateIDs []string, deviceIDs []string) (*appstoreconnect.Profile, error) {
	name, err := profileName(profileType, bundleID.Attributes.Identifier)
	if err != nil {
		return nil, err
	}
	// Create new Bitrise profile on App Store Connect
	r, err := client.Provisioning.CreateProfile(
		appstoreconnect.NewProfileCreateRequest(
			profileType,
			name,
			bundleID.ID,
			certificateIDs,
			deviceIDs,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Manual %s provisioning profile for %s bundle ID, error: %s", profileType.ReadableString(), bundleID.Attributes.Identifier, err)
	}
	return &r.Data, nil
}
