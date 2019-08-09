package autoprovision

import (
	"errors"

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

func iCloudEquals(cap appstoreconnect.BundleIDCapability) bool {
	if len(cap.Attributes.Settings) != 1 {
		return false
	}

	capSett := cap.Attributes.Settings[0]
	if capSett.Key != appstoreconnect.IcloudVersion {
		return false
	}
	if len(capSett.Options) != 1 {
		return false
	}

	capSettOpt := capSett.Options[0]
	if capSettOpt.Key != appstoreconnect.Xcode6 {
		return false
	}
	return true
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
		return iCloudEquals(cap), nil
	} else if capType == appstoreconnect.DataProtection {
		entVal, err := serialized.Object(e).String(entKey)
		if err != nil {
			return false, err
		}
		return dataProtectionEquals(entVal, cap)
	}

	return true, nil
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
