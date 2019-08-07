package autoprovision

import (
	"errors"
	"fmt"

	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// Profile ...
type Profile struct {
	Attributes   appstoreconnect.ProfileAttributes
	Devices      []appstoreconnect.Device
	Certificates []appstoreconnect.Certificate
	BundleID     BundleID
	ID           string
}

// FindProfile ...
func FindProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleIDIdentifier string) (*Profile, error) {
	return fetchProfile(client, profileType, bundleIDIdentifier)
}

// CheckProfile ...
func CheckProfile(prof Profile, entitlements Entitlement, devices []appstoreconnect.Device, certificates []appstoreconnect.Certificate) (bool, error) {
	for k, v := range entitlements {
		ent := Entitlement{k: v}

		found := false
		for _, cap := range prof.BundleID.Capabilities {
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

	ids := map[string]bool{}
	for _, cert := range prof.Certificates {
		ids[cert.ID] = true
	}
	for _, cert := range certificates {
		if !ids[cert.ID] {
			return false, nil
		}
	}

	ids = map[string]bool{}
	for _, dev := range prof.Devices {
		ids[dev.ID] = true
	}
	for _, dev := range devices {
		if !ids[dev.ID] {
			return false, nil
		}
	}

	return true, nil
}

// DeleteProfile ...
func DeleteProfile(client *appstoreconnect.Client, prof Profile) error {
	return client.Provisioning.DeleteProfile(prof.ID)
}

// CreateProfile ...
func CreateProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleID BundleID, certificateIDs []string, deviceIDs []string) (*Profile, error) {
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
	profile := &Profile{
		Attributes: r.Data.Attributes,
	}
	return profile, nil
}

// fetchProfile fetches a Bitrise managed profile of the given profile type for the given bundle ID.
func fetchProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType, bundleIDIdentifier string) (*Profile, error) {
	name, err := profileName(profileType, bundleIDIdentifier)
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
		// BundleID:     bundleIDs[0],
	}

	return &profile, nil
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

func certificateIncluded(profile Profile, certificate appstoreconnect.Certificate) bool {
	for _, cert := range profile.Certificates {
		if cert.ID == certificate.ID {
			return true
		}
	}
	return false
}

func devicesIncluded(profile Profile, devices []appstoreconnect.Device) bool {
	for _, desiredDev := range devices {
		included := false
		for _, profileDev := range profile.Devices {
			if profileDev.ID == desiredDev.ID {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}

func ensureProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType,
	bundleID string, entitlements []Entitlement,
	devices []appstoreconnect.Device, certificates []appstoreconnect.Certificate) (*Profile, error) {

	profile, err := fetchProfile(client, profileType, bundleID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		matching := false
		if profile.BundleID.Attributes.Identifier != bundleID {
			matching = false
		}

		if matching {
			return profile, nil
		}
	}
	return nil, nil
}

// EnsureProfile ...
func EnsureProfile(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType,
	bundleID string, entitlements Entitlement,
	devices []appstoreconnect.Device, certificates []appstoreconnect.Certificate) ([]Profile, error) {
	var profiles []Profile

	p, err := ensureManualProfiles(client, profileType, bundleID, certificates, devices)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s Manual profile for %s, error: %s", profileType.ReadableString(), bundleID, err)
	}
	profiles = append(profiles, p...)

	return profiles, nil
}

func ensureManualProfiles(client *appstoreconnect.Client, profileType appstoreconnect.ProfileType,
	bundleID string, certificates []appstoreconnect.Certificate, devices []appstoreconnect.Device) ([]Profile, error) {

	// TODO
	return nil, nil
}

//func ensureManualProfile(client *appstoreconnect.Client, certificate appstoreconnect.Certificate, devices []appstoreconnect.Device, profileType appstoreconnect.ProfileType, bundleID string) (Profile, error) {
//
//	profile, err := fetchProfile(client, profileType, bundleID)
//	if err != nil {
//		return Profile{}, err
//	}
//
//	if profile != nil {
//		if validateProfile(*profile, certificate, devices) {
//			return *profile, nil
//		}
//	}
//
//	if profile == nil {
//		name, err := profileName(profileType, bundleID)
//		if err != nil {
//			return Profile{}, fmt.Errorf("failed to generate name for manual profile, error: %s", err)
//		}
//
//		bundleIDEntity, err := client.Provisioning.FetchBundleID(bundleID)
//		if err != nil {
//			return Profile{}, fmt.Errorf("failed to fetch entity ID for bundleIDL %s, error: %s", bundleID, err)
//		}
//
//		// Create new Bitrise profile on App Store Connect
//		profileResponse, err := client.Provisioning.CreateProfile(
//			appstoreconnect.NewProfileCreateRequest(
//				profileType,
//				name,
//				bundleIDEntity.ID,
//				[]appstoreconnect.Certificate{certificate},
//				devices,
//			),
//		)
//		if err != nil {
//			return Profile{}, fmt.Errorf("failed to create Manual %s provisioning profile for %s bundle ID, error: %s", profileType.ReadableString(), bundleID, err)
//		}
//		profile = &Profile{
//			Attributes: appstoreconnect.ProfileAttributes{
//				Name:           profileResponse.Data.Attributes.Name,
//				Platform:       profileResponse.Data.Attributes.Platform,
//				ProfileContent: profileResponse.Data.Attributes.ProfileContent,
//				UUID:           profileResponse.Data.Attributes.UUID,
//				CreatedDate:    profileResponse.Data.Attributes.CreatedDate,
//				ProfileState:   profileResponse.Data.Attributes.ProfileState,
//				ProfileType:    profileResponse.Data.Attributes.ProfileType,
//				ExpirationDate: profileResponse.Data.Attributes.ExpirationDate,
//			},
//			Devices: devices,
//			// BundleID:     bundleIDEntity,
//			Certificates: []appstoreconnect.Certificate{certificate},
//		}
//		if err != nil {
//			return Profile{}, fmt.Errorf("failed to generate %s manual profile for %s bundle ID, error: %s", profileType.ReadableString(), bundleID, err)
//		}
//	}
//	return *profile, nil
//}

func validateProfile(profile Profile, certificate appstoreconnect.Certificate, devices []appstoreconnect.Device) bool {
	// TODO check device list & certificate list. The capabilities are checked by apple already (valid / invalid state of the profile)
	for _, cert := range profile.Certificates {
		if cert.ID != certificate.ID {

		}
	}

	return false
}

func bundleIDAttributes(attributes serialized.Object) (*appstoreconnect.BundleIDAttributes, error) {
	name, err := attributes.String("name")
	if err != nil {
		return nil, errors.New("missing attribute: name")
	}
	identifier, err := attributes.String("identifier")
	if err != nil {
		return nil, errors.New("missing attribute: identifier")
	}
	platform, err := attributes.String("platform")
	if err != nil {
		return nil, errors.New("missing attribute: platform")
	}

	return &appstoreconnect.BundleIDAttributes{
		Name:       name,
		Identifier: identifier,
		Platform:   platform,
	}, nil
}

func deviceAttributes(attributes serialized.Object) (*appstoreconnect.DeviceAttributes, error) {
	addedDate, err := attributes.String("addedDate")
	if err != nil {
		return nil, errors.New("missing attribute: addedDate")
	}
	name, err := attributes.String("name")
	if err != nil {
		return nil, errors.New("missing attribute: name")
	}
	deviceClass, err := attributes.String("deviceClass")
	if err != nil {
		return nil, errors.New("missing attribute: deviceClass")
	}
	model, err := attributes.String("model")
	if err != nil {
		// model can be null
		// return nil, errors.New("missing attribute: model")
	}
	udid, err := attributes.String("udid")
	if err != nil {
		return nil, errors.New("missing attribute: udid")
	}
	platform, err := attributes.String("platform")
	if err != nil {
		return nil, errors.New("missing attribute: platform")
	}
	status, err := attributes.String("status")
	if err != nil {
		return nil, errors.New("missing attribute: status")
	}

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
	serialNumber, err := attributes.String("serialNumber")
	if err != nil {
		return nil, errors.New("missing attribute: serialNumber")
	}
	certificateContent, err := attributes.String("certificateContent")
	if err != nil {
		return nil, errors.New("missing attribute: certificateContent")
	}
	displayName, err := attributes.String("displayName")
	if err != nil {
		return nil, errors.New("missing attribute: displayName")
	}
	name, err := attributes.String("name")
	if err != nil {
		return nil, errors.New("missing attribute: name")
	}
	platform, err := attributes.String("platform")
	if err != nil {
		return nil, errors.New("missing attribute: platform")
	}
	expirationDate, err := attributes.String("expirationDate")
	if err != nil {
		return nil, errors.New("missing attribute: expirationDate")
	}
	certificateType, err := attributes.String("certificateType")
	if err != nil {
		return nil, errors.New("missing attribute: certificateType")
	}

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
