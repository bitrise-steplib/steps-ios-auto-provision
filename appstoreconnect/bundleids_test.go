package appstoreconnect

import (
	"os"
	"testing"

	"github.com/bitrise-io/go-utils/fileutil"
)

// initTestClient creates an AppStore client with a JWT token to communicate with the App Store connect API
// Export the BITRISE_PRIVATE_KEY_PATH, BITRISE_JWT_KEY_ID, and the BITRISE_JWT_ISSUER envs
// BITRISE_PRIVATE_KEY_PATH contains the path of the file which includes the private key of the Apple API key from App Store Connect
// BITRISE_JWT_KEY_ID contains the key of the Apple API key from App Store Connect
// BITRISE_JWT_ISSUER contains the issuer of the Apple API key from App Store Connect")
func initTestClient(t *testing.T) *Client {
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
	c, err := NewClient(jwtKey, jwtIssuer, b)
	if err != nil {
		t.Fatalf("Failed to generate appstoreconnec Client for test")
	}
	return c
}

func TestProvisioningService_ListBundleIDs(t *testing.T) {
	client := initTestClient(t)

	tests := []struct {
		name    string
		opt     *ListBundleIDsOptions
		wantErr bool
	}{
		{
			name: "Get bundle ID if for com.bitrise.Test-Xcode-Managed",
			opt: &ListBundleIDsOptions{
				FilterIdentifier: "com.bitrise.Test-Xcode-Managed",
				Include:          "bundleIdCapabilities",
			},
			wantErr: false,
		},
		{
			name: "Get bundle ID if for com.bitrise.io.testing.firefox",
			opt: &ListBundleIDsOptions{
				FilterIdentifier: "com.bitrise.io.testing.firefox",
				Include:          "bundleIdCapabilities",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ProvisioningService{
				client: client,
			}
			got, err := s.ListBundleIDs(tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisioningService.ListBundleIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ProvisioningService.ListBundleIDs() = is NIL")
			}
		})
	}
}
