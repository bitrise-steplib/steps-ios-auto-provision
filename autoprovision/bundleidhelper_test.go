package autoprovision

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func TestEnsureApp(t *testing.T) {
	var err error
	schemeCases, targetCases, xcProjCases, projHelpCases, configCases, err = initTestCases()
	if err != nil {
		t.Fatalf("Failed to initialize test cases, error: %s", err)
	}

	tests := []struct {
		name          string
		client        *appstoreconnect.Client
		projectHelper ProjectHelper
		config        string
		wantErr       bool
	}{
		{
			name:          "Ensure app for " + schemeCases[0] + "with config: " + configCases[0],
			client:        appstoreconnect.InitTestClient(t),
			projectHelper: projHelpCases[0],
			config:        configCases[0],
			wantErr:       false,
		},
		{
			name:          "Ensure app for " + schemeCases[1] + "with config: " + configCases[1],
			client:        appstoreconnect.InitTestClient(t),
			projectHelper: projHelpCases[1],
			config:        configCases[1],
			wantErr:       false,
		},
		{
			name:          "Ensure app for " + schemeCases[2] + "with config: " + configCases[2],
			client:        appstoreconnect.InitTestClient(t),
			projectHelper: projHelpCases[2],
			config:        configCases[2],
			wantErr:       false,
		},
		{
			name:          "Ensure app for " + schemeCases[3] + "with config: " + configCases[3],
			client:        appstoreconnect.InitTestClient(t),
			projectHelper: projHelpCases[3],
			config:        configCases[3],
			wantErr:       false,
		},
		{
			name:          "Ensure app for " + schemeCases[4] + "with config: " + configCases[4],
			client:        appstoreconnect.InitTestClient(t),
			projectHelper: projHelpCases[4],
			config:        configCases[4],
			wantErr:       false,
		},
		{
			name:          "Ensure app for " + schemeCases[5] + "with config: " + configCases[5],
			client:        appstoreconnect.InitTestClient(t),
			projectHelper: projHelpCases[5],
			config:        configCases[5],
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := EnsureApp(tt.client, tt.projectHelper, tt.config); (err != nil) != tt.wantErr {
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

func Test_appIDNameFrom(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		targetID string
		want     string
	}{
		{
			name:     "Generate AppID name for bundleID without (-,_)",
			bundleID: "com.bitrise.TV.OSUITests",
			targetID: "1B11981D2164AF70001D927B",
			want:     "Bitrise com bitrise TV OSUITests 1B11981D2164AF70001D927B",
		},
		{
			name:     "Generate AppID name for bundleID with (-,_)",
			bundleID: "auto_provision.ios-simple-objc",
			targetID: "bc7cd9d1cc241639c4457975fefd920f",
			want:     "Bitrise auto provision ios simple objc bc7cd9d1cc241639c4457975fefd920f",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appIDNameFrom(tt.bundleID, tt.targetID); got != tt.want {
				t.Errorf("appIDNameFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
