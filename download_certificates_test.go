package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-tools/go-xcode/certificateutil"
)

// generateTestCertificate creates a mock certificate for test purposes
func generateTestCertificate(serial int64, teamID, teamName, commonName string, expiry time.Time) (*x509.Certificate, *rsa.PrivateKey, error) {
	CAtemplate := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"Pear Worldwide Developer Relations"},
			CommonName:   "Pear Worldwide Developer Relations CA",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// generate private key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Self-signed certificate, parent is the template
	CAcertData, err := x509.CreateCertificate(rand.Reader, CAtemplate, CAtemplate, &privatekey.PublicKey, privatekey)
	if err != nil {
		return nil, nil, err
	}
	CAcert, err := x509.ParseCertificate(CAcertData)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		IsCA:                  true,
		BasicConstraintsValid: true,
		SerialNumber:          big.NewInt(serial),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{teamName},
			OrganizationalUnit: []string{teamID},
			CommonName:         commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  expiry,
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		KeyUsage: x509.KeyUsageDigitalSignature,
	}

	// generate private key
	privatekey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	certData, err := x509.CreateCertificate(rand.Reader, template, CAcert, &privatekey.PublicKey, privatekey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, nil, err
	}

	return cert, privatekey, nil
}

func TestDownloadLocalCertificates(t *testing.T) {
	const teamID = "MYTEAMID"
	const commonName = "Apple Developer: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)
	serial := int64(1234)

	cert, privateKey, err := generateTestCertificate(serial, teamID, teamName, commonName, expiry)
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
			URLs: []p12URL{p12URL{
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
