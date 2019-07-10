package main

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-xcode/certificateutil"
)

func TestDownloadLocalCertificates(t *testing.T) {
	const teamID = "MYTEAMID"
	const commonName = "Apple Developer: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)
	serial := int64(1234)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(serial, teamID, teamName, commonName, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}

	certInfo := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. Serial: %s Team ID: %s Common name: %s", certInfo.Serial, certInfo.TeamID, certInfo.CommonName)

	passphrase := ""
	certData, err := certInfo.EncodeToP12(passphrase)
	if err != nil {
		t.Errorf("init: failed to encode certificate, error: %s", err)
	}

	p12File, err := ioutil.TempFile("", "*.p12")
	if err != nil {
		t.Errorf("init: failed to create temp test file, error: %s", err)
	}

	if _, err = p12File.Write(certData); err != nil {
		t.Errorf("init: failed to write test file, error: %s", err)
	}

	if err = p12File.Close(); err != nil {
		t.Errorf("init: failed to close file, error: %s", err)
	}

	p12path := "file://" + p12File.Name()

	tests := []struct {
		name    string
		URLs    []p12URL
		want    []certificateutil.CertificateInfoModel
		wantErr bool
	}{
		{
			name: "Certificate matches generated.",
			URLs: []p12URL{{
				URL:        p12path,
				Passphrase: passphrase,
			}},
			want: []certificateutil.CertificateInfoModel{
				certInfo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DownloadLocalCertificates(tt.URLs)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadLocalCertificates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DownloadLocalCertificates() = %v, want %v", got, tt.want)
			}
		})
	}
}
