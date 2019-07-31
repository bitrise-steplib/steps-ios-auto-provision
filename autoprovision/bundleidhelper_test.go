package autoprovision

import (
	"testing"

	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func TestEnsureApp(t *testing.T) {
	var err error
	schemeCases, targetCases, xcProjCases, projHelpCases, configCases, err = initTestCases()
	if err != nil {
		t.Fatalf("Failed to initialize test cases, error: %s", err)
	}

	tests := []struct {
		name     string
		client   *appstoreconnect.Client
		platform Platform
		bundleID string
		wantErr  bool
	}{
		{
			name:     "Ensure app for Xcode-10_default for config " + configCases[0],
			client:   appstoreconnect.InitTestClient(t),
			platform: Platform("iOS"),
			bundleID: "com.bitrise.Xcode-10-default",
			wantErr:  false,
		},
		{

			name:     "Ensure app for Xcode-10_defaultTests for config " + configCases[0],
			client:   appstoreconnect.InitTestClient(t),
			platform: Platform("iOS"),
			bundleID: "com.bitrise.Xcode-10-defaultTests",
			wantErr:  false,
		},
		{
			name:     "Ensure app for Xcode-10_defaultUITests for config " + configCases[0],
			client:   appstoreconnect.InitTestClient(t),
			platform: Platform("iOS"),
			bundleID: "com.bitrise.Xcode-10-defaultUITests",
			wantErr:  false,
		},

		{

			name:     "Ensure app for TV_OS for config " + configCases[0],
			client:   appstoreconnect.InitTestClient(t),
			platform: Platform("tvOS"),
			bundleID: "com.bitrise.TV-OS",
			wantErr:  false,
		},
		{

			name:     "Ensure app for TV_OSTests for config " + configCases[0],
			client:   appstoreconnect.InitTestClient(t),
			platform: Platform("tvOS"),
			bundleID: "com.bitrise.TV-OSTests",
			wantErr:  false,
		},
		{
			name:     "Ensure app for TV_OSUITests for config " + configCases[0],
			client:   appstoreconnect.InitTestClient(t),
			platform: Platform("tvOS"),
			bundleID: "com.bitrise.TV-OSUITests",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := EnsureApp(tt.client, tt.platform, tt.bundleID); (err != nil) != tt.wantErr {
				t.Errorf("EnsureApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fetchBundleID(t *testing.T) {
	client := appstoreconnect.InitTestClient(t)

	tests := []struct {
		name               string
		bundleIDIdentifier string
		wantErr            bool
	}{
		{
			name:               "Fetch bundleID com.bitrise.io.testing.firefox",
			bundleIDIdentifier: "com.bitrise.io.testing.firefox",
			wantErr:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchBundleID(client, tt.bundleIDIdentifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchBundleID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Attributes.Identifier != tt.bundleIDIdentifier {
				t.Errorf("fetchBundleID() wrong bundleID = %v, want with identifier %v", got, tt.bundleIDIdentifier)
				return
			}
		})
	}
}

func Test_appIDName(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     string
	}{
		{
			name:     "Generate AppID name for bundleID without (-,_)",
			bundleID: "com.bitrise.TV.OSUITests",
			want:     "Bitrise com bitrise TV OSUITests",
		},
		{
			name:     "Generate AppID name for bundleID with (-,_)",
			bundleID: "auto_provision.ios-simple-objc",
			want:     "Bitrise auto provision ios simple objc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appIDName(tt.bundleID); got != tt.want {
				t.Errorf("appIDNameFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncAppServices(t *testing.T) {
	var err error
	schemeCases, targetCases, xcProjCases, projHelpCases, configCases, err = initTestCases()
	if err != nil {
		t.Fatalf("failed to generate test cases, error: %s", err)
	}
	client := appstoreconnect.InitTestClient(t)

	tests := []struct {
		name         string
		entitlement  serialized.Object
		bundleID     string
		capabilities []appstoreconnect.BundleIDCapability
		wantErr      bool
	}{
		{
			entitlement: serialized.Object(map[string]interface{}{
				"aps-environment": "development",
				"com.apple.developer.default-data-protection": "NSFileProtectionComplete",
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.com.bitrise.Xcode-10-default",
				},
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit", "CloudDocuments",
				},
				"com.apple.developer.siri":                           true,
				"com.apple.developer.ubiquity-container-identifiers": []interface{}{"iCloud.com.bitrise.Xcode-10-default"},
				"com.apple.developer.ubiquity-kvstore-identifier":    "com.bitrise.Xcode-10-default",
			}),
			bundleID: "com.bitrise.Xcode-10-default",
			capabilities: func() []appstoreconnect.BundleIDCapability {
				bundleID, err := fetchBundleID(client, "com.bitrise.Xcode-10-default")
				if err != nil {
					t.Fatalf("failed to fetch bundleID from Dev Portal for %s", "com.bitrise.Xcode-10-default")
				}
				return bundleID.Capabilities
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := syncAppServices(client, tt.entitlement, tt.bundleID, tt.capabilities); (err != nil) != tt.wantErr {
				t.Errorf("syncAppServices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_updateAppService(t *testing.T) {
	tests := []struct {
		name               string
		client             *appstoreconnect.Client
		capabilityID       string
		capabilityType     appstoreconnect.CapabilityType
		capabilitySettings []appstoreconnect.CapabilitySetting
		wantErr            bool
	}{
		{
			name:           "Data protection - complete for 25Z8895ZJC (com.bitrise.Xcode-10-default)",
			client:         appstoreconnect.InitTestClient(t),
			capabilityID:   "25Z8895ZJC_DATA_PROTECTION",
			capabilityType: appstoreconnect.DataProtection,
			capabilitySettings: []appstoreconnect.CapabilitySetting{
				appstoreconnect.CapabilitySetting{
					Options: []appstoreconnect.CapabilityOption{
						appstoreconnect.CapabilityOption{
							Key: appstoreconnect.CompleteProtection,
						},
					},
					Key: appstoreconnect.DataProtectionPermissionLevel,
				},
			},
			wantErr: false,
		},

		{
			name:           "Data protection - Unless_open for 25Z8895ZJC (com.bitrise.Xcode-10-default)",
			client:         appstoreconnect.InitTestClient(t),
			capabilityID:   "25Z8895ZJC_DATA_PROTECTION",
			capabilityType: appstoreconnect.DataProtection,
			capabilitySettings: []appstoreconnect.CapabilitySetting{
				appstoreconnect.CapabilitySetting{
					Options: []appstoreconnect.CapabilityOption{
						appstoreconnect.CapabilityOption{
							Key: appstoreconnect.ProtectedUnlessOpen,
						},
					},
					Key: appstoreconnect.DataProtectionPermissionLevel,
				},
			},
			wantErr: false,
		},

		{
			name:           "Data protection - until_first_auth for 25Z8895ZJC (com.bitrise.Xcode-10-default)",
			client:         appstoreconnect.InitTestClient(t),
			capabilityID:   "25Z8895ZJC_DATA_PROTECTION",
			capabilityType: appstoreconnect.DataProtection,
			capabilitySettings: []appstoreconnect.CapabilitySetting{
				appstoreconnect.CapabilitySetting{
					Options: []appstoreconnect.CapabilityOption{
						appstoreconnect.CapabilityOption{
							Key: appstoreconnect.ProtectedUntilFirstUserAuth,
						},
					},
					Key: appstoreconnect.DataProtectionPermissionLevel,
				},
			},
			wantErr: false,
		},
		{
			name:           "iCloud - xcode_5 for 25Z8895ZJC (com.bitrise.Xcode-10-default)",
			client:         appstoreconnect.InitTestClient(t),
			capabilityID:   "25Z8895ZJC_ICLOUD",
			capabilityType: appstoreconnect.ICloud,
			capabilitySettings: []appstoreconnect.CapabilitySetting{
				appstoreconnect.CapabilitySetting{
					Options: []appstoreconnect.CapabilityOption{
						appstoreconnect.CapabilityOption{
							Key: appstoreconnect.Xcode5,
						},
					},
					Key: appstoreconnect.IcloudVersion,
				},
			},
			wantErr: false,
		},
		{
			name:           "iCloud - xcode_6 for 25Z8895ZJC (com.bitrise.Xcode-10-default)",
			client:         appstoreconnect.InitTestClient(t),
			capabilityID:   "25Z8895ZJC_ICLOUD",
			capabilityType: appstoreconnect.ICloud,
			capabilitySettings: []appstoreconnect.CapabilitySetting{
				appstoreconnect.CapabilitySetting{
					Options: []appstoreconnect.CapabilityOption{
						appstoreconnect.CapabilityOption{
							Key: appstoreconnect.Xcode6,
						},
					},
					Key: appstoreconnect.IcloudVersion,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateAppService(tt.client, tt.capabilityID, tt.capabilityType, tt.capabilitySettings); (err != nil) != tt.wantErr {
				t.Errorf("updateAppService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
