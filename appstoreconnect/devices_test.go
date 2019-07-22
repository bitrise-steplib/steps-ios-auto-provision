package appstoreconnect

import (
	"testing"
)

func TestProvisioningService_ListDevices(t *testing.T) {
	client := InitTestClient(t)
	tests := []struct {
		name    string
		opt     *ListDevicesOptions
		want    *DevicesResponse
		wantErr bool
	}{
		{
			name:    "List devices",
			opt:     nil,
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ProvisioningService{
				client: client,
			}
			got, err := s.ListDevices(tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisioningService.ListDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ProvisioningService.ListDevices() = response is NIL")
			}
			if got.Data == nil {
				t.Errorf("ProvisioningService.ListDevices() = device count is 0")
			}
		})
	}
}
