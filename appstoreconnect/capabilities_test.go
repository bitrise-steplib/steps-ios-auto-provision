package appstoreconnect

import "testing"

func TestProvisioningService_CapabilitiesOf(t *testing.T) {
	client := InitTestClient(t)
	s := ProvisioningService{
		client: client,
	}

	bundleIDResponse, err := s.ListBundleIDs(&ListBundleIDsOptions{
		FilterIdentifier: "com.bitrise.io.testing.firefox",
		Include:          "bundleIdCapabilities",
	})
	if err != nil {
		t.Fatalf("failed to fetch bundleID for testing bundleIDCapabilities")
	}

	tests := []struct {
		name     string
		bundleID BundleID
		wantErr  bool
	}{
		{
			name: "Fetch bundleID capabilitios of com.bitrise.io.testing.firefox",
			bundleID: BundleID{
				Relationships: BundleIDRelationships{
					Capabilities: struct {
						Links struct {
							Related string `json:"related"`
							Self    string `json:"next"`
						} `json:"links"`
					}{
						Links: struct {
							Related string `json:"related"`
							Self    string `json:"next"`
						}{
							Related: bundleIDResponse.Data[0].Relationships.Capabilities.Links.Related,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := s.CapabilitiesOf(tt.bundleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisioningService.CapabilitiesOf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
