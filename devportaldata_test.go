package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyWithHeaderAddsHeader(t *testing.T) {
	// Arrange
	expectedResult := "-----BEGIN PRIVATE KEY-----\nprivate key without header\n-----END PRIVATE KEY-----"
	testSubject := DevPortalData{
		IssuerID:    "",
		KeyID:       "",
		PrivateKey:  "private key without header",
		TestDevices: []DeviceData{},
	}

	// Act
	result := testSubject.PrivateKeyWithHeader()

	// Assert
	assert.Equal(t, expectedResult, result, "private key should be equal to the expected one containing header and footer")
}

func TestPrivateKeyWithHeaderSkipsAddingHeader(t *testing.T) {
	// Arrange
	expectedResult := "-----BEGIN PRIVATE KEY-----\nprivate key without header\n-----END PRIVATE KEY-----"
	testSubject := DevPortalData{
		IssuerID:    "",
		KeyID:       "",
		PrivateKey:  expectedResult,
		TestDevices: []DeviceData{},
	}

	// Act
	result := testSubject.PrivateKeyWithHeader()

	// Assert
	assert.Equal(t, expectedResult, result, "private key should be equal to the expected one containing header and footer")
}
