package appstoreconnect

import (
	"testing"
)

func TestProvisioningService_FetchCertificate(t *testing.T) {
	client := InitTestClient(t)

	tests := []struct {
		name         string
		serialNumber string
		wantErr      bool
	}{
		{
			name:         "Fetch certificae 5E77CAFD383D665A",
			serialNumber: "5E77CAFD383D665A",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ProvisioningService{
				client: client,
			}
			got, err := s.FetchCertificate(tt.serialNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisioningService.FetchCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == (Certificate{}) {
				t.Errorf("ProvisioningService.FetchCertificate() = is NIL")
			}
		})
	}
}

func TestProvisioningService_ListCertificates(t *testing.T) {
	client := InitTestClient(t)

	tests := []struct {
		name    string
		opt     *ListCertificatesOptions
		wantErr bool
	}{
		{
			name:    "List of certificates",
			opt:     nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ProvisioningService{
				client: client,
			}
			got, err := s.ListCertificates(tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisioningService.ListCertificates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ProvisioningService.ListCertificates() = NIL")
			}
		})
	}
}
