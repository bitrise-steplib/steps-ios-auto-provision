package autoprovision

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// Profile ...
type Profile struct {
	Attributes   appstoreconnect.ProfileAttributes
	Devices      []appstoreconnect.Device
	BundleID     appstoreconnect.BundleID
	Certificates []appstoreconnect.Certificate
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

// EnsureProfiles returns the profiles for the selected profile type.
// If the selected profile type is not a development one, it will return it's development pair too.
// For profile generation provide the devices to regeister to the developer / ad-hoc profile (devices registered on bitrise)
// and the certificates (development and distribution) which need to be included in the profiles
func EnsureProfiles(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID string,
	capabilityIDs []string, devices []appstoreconnect.Device, certificates []appstoreconnect.Certificate, isXcodeManaged, generateProfiles bool) ([]Profile, error) {
	var profiles []Profile

	profileTypes := []appstoreconnect.ProfileType{profileType}
	if developmentPair := profileType.ProfileTypeDevelopmentPair(); developmentPair != "" {
		profileTypes = append(profileTypes, developmentPair)
	}

	log.Debugf("distribution types: %s", strings.Join(
		func() []string {
			var s []string
			for _, pType := range profileTypes {
				s = append(s, pType.ReadableString())
			}
			return s
		}(), ","))

	// validate profile
	if isXcodeManaged && generateProfiles {
		log.Warnf("Project uses Xcode managed signing, but generate_profiles set to true, trying to generate Provisioning Profiles")

		var err error
		for _, profType := range profileTypes {
			var p Profile
			p, err = ensureManualProfile(client, profType, bundleID, certificates, devices)
			if err != nil {
				break
			}
			profiles = append(profiles, p)
		}

		// The manual profile generation failed
		// Rollback to Xcode Managed ones
		if err != nil {
			log.Errorf("generate_profiles set to true, but failed to generate Provisioning Profiles with error: %s", err)
			log.Infof("\nTrying to use Xcode managed Provisioning Profiles")

			for _, profType := range profileTypes {
				profiles := []Profile{} // Empty the manual profiles
				var p Profile
				p, err = ensureManagedProfile(client, profType, bundleID)
				if err != nil {
					return nil, fmt.Errorf("failed to get %s Xcode Managed profile for %s, error: %s", profType.ReadableString(), bundleID, err)
				}
				profiles = append(profiles, p)
			}
		}
	} else if isXcodeManaged {
		for _, profType := range profileTypes {
			p, err := ensureManagedProfile(client, profType, bundleID)
			if err != nil {
				return nil, fmt.Errorf("failed to get %s Xcode Managed profile for %s, error: %s", profType.ReadableString(), bundleID, err)
			}
			profiles = append(profiles, p)
		}
	} else {
		for _, profType := range profileTypes {
			p, err := ensureManualProfile(client, profType, bundleID, certificates, devices)
			if err != nil {
				return nil, fmt.Errorf("failed to get %s Manual profile for %s, error: %s", profType.ReadableString(), bundleID, err)
			}
			profiles = append(profiles, p)
		}
	}
	return profiles, nil

}

func ensureManualProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType,
	bundleID string, certificates []appstoreconnect.Certificate, devices []appstoreconnect.Device) (Profile, error) {

	profile, err := fetchProfile(client, profileType, bundleID)
	if err != nil {
		return Profile{}, err
	}
	if profile == nil {
		name, err := profileName(profileType, bundleID)
		if err != nil {
			return Profile{}, fmt.Errorf("failed to generate name for manual profile, error: %s", err)
		}

		// Create new Bitrise profile on App Store Connect
		profileResponse, err := client.Provisioning.CreateProfile(
			appstoreconnect.NewProfileCreateRequest(
				profileType,
				name,
				bundleID,
				certificates,
				devices,
			),
		)
		if err != nil {
			return Profile{}, fmt.Errorf("failed to create Manual %s provisioning profile for %s bundle ID, error: %s", profileType.ReadableString(), bundleID, err)
		}
		profile = &Profile{
			Attributes: appstoreconnect.ProfileAttributes{
				Name:           profileResponse.Data.Attributes.Name,
				Platform:       profileResponse.Data.Attributes.Platform,
				ProfileContent: profileResponse.Data.Attributes.ProfileContent,
				UUID:           profileResponse.Data.Attributes.UUID,
				CreatedDate:    profileResponse.Data.Attributes.CreatedDate,
				ProfileState:   profileResponse.Data.Attributes.ProfileState,
				ProfileType:    profileResponse.Data.Attributes.ProfileType,
				ExpirationDate: profileResponse.Data.Attributes.ExpirationDate,
			},
			Devices:      []appstoreconnect.Device{},
			BundleID:     appstoreconnect.BundleID{},
			Certificates: []appstoreconnect.Certificate{},
		}
		if err != nil {
			return Profile{}, fmt.Errorf("failed to generate %s manual profile for %s bundle ID, error: %s", profileType.ReadableString(), bundleID, err)
		}
	}
	// return Profile{}, fmt.Errorf("failed to find Manual provisioning profile with bundle id: %s for %s", bundleID, profileType)
	return *profile, nil
}

func ensureManagedProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID string) (Profile, error) {
	// TODO
	return Profile{}, nil
}

func validateProfile(profile Profile) []string {
	return nil
}

func fetchProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID string) (*Profile, error) {
	name, err := profileName(profileType, bundleID)
	if err != nil {
		return nil, err
	}

	opt := &appstoreconnect.ListProfilesOptions{
		FilterProfileState: appstoreconnect.Active,
		FilterProfileType:  profileType,
		FilterName:         name,
		Include:            "bundleId,certificates,devices",
		Limit:              1,
	}

	r, err := client.Provisioning.ListProfiles(opt)
	if err != nil {
		return nil, err
	}
	if len(r.Data) == 0 {
		return nil, nil
	}

	var devices []appstoreconnect.Device
	var bundleIDs []appstoreconnect.BundleID
	var certificates []appstoreconnect.Certificate

	if len(r.Included) > 0 {
		for _, v := range r.Included {
			switch v.Type {
			case "certificates":
				attributes, err := certificateAttributes(v.Attributes)
				if err != nil {
					return nil, err
				}

				certificates = append(certificates, appstoreconnect.Certificate{
					Attributes: *attributes,
					ID:         v.ID,
					Type:       v.Type,
				})
			case "devices":
				attributes, err := deviceAttributes(v.Attributes)
				if err != nil {
					return nil, err
				}

				devices = append(devices, appstoreconnect.Device{
					Attributes: *attributes,
					ID:         v.ID,
					Type:       v.Type,
				})
			case "bundleIds":
				attributes, err := bundleIDAttributes(v.Attributes)
				if err != nil {
					return nil, err
				}

				bundleIDs = append(bundleIDs, appstoreconnect.BundleID{
					Attributes: *attributes,
					ID:         v.ID,
					Type:       v.Type,
				})
			}
		}

	}

	profile := Profile{
		Attributes:   r.Data[0].Attributes,
		Certificates: certificates,
		Devices:      devices,
		BundleID:     bundleIDs[0],
	}

	return &profile, nil
}

func bundleIDAttributes(attributes serialized.Object) (*appstoreconnect.BundleIDAttributes, error) {
	name, _ := attributes.String("name")
	identifier, _ := attributes.String("identifier")
	platform, _ := attributes.String("platform")

	return &appstoreconnect.BundleIDAttributes{
		Name:       name,
		Identifier: identifier,
		Platform:   platform,
	}, nil
}

func deviceAttributes(attributes serialized.Object) (*appstoreconnect.DeviceAttributes, error) {
	addedDate, _ := attributes.String("addedDate")
	name, _ := attributes.String("name")
	deviceClass, _ := attributes.String("deviceClass")
	model, _ := attributes.String("model")
	udid, _ := attributes.String("udid")
	platform, _ := attributes.String("platform")
	status, _ := attributes.String("status")

	return &appstoreconnect.DeviceAttributes{
		AddedDate:   addedDate,
		Name:        name,
		DeviceClass: appstoreconnect.DeviceClass(deviceClass),
		Model:       model,
		UDID:        udid,
		Platform:    appstoreconnect.BundleIDPlatform(platform),
		Status:      appstoreconnect.Status(status),
	}, nil
}

func certificateAttributes(attributes serialized.Object) (*appstoreconnect.CertificateAttributes, error) {
	serialNumber, _ := attributes.String("serialNumber")
	certificateContent, _ := attributes.String("certificateContent")
	displayName, _ := attributes.String("displayName")
	name, _ := attributes.String("name")
	platform, _ := attributes.String("platform")
	expirationDate, _ := attributes.String("expirationDate")
	certificateType, _ := attributes.String("certificateType")

	return &appstoreconnect.CertificateAttributes{
		SerialNumber:       serialNumber,
		CertificateContent: certificateContent,
		DisplayName:        displayName,
		Name:               name,
		Platform:           appstoreconnect.BundleIDPlatform(platform),
		ExpirationDate:     expirationDate,
		CertificateType:    appstoreconnect.CertificateType(certificateType),
	}, nil
}

func ProfileDevices() {

}
