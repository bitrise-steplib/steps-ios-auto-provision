package autoprovision

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func Test_checkBundleIDEntitlements(t *testing.T) {
	tests := []struct {
		name                 string
		bundleIDEntitlements []appstoreconnect.BundleIDCapability
		projectEntitlements  Entitlement
		want                 bool
		wantErr              bool
	}{
		{
			name:                 "Check known entitlements, which does not need to be registered on the Developer Portal",
			bundleIDEntitlements: []appstoreconnect.BundleIDCapability{},
			projectEntitlements: Entitlement(map[string]interface{}{
				"keychain-access-groups":                           "",
				"com.apple.developer.ubiquity-kvstore-identifier":  "",
				"com.apple.developer.icloud-container-identifiers": "",
			}),
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkBundleIDEntitlements(tt.bundleIDEntitlements, tt.projectEntitlements)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkBundleIDEntitlements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkBundleIDEntitlements() = %v, want %v", got, tt.want)
			}
		})
	}
}
