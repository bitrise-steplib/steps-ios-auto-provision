package main

import (
	"reflect"
	"testing"

	"github.com/bitrise-tools/go-xcode/certificateutil"
)

// func TestToP12(t *testing.T) {
// 	s, err := New()
// 	require.NoError(t, err)
// 	certs, err := s.Download("file:///Users/godrei/Downloads/NewBitfallDevDistrCertificates.p12", "")
// 	require.NoError(t, err)
// 	for _, c := range certs {
// 		pth, err := s.ToP12(c)
// 		fmt.Println(pth)
// 		require.NoError(t, err)
// 	}
// 	require.Error(t, err)
// }

func TestDownload(t *testing.T) {
	tests := []struct {
		name    string
		URLs    []P12URL
		want    []certificateutil.CertificateInfoModel
		wantErr bool
	}{
		{
			name: "",
			URLs: []P12URL{P12URL{
				URL: "file:///Users/lpusok/Desktop/test_multi_cert.p12",
			}},
			want:    nil,
			wantErr: false,
		},
		{
			name: "",
			URLs: []P12URL{P12URL{
				URL:        "file:///Users/lpusok/Desktop/test_multi_cert_3.p12",
				Passphrase: "test",
			}},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DownloadCertificates(tt.URLs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Download() = %v, want %v", got, tt.want)
			}
		})
	}
}
