package autoprovision

import (
	"errors"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// Platform ...
type Platform string

// Const
const (
	IOS   Platform = "iOS"
	TVOS  Platform = "tvOS"
	MacOS Platform = "macOS"
)

// BundleIDPlatform ...
func (p Platform) BundleIDPlatform() (*appstoreconnect.BundleIDPlatform, error) {
	var apiPlatform appstoreconnect.BundleIDPlatform
	switch p {
	case IOS, TVOS:
		apiPlatform = appstoreconnect.IOS
	case MacOS:
		apiPlatform = appstoreconnect.MacOS
	default:
		return nil, errors.New("unknown platform: " + string(p))
	}
	return &apiPlatform, nil
}

// ProfileTypeToPlatform ...
var ProfileTypeToPlatform = map[appstoreconnect.ProfileType]Platform{
	appstoreconnect.IOSAppDevelopment: IOS,
	appstoreconnect.IOSAppStore:       IOS,
	appstoreconnect.IOSAppAdHoc:       IOS,
	appstoreconnect.IOSAppInHouse:     IOS,

	appstoreconnect.MacAppDevelopment: MacOS,
	appstoreconnect.MacAppStore:       MacOS,
	appstoreconnect.MacAppDirect:      MacOS,

	appstoreconnect.TvOSAppDevelopment: TVOS,
	appstoreconnect.TvOSAppStore:       TVOS,
	appstoreconnect.TvOSAppAdHoc:       TVOS,
	appstoreconnect.TvOSAppInHouse:     TVOS,
}

// ProfileTypeToDistribution ...
var ProfileTypeToDistribution = map[appstoreconnect.ProfileType]DistributionType{
	appstoreconnect.IOSAppDevelopment: Development,
	appstoreconnect.IOSAppStore:       AppStore,
	appstoreconnect.IOSAppAdHoc:       AdHoc,
	appstoreconnect.IOSAppInHouse:     Enterprise,

	appstoreconnect.MacAppDevelopment: Development,
	appstoreconnect.MacAppStore:       AppStore,
	appstoreconnect.MacAppDirect:      Direct,

	appstoreconnect.TvOSAppDevelopment: Development,
	appstoreconnect.TvOSAppStore:       AppStore,
	appstoreconnect.TvOSAppAdHoc:       AdHoc,
	appstoreconnect.TvOSAppInHouse:     Enterprise,
}

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
