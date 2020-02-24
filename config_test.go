package main

import (
	"reflect"
	"testing"
)

func TestConfig_ValidateCertificates(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		want, want1 []string
		wantErr     string
	}{
		{
			name:    "",
			config:  Config{CertificateURLList: "url", CertificatePassphraseList: "pass"},
			want:    []string{"url"},
			want1:   []string{"pass"},
			wantErr: "",
		},
		{
			name:    "",
			config:  Config{CertificateURLList: "url1|url2", CertificatePassphraseList: "pass1|"},
			want:    []string{"url1", "url2"},
			want1:   []string{"pass1", ""},
			wantErr: "",
		},
		{
			name:    "",
			config:  Config{CertificateURLList: "url1|url2", CertificatePassphraseList: "pass1"},
			want:    nil,
			want1:   nil,
			wantErr: "certificates count (2) and passphrases count (1) should match",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.config.ValidateCertificates()
			if (len(tt.wantErr) > 0 && err == nil) || (len(tt.wantErr) > 0 && err.Error() != tt.wantErr) || (len(tt.wantErr) == 0 && err != nil) {
				t.Errorf("Config.ValidateCertificateAndPassphraseCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.ValidateCertificates() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Config.ValidateCertificates() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
