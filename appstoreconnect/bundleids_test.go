package appstoreconnect

import "testing"

func TestProvisioningService_ListBundleIDs(t *testing.T) {
	client := InitTestClient(t)

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
