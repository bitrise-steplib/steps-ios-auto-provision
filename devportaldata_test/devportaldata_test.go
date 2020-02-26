package devportaldata_test

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/devportaldata"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyWithHeader(t *testing.T) {
	// Arrange
	expectedResult := "-----BEGIN PRIVATE KEY-----\nprivate key without header\n-----END PRIVATE KEY-----"
	tests := []struct {
		name       string
		privateKey string
		want       string
	}{
		{
			name:       "adds header",
			privateKey: "private key without header",
			want:       expectedResult,
		},
		{
			name:       "skips adding header",
			privateKey: expectedResult,
			want:       expectedResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSubject := devportaldata.DevPortalData{
				IssuerID:    "",
				KeyID:       "",
				PrivateKey:  "private key without header",
				TestDevices: []devportaldata.DeviceData{},
			}

			// Act
			result := testSubject.PrivateKeyWithHeader()

			// Assert
			assert.Equal(t, expectedResult, result, "private key should be equal to the expected one containing header and footer")
		})
	}
}
