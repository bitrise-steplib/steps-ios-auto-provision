package autoprovision

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// PlatformToProfileTypeByDistribution ...
var PlatformToProfileTypeByDistribution = map[Platform]map[DistributionType]appstoreconnect.ProfileType{
	IOS: map[DistributionType]appstoreconnect.ProfileType{
		Development: appstoreconnect.IOSAppDevelopment,
		AppStore:    appstoreconnect.IOSAppStore,
		AdHoc:       appstoreconnect.IOSAppAdHoc,
		Enterprise:  appstoreconnect.IOSAppInHouse,
	},
	TVOS: map[DistributionType]appstoreconnect.ProfileType{
		Development: appstoreconnect.TvOSAppDevelopment,
		AppStore:    appstoreconnect.TvOSAppStore,
		AdHoc:       appstoreconnect.TvOSAppAdHoc,
		Enterprise:  appstoreconnect.TvOSAppInHouse,
	},
}

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
	return CheckBundleIDEntitlements(client, bundleIDresp.Data, entitlements)
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
func CreateProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID appstoreconnect.BundleID, certificateIDs []string, deviceIDs []string) (*appstoreconnect.Profile, error) {
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

// WriteProfile writes the provided profile under the `$HOME/Library/MobileDevice/Provisioning Profiles` directory.
// Xcode uses profiles located in that directory.
// The file extension depends on the profile's platform `IOS` => `.mobileprovision`, `MAC_OS` => `.provisionprofile`
func WriteProfile(profile appstoreconnect.Profile) error {
	homeDir := os.Getenv("HOME")
	profilesDir := path.Join(homeDir, "Library/MobileDevice/Provisioning Profiles")
	if exists, err := pathutil.IsDirExists(profilesDir); err != nil {
		return fmt.Errorf("failed to check directory for provisioning profiles (%s), error: %s", profilesDir, err)
	} else if !exists {
		if err := os.MkdirAll(profilesDir, 0600); err != nil {
			return fmt.Errorf("failed to generate directory for provisioning profiles (%s), error: %s", profilesDir, err)
		}
	}

	var ext string
	if profile.Attributes.Platform == appstoreconnect.IOS {
		ext = ".mobileprovision"
	} else if profile.Attributes.Platform == appstoreconnect.MacOS {
		ext = ".provisionprofile"
	} else {
		return fmt.Errorf("failed to write profile to file, unsupported platform: (%s). Supported platforms: %s, %s", profile.Attributes.Platform, appstoreconnect.IOS, appstoreconnect.MacOS)
	}

	b, err := base64.StdEncoding.DecodeString(profile.Attributes.ProfileContent)
	if err != nil {
		return fmt.Errorf("failed to decode ( base 64 ) the profile content, error: %s", err)
	}
	name := path.Join(profilesDir, profile.Attributes.UUID+ext)
	if err := ioutil.WriteFile(name, b, 0600); err != nil {
		return fmt.Errorf("failed to write profile to file, error: %s", err)
	}
	return nil
}
