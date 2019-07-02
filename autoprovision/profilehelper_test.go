package autoprovision

// How to run the tests
// Export the BITRISE_PRIVATE_KEY_PATH, BITRISE_JWT_KEY_ID, and the BITRISE_JWT_ISSUER envs and run go test.
// If you export this envs via .bitrise.secrets.yml, then run the gotests bitrise workflow.

// BITRISE_PRIVATE_KEY_PATH contains the path of the file which includes the private key of the Apple API key from App Store Connect
// BITRISE_JWT_KEY_ID contains the key of the Apple API key from App Store Connect
// BITRISE_JWT_ISSUER contains the issuer of the Apple API key from App Store Connect

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
		want        *Profile
		wantErr     bool
	}{
		{
			name:        "Fetch development profile for bundleID - ",
			client:      initTestClient(t),
			profileType: appstoreconnect.IOSAppDevelopment,
			bundleID:    "com.bitrise.code-sign-test",
			want:        &Profile{},
			wantErr:     false,
		},
		{
			name:        "Fetch development profile for bundleID - ",
			client:      initTestClient(t),
			profileType: appstoreconnect.IOSAppStore,
			bundleID:    "com.bitrise.code-sign-test",
			want:        &Profile{},
			wantErr:     false,
		},
		{
			name:        "Fetch development profile for bundleID - ",
			client:      initTestClient(t),
			profileType: appstoreconnect.IOSAppAdHoc,
			bundleID:    "com.bitrise.code-sign-test",
			want:        &Profile{},
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
