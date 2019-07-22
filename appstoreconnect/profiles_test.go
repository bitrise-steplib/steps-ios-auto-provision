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

func TestProfile_isXcodeManaged(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		want    bool
	}{
		{
			name: "Non Xcode Managed - Bitrise Develpment (io.bitrise.sample)",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Bitrise Develpment (io.bitrise.sample)",
				},
			},
			want: false,
		},
		{
			name: "Non Xcode Managed - Bitrise App Store (io.bitrise.sample)",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Bitrise App Store (io.bitrise.sample)",
				},
			},
			want: false,
		},
		{
			name: "Non Xcode Managed - Bitrise Ad Hoc (io.bitrise.sample)",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Bitrise Ad Hoc (io.bitrise.sample)",
				},
			},
			want: false,
		},
		{
			name: "Xcode Managed - XC Ad Hoc: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC Ad Hoc: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - XC: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - XC Ad Hoc: { bundle id }",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC Ad Hoc: { bundle id }",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - XC: { bundle id }",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC: { bundle id }",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Ad Hoc Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Ad Hoc Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Ad Hoc Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Ad Hoc Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Ad Hoc Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Ad Hoc Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Ad Hoc Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Ad Hoc Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Ad Hoc Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Ad Hoc Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Ad Hoc Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Ad Hoc Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.profile

			if got := p.isXcodeManaged(); got != tt.want {
				t.Errorf("Profile.isXcodeManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvisioningService_ListProfiles(t *testing.T) {
	tests := []struct {
		name    string
		opt     *ListProfilesOptions
		wantErr bool
	}{
		{
			name:    "List profiles",
			opt:     nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ProvisioningService{
				client: initTestClient(t),
			}
			got, err := s.ListProfiles(tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisioningService.ListProfiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ProvisioningService.ListProfiles() = is NIL,")
			}
			if got.Data == nil {
				t.Errorf("ProvisioningService.ListProfiles() Response.Data is NIL,")
			}
		})
	}
}

// func TestProvisioningService_CreateProfile(t *testing.T) {
// 	client := initTestClient(t)

// 	tests := []struct {
// 		name    string
// 		body    ProfileCreateRequest
// 		wantErr bool
// 	}{
// 		{
// 			name: "Create development profile for com.bitrise.Test-Xcode-Managed",
// 			body: NewProfileCreateRequest(
// 				IOSAppDevelopment,
// 				"Bitrise development (com.bitrise.Test-Xcode-Managed)",
// 				"7ZDPMNJW89", // com.bitrise.Test-Xcode-Managed
// 				[]Certificate{
// 					Certificate{
// 						ID:         "7JF32NQGYF",
// 						Type:       "Certificate",
// 						Attributes: CertificateAttributes{},
// 					},
// 				}, []Device{
// 					Device{
// 						Type:       "Device",
// 						ID:         "T6SV2G2HNM",
// 						Attributes: DeviceAttributes{},
// 					},
// 				}),
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := ProvisioningService{
// 				client: client,
// 			}
// 			_, err := s.CreateProfile(tt.body)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ProvisioningService.CreateProfile() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 		})
// 	}
// }
