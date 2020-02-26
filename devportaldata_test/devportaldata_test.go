package devportaldata_test

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/devportaldata"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyWithHeader(t *testing.T) {
	// Arrange
	tests := []struct {
		name       string
		privateKey string
		want       string
	}{
		{
			name:       "adds header",
			privateKey: "private key without header",
			want:       "-----BEGIN PRIVATE KEY-----\nprivate key without header\n-----END PRIVATE KEY-----",
		},
		{
			name:       "skips adding header",
			privateKey: "-----BEGIN PRIVATE KEY-----\nprivate key with header\n-----END PRIVATE KEY-----",
			want:       "-----BEGIN PRIVATE KEY-----\nprivate key with header\n-----END PRIVATE KEY-----",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSubject := devportaldata.DevPortalData{
				IssuerID:    "",
				KeyID:       "",
				PrivateKey:  tt.privateKey,
				TestDevices: []devportaldata.DeviceData{},
			}

			// Act
			result := testSubject.PrivateKeyWithHeader()

			// Assert
			assert.Equal(t, tt.want, result, "private key should be equal to the expected one containing header and footer")
		})
	}
}
