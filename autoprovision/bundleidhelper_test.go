package autoprovision

import (
	"testing"

	"github.com/bitrise-io/xcode-project/xcodeproj"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// func TestEnsureApp(t *testing.T) {
// 	var err error
// 	schemeCases, targetCases, xcProjCases, projHelpCases, configCases, err = initTestCases()
// 	if err != nil {
// 		t.Fatalf("Failed to initialize test cases, error: %s", err)
// 	}

// 	tests := []struct {
// 		name          string
// 		client        *appstoreconnect.Client
// 		projectHelper ProjectHelper
// 		target        xcodeproj.Target
// 		config        string
// 		wantErr       bool
// 	}{
// 		{
// 			name:          "Ensure app for Xcode-10_default for config " + configCases[0],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[0],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[0].Targets {
// 					if t.Name == "Xcode-10_default" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[0],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for Xcode-10_defaultTests for config " + configCases[0],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[0],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[0].Targets {
// 					if t.Name == "Xcode-10_defaultTests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[0],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for Xcode-10_defaultUITests for config " + configCases[0],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[0],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[0].Targets {
// 					if t.Name == "Xcode-10_defaultUITests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[0],
// 			wantErr: false,
// 		},

// 		{
// 			name:          "Ensure app for Xcode-10_default for config " + configCases[1],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[0],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[0].Targets {
// 					if t.Name == "Xcode-10_default" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[1],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for Xcode-10_defaultTests for config " + configCases[1],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[0],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[0].Targets {
// 					if t.Name == "Xcode-10_defaultTests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[1],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for Xcode-10_defaultUITests for config " + configCases[1],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[0],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[0].Targets {
// 					if t.Name == "Xcode-10_defaultUITests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[1],
// 			wantErr: false,
// 		},

// 		{
// 			name:          "Ensure app for TV_OS for config " + configCases[0],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[4],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[4].Targets {
// 					if t.Name == "TV_OS" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[0],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for TV_OSTests for config " + configCases[0],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[4],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[4].Targets {
// 					if t.Name == "TV_OSTests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[0],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for TV_OSUITests for config " + configCases[0],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[4],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[4].Targets {
// 					if t.Name == "TV_OSUITests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[0],
// 			wantErr: false,
// 		},

// 		{
// 			name:          "Ensure app for TV_OS for config " + configCases[1],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[4],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[4].Targets {
// 					if t.Name == "TV_OS" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[1],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for TV_OSTests for config " + configCases[1],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[4],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[4].Targets {
// 					if t.Name == "TV_OSTests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[1],
// 			wantErr: false,
// 		},
// 		{
// 			name:          "Ensure app for TV_OSUITests for config " + configCases[1],
// 			client:        appstoreconnect.InitTestClient(t),
// 			projectHelper: projHelpCases[4],
// 			target: func() xcodeproj.Target {
// 				for _, t := range projHelpCases[4].Targets {
// 					if t.Name == "TV_OSUITests" {
// 						return t
// 					}
// 				}
// 				return xcodeproj.Target{}
// 			}(),
// 			config:  configCases[1],
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if _, err := EnsureApp(tt.client, tt.projectHelper, tt.target, tt.config); (err != nil) != tt.wantErr {
// 				t.Errorf("EnsureApp() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_fetchBundleID(t *testing.T) {
// 	client := appstoreconnect.InitTestClient(t)

// 	tests := []struct {
// 		name               string
// 		bundleIDIdentifier string
// 		wantErr            bool
// 	}{
// 		{
// 			name:               "Fetch bundleID com.bitrise.io.testing.firefox",
// 			bundleIDIdentifier: "com.bitrise.io.testing.firefox",
// 			wantErr:            false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := fetchBundleID(client, tt.bundleIDIdentifier)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fetchBundleID() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if got.Attributes.Identifier != tt.bundleIDIdentifier {
// 				t.Errorf("fetchBundleID() wrong bundleID = %v, want with identifier %v", got, tt.bundleIDIdentifier)
// 				return
// 			}
// 		})
// 	}
// }

// func Test_appIDNameFrom(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		bundleID string
// 		targetID string
// 		want     string
// 	}{
// 		{
// 			name:     "Generate AppID name for bundleID without (-,_)",
// 			bundleID: "com.bitrise.TV.OSUITests",
// 			targetID: "1B11981D2164AF70001D927B",
// 			want:     "Bitrise com bitrise TV OSUITests 1B11981D2164AF70001D927B",
// 		},
// 		{
// 			name:     "Generate AppID name for bundleID with (-,_)",
// 			bundleID: "auto_provision.ios-simple-objc",
// 			targetID: "bc7cd9d1cc241639c4457975fefd920f",
// 			want:     "Bitrise auto provision ios simple objc bc7cd9d1cc241639c4457975fefd920f",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := appIDNameFrom(tt.bundleID, tt.targetID); got != tt.want {
// 				t.Errorf("appIDNameFrom() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_syncAppServices(t *testing.T) {
	var err error
	schemeCases, targetCases, xcProjCases, projHelpCases, configCases, err = initTestCases()
	if err != nil {
		t.Fatalf("failed to generate test cases, error: %s", err)
	}
	client := appstoreconnect.InitTestClient(t)

	tests := []struct {
		name              string
		projectHelper     ProjectHelper
		target            xcodeproj.Target
		configurationName string
		bundleID          BundleID
		wantErr           bool
	}{
		{
			projectHelper:     projHelpCases[0],
			target:            projHelpCases[0].MainTarget,
			configurationName: "Debug",
			bundleID: func() BundleID {
				targetBundleID, err := projHelpCases[0].TargetBundleID(projHelpCases[0].MainTarget.Name, "Debug")
				if err != nil {
					t.Fatalf("failed to get target bundle ID for test")
				}
				bundleID, err := fetchBundleID(client, targetBundleID)
				if err != nil {
					t.Fatalf("failed to fetch bundleID from Dev Portal for %s", targetBundleID)
				}
				return *bundleID
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := syncAppServices(client, tt.projectHelper, tt.target, tt.configurationName, tt.bundleID); (err != nil) != tt.wantErr {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateAppService(tt.client, tt.capabilityID, tt.capabilityType, tt.capabilitySettings); (err != nil) != tt.wantErr {
				t.Errorf("updateAppService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
