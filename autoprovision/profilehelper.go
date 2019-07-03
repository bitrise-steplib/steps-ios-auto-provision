package autoprovision

import (
	"fmt"

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

// EnsureProfile ...
// func EnsureProfile(platform Platform, distributionType DistributionType, capabilities []string, devices []Device, managed bool, expire time.Time) error {

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

// EnsureProfile ...
func EnsureProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID string,
	capabilityIDs []string, deviceIDs []string, isXcodeManaged, generateProfiles bool) (*Profile, error) {

	profile, err := fetchProfile(client, profileType, bundleID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		return profile, nil
	}

	// validate profile
	if isXcodeManaged && generateProfiles {
		log.Warnf("Project uses Xcode managed signing, but generate_profiles set to true, trying to generate Provisioning Profiles")
	}

	return nil, nil
}

func ensureManualProfiles() {

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
