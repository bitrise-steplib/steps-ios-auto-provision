package devportaldata_test

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/devportaldata"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyWithHeaderAddsHeader(t *testing.T) {
	// Arrange
	expectedResult := "-----BEGIN PRIVATE KEY-----\nprivate key without header\n-----END PRIVATE KEY-----"
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
}

func TestPrivateKeyWithHeaderSkipsAddingHeader(t *testing.T) {
	// Arrange
	expectedResult := "-----BEGIN PRIVATE KEY-----\nprivate key without header\n-----END PRIVATE KEY-----"
	testSubject := devportaldata.DevPortalData{
		IssuerID:    "",
		KeyID:       "",
		PrivateKey:  expectedResult,
		TestDevices: []devportaldata.DeviceData{},
	}

	// Act
	result := testSubject.PrivateKeyWithHeader()

	// Assert
	assert.Equal(t, expectedResult, result, "private key should be equal to the expected one containing header and footer")
}
