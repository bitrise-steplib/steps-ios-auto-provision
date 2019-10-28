package autoprovision

import (
	"errors"

	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// Entitlement ...
type Entitlement serialized.Object

// DataProtections ...
var DataProtections = map[string]appstoreconnect.CapabilityOptionKey{
	"NSFileProtectionComplete":                             appstoreconnect.CompleteProtection,
	"NSFileProtectionCompleteUnlessOpen":                   appstoreconnect.ProtectedUnlessOpen,
	"NSFileProtectionCompleteUntilFirstUserAuthentication": appstoreconnect.ProtectedUntilFirstUserAuth,
}

func iCloudEquals(ent Entitlement, cap appstoreconnect.BundleIDCapability) (bool, error) {
	documents, cloudKit, kvStorage, err := ent.iCloudServices()
	if err != nil {
		return false, err
	}

	if len(cap.Attributes.Settings) != 1 {
		return false, nil
	}

	capSett := cap.Attributes.Settings[0]
	if capSett.Key != appstoreconnect.IcloudVersion {
		return false, nil
	}
	if len(capSett.Options) != 1 {
		return false, nil
	}

	capSettOpt := capSett.Options[0]
	if (documents || cloudKit || kvStorage) && capSettOpt.Key != appstoreconnect.Xcode6 {
		return false, nil
	}
	return true, nil
}

func dataProtectionEquals(entVal string, cap appstoreconnect.BundleIDCapability) (bool, error) {
	key, ok := DataProtections[entVal]
	if !ok {
		return false, errors.New("no data protection level found for entitlement value: " + entVal)
	}

	if len(cap.Attributes.Settings) != 1 {
		return false, nil
	}

	capSett := cap.Attributes.Settings[0]
	if capSett.Key != appstoreconnect.DataProtectionPermissionLevel {
		return false, nil
	}
	if len(capSett.Options) != 1 {
		return false, nil
	}

	capSettOpt := capSett.Options[0]
	if capSettOpt.Key != key {
		return false, nil
	}
	return true, nil
}

// AppearsOnDeveloperPortal reports whether the given (project) Entitlement needs to be registered on Apple Developer Portal or not.
// List of services, to be registered: https://developer.apple.com/documentation/appstoreconnectapi/capabilitytype.
func (e Entitlement) AppearsOnDeveloperPortal() bool {
	if len(e) == 0 {
		return false
	}
	entKey := serialized.Object(e).Keys()[0]

	_, ok := appstoreconnect.ServiceTypeByKey[entKey]
	return ok
}

// Equal ...
func (e Entitlement) Equal(cap appstoreconnect.BundleIDCapability) (bool, error) {
	if len(e) == 0 {
		return false, nil
	}

	entKey := serialized.Object(e).Keys()[0]

	capType, ok := appstoreconnect.ServiceTypeByKey[entKey]
	if !ok {
		return false, errors.New("unknown entitlement key: " + entKey)
	}

	if cap.Attributes.CapabilityType != capType {
		return false, nil
	}

	if capType == appstoreconnect.ICloud {
		return iCloudEquals(e, cap)
	} else if capType == appstoreconnect.DataProtection {
		entVal, err := serialized.Object(e).String(entKey)
		if err != nil {
			return false, err
		}
		return dataProtectionEquals(entVal, cap)
	}

	return true, nil
}

func (e Entitlement) iCloudServices() (iCloudDocuments, iCloudKit, keyValueStorage bool, err error) {
	v, err := serialized.Object(e).String("com.apple.developer.ubiquity-kvstore-identifier")
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		return false, false, false, err
	}
	keyValueStorage = v != ""

	iCloudServices, err := serialized.Object(e).StringSlice("com.apple.developer.icloud-services")
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		return false, false, false, err
	}

	if len(iCloudServices) > 0 {
		iCloudDocuments = sliceutil.IsStringInSlice("CloudDocuments", iCloudServices)
		iCloudKit = sliceutil.IsStringInSlice("CloudKit", iCloudServices)
	}
	return
}

// Capability ...
func (e Entitlement) Capability() (*appstoreconnect.BundleIDCapability, error) {
	if len(e) == 0 {
		return nil, nil
	}

	entKey := serialized.Object(e).Keys()[0]

	capType, ok := appstoreconnect.ServiceTypeByKey[entKey]
	if !ok {
		return nil, errors.New("unknown entitlement key: " + entKey)
	}

	capSetts := []appstoreconnect.CapabilitySetting{}
	if capType == appstoreconnect.ICloud {
		capSett := appstoreconnect.CapabilitySetting{
			Key: appstoreconnect.IcloudVersion,
			Options: []appstoreconnect.CapabilityOption{
				appstoreconnect.CapabilityOption{
					Key: appstoreconnect.Xcode6,
				},
			},
		}
		capSetts = append(capSetts, capSett)
	} else if capType == appstoreconnect.DataProtection {
		entVal, err := serialized.Object(e).String(entKey)
		if err != nil {
			return nil, errors.New("no entitlements value for key: " + entKey)
		}

		key, ok := DataProtections[entVal]
		if !ok {
			return nil, errors.New("no data protection level found for entitlement value: " + entVal)
		}

		capSett := appstoreconnect.CapabilitySetting{
			Key: appstoreconnect.DataProtectionPermissionLevel,
			Options: []appstoreconnect.CapabilityOption{
				appstoreconnect.CapabilityOption{
					Key: key,
				},
			},
		}
		capSetts = append(capSetts, capSett)
	}

	return &appstoreconnect.BundleIDCapability{
		Attributes: appstoreconnect.BundleIDCapabilityAttributes{
			CapabilityType: capType,
			Settings:       capSetts,
		},
	}, nil
}
