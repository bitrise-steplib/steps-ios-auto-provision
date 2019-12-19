package autoprovision

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func Test_profileName(t *testing.T) {
	tests := []struct {
		profileType appstoreconnect.ProfileType
		bundleID    string
		want        string
		wantErr     bool
	}{
		{
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS ad-hoc - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.IOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise iOS enterprise - (io.bitrise.app)",
			wantErr:     false,
		},

		{
			profileType: appstoreconnect.MacAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise macOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.MacAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise macOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.MacAppDirect,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise macOS direct - (io.bitrise.app)",
			wantErr:     false,
		},

		{
			profileType: appstoreconnect.TvOSAppDevelopment,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS development - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppStore,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS app-store - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppAdHoc,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS ad-hoc - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.TvOSAppInHouse,
			bundleID:    "io.bitrise.app",
			want:        "Bitrise tvOS enterprise - (io.bitrise.app)",
			wantErr:     false,
		},
		{
			profileType: appstoreconnect.ProfileType("unknown"),
			bundleID:    "io.bitrise.app",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.profileType), func(t *testing.T) {
			got, err := profileName(tt.profileType, tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("profileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("profileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
