package autoprovision

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func Test_fetchProfile(t *testing.T) {
	tests := []struct {
		name        string
		client      *appstoreconnect.Client
		profileType appstoreconnect.ProfileType
		bundleID    string
		wantErr     bool
	}{
		{
			name:        "Fetch development profile for bundleID - com.bitrise.code-sign-test",
			client:      initTestClient(t),
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "com.bitrise.code-sign-test",
			wantErr:     false,
		},
		{
			name:        "Fetch app store profile for bundleID - com.bitrise.code-sign-test",
			client:      initTestClient(t),
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "com.bitrise.code-sign-test",
			wantErr:     false,
		},
		{
			name:        "Fetch ad-hoc profile for bundleID - com.bitrise.code-sign-test",
			client:      initTestClient(t),
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "com.bitrise.code-sign-test",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchProfile(tt.client, tt.profileType, tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("No Bitrise profile found for %v with a %v profile type.\nMakes sure there is a profile named Bitrise {ProfileType} (%v) on App Store Connect and it's valid.", tt.bundleID, tt.profileType, tt.bundleID)
			}
		})
	}
}

// initTestClient creates an AppStore client with a JWT token to communicate with the App Store connect API
// Export the BITRISE_PRIVATE_KEY_PATH, BITRISE_JWT_KEY_ID, and the BITRISE_JWT_ISSUER envs
// BITRISE_PRIVATE_KEY_PATH contains the path of the file which includes the private key of the Apple API key from App Store Connect
// BITRISE_JWT_KEY_ID contains the key of the Apple API key from App Store Connect
// BITRISE_JWT_ISSUER contains the issuer of the Apple API key from App Store Connect")
func initTestClient(t *testing.T) *appstoreconnect.Client {
	privateKeyPath := os.Getenv("BITRISE_PRIVATE_KEY_PATH")
	jwtKey := os.Getenv("BITRISE_JWT_KEY_ID")
	jwtIssuer := os.Getenv("BITRISE_JWT_ISSUER")
	if privateKeyPath == "" {
		t.Fatalf("Failed to init test client. BITRISE_PRIVATE_KEY_PATH env is missing. Export the path of the file which includes the private key of the Apple API key from App Store Connect")
	}
	if jwtKey == "" {
		t.Fatalf("Failed to init test client. BITRISE_JWT_KEY_ID env is missing. Export the key of the Apple API key from App Store Connect")
	}
	if jwtIssuer == "" {
		t.Fatalf("Failed to init test client. BITRISE_JWT_ISSUER env is missing. Export the issuer of the Apple API key from App Store Connect")
	}

	b, err := fileutil.ReadBytesFromFile(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to read bytes from $BITRISE_PRIVATE_KEY_PATH, error: %s", err)
	}
	c, err := appstoreconnect.NewClient(jwtKey, jwtIssuer, b)
	if err != nil {
		t.Fatalf("Failed to generate appstoreconnec Client for test")
	}
	return c
}

func Test_profileName(t *testing.T) {
	tests := []struct {
		name        string
		profileType appstoreconnect.ProfileType
		bundleID    string
		want        string
		wantErr     bool
	}{
		{
			name:        "Test Bitrise iOS development profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "Bitrise development - (com.bitrise.code-sign-test bundleID)",
			wantErr:     false,
		},
		{
			name:        "Test Bitrise iOS app store profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "Bitrise app-store - (com.bitrise.code-sign-test bundleID)",
			wantErr:     false,
		},
		{
			name:        "Test Bitrise iOS ad-hoc profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "Bitrise ad-hoc - (com.bitrise.code-sign-test bundleID)",
			wantErr:     false,
		},
		{
			name:        "Test Bitrise TVOS development profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.TvOSAppDevelopment,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "Bitrise development - (com.bitrise.code-sign-test bundleID)",
			wantErr:     false,
		},
		{
			name:        "Test Bitrise TVOS app store profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.TvOSAppStore,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "Bitrise app-store - (com.bitrise.code-sign-test bundleID)",
			wantErr:     false,
		},
		{
			name:        "Test Bitrise TVOS ad-hoc profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.TvOSAppAdHoc,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "Bitrise ad-hoc - (com.bitrise.code-sign-test bundleID)",
			wantErr:     false,
		},
		{
			name:        "Test Bitrise Mac development profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.MacAppDevelopment,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Test Bitrise Mac app store profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.MacAppStore,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Test Bitrise Mac developer ID store profile generation for com.bitrise.code-sign-test bundleID",
			profileType: appstoreconnect.MacAppDirect,
			bundleID:    "com.bitrise.code-sign-test bundleID",
			want:        "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestEnsureManualProfile(t *testing.T) {
	client := initTestClient(t)

	tests := []struct {
		name             string
		profileType      appstoreconnect.ProfileType
		bundleID         string
		devices          []appstoreconnect.Device
		certifcates      appstoreconnect.Certificate
		isXcodeManaged   bool
		generateProfiles bool
		wantErr          bool
	}{
		{
			name:        "iOS Development signing, profile generation enabled - com.bitrise.code-sign-test",
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "com.bitrise.code-sign-test",
			devices: []appstoreconnect.Device{
				appstoreconnect.Device{
					Type:       "Device",
					ID:         "T6SV2G2HNM",
					Attributes: appstoreconnect.DeviceAttributes{},
				},
			},
			certifcates: appstoreconnect.Certificate{
				ID:         "7JF32NQGYF",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          false,
		},
		{
			name:        "iOS App Store signing, profile generation enabled - com.bitrise.code-sign-test",
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "com.bitrise.code-sign-test",
			devices:     nil,
			certifcates: appstoreconnect.Certificate{
				ID:         "62NM5AVBZ9",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          false,
		},
		{
			name:        "iOS Enterprise signing, profile generation enabled - com.bitrise.code-sign-test",
			profileType: appstoreconnect.IOSAppInHouse,
			bundleID:    "com.bitrise.code-sign-test",
			devices:     nil,
			certifcates: appstoreconnect.Certificate{
				ID:         "62NM5AVBZ9",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          true, // Enterprise subscription needed. Depends on the new Apple Developer Team.
		},
		{
			name:        "iOS Ad Hoc signing, profile generation enabled - com.bitrise.code-sign-test",
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "com.bitrise.code-sign-test",
			devices: []appstoreconnect.Device{
				appstoreconnect.Device{
					Type:       "Device",
					ID:         "T6SV2G2HNM",
					Attributes: appstoreconnect.DeviceAttributes{},
				},
			},
			certifcates: appstoreconnect.Certificate{
				ID:         "62NM5AVBZ9",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          false,
		},
		{
			name:        "iOS Development signing, profile generation enabled - com.bitrise.Test-Xcode-Managed",
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "com.bitrise.Test-Xcode-Managed",
			devices: []appstoreconnect.Device{
				appstoreconnect.Device{
					Type:       "Device",
					ID:         "T6SV2G2HNM",
					Attributes: appstoreconnect.DeviceAttributes{},
				},
			},
			certifcates: appstoreconnect.Certificate{
				ID:         "7JF32NQGYF",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          false,
		},
		{
			name:        "iOS App Store signing, profile generation enabled - com.bitrise.Test-Xcode-Managed",
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "com.bitrise.Test-Xcode-Managed",
			devices:     nil,
			certifcates: appstoreconnect.Certificate{
				ID:         "62NM5AVBZ9",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          false,
		},
		{
			name:        "iOS Enterprise signing, profile generation enabled - com.bitrise.Test-Xcode-Managed",
			profileType: appstoreconnect.IOSAppInHouse,
			bundleID:    "com.bitrise.Test-Xcode-Managed",
			devices:     nil,
			certifcates: appstoreconnect.Certificate{
				ID:         "62NM5AVBZ9",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          true, // Enterprise subscription needed. Depends on the new Apple Developer Team.
		},
		{
			name:        "iOS Ad Hoc signing, profile generation enabled - com.bitrise.Test-Xcode-Managed",
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "com.bitrise.Test-Xcode-Managed",
			devices: []appstoreconnect.Device{
				appstoreconnect.Device{
					Type:       "Device",
					ID:         "T6SV2G2HNM",
					Attributes: appstoreconnect.DeviceAttributes{},
				},
			},
			certifcates: appstoreconnect.Certificate{

				ID:         "62NM5AVBZ9",
				Type:       "Certificate",
				Attributes: appstoreconnect.CertificateAttributes{},
			},
			isXcodeManaged:   false,
			generateProfiles: true,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ensureManualProfile(client, tt.profileType, tt.bundleID, tt.certifcates, tt.devices)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureProfiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
