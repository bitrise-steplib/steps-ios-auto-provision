package devportaldata_test

import (
	"fmt"
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/devportaldata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockIOUtils struct {
	mock.Mock
}

func (m *MockIOUtils) ReadBytesFromFile(pth string) ([]byte, error) {
	args := m.Called(pth)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIOUtils) DownloadContent(url string, token string) ([]byte, error) {
	args := m.Called(url, token)
	return args.Get(0).([]byte), args.Error(1)
}

const testJSON = `{
	"key_id":"4RUVJ4SC38",
	"issuer_id":"69a6de7b-7325-47e3-e053-5b8c7c11a4d1",
	"private_key":"-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----",
	"test_devices":[]
}`

func TestGetDevPortalDataGetsDataFromDisk(t *testing.T) {
	// Arrange
	testToken := "testToken"
	testURL := "file:///test"
	mockIOUtils := new(MockIOUtils)
	mockIOUtils.On("ReadBytesFromFile", mock.Anything).Return([]byte(testJSON), nil)

	testSubject := devportaldata.Downloader{
		BuildAPIToken:     testToken,
		BuildURL:          testURL,
		DownloadContent:   mockIOUtils.DownloadContent,
		ReadBytesFromFile: mockIOUtils.ReadBytesFromFile,
	}

	// Act
	result, err := testSubject.GetDevPortalData()

	// Assert
	assert.Nil(t, err, "error should be nil")
	assert.NotNil(t, result, "result should not be nil")
	assert.Equal(t, "4RUVJ4SC38", result.KeyID)
	mockIOUtils.AssertNotCalled(t, "DownloadContent", mock.AnythingOfType("string"), mock.AnythingOfType("string"))
	mockIOUtils.AssertCalled(t, "ReadBytesFromFile", mock.AnythingOfType("string"))
}

func TestGetDevPortalDataGetsDataFromNetwork(t *testing.T) {
	// Arrange
	testToken := "testToken"
	testURL := "https:///test"
	expectedFullURL := "https:///test/apple_developer_portal_data.json"
	mockIOUtils := new(MockIOUtils)
	mockIOUtils.On("DownloadContent", expectedFullURL, testToken).Return([]byte(testJSON), nil)

	testSubject := devportaldata.Downloader{
		BuildAPIToken:     testToken,
		BuildURL:          testURL,
		DownloadContent:   mockIOUtils.DownloadContent,
		ReadBytesFromFile: mockIOUtils.ReadBytesFromFile,
	}

	// Act
	result, err := testSubject.GetDevPortalData()

	// Assert
	assert.Nil(t, err, "error should be nil")
	assert.NotNil(t, result, "result should not be nil")
	assert.Equal(t, "4RUVJ4SC38", result.KeyID)
	mockIOUtils.AssertCalled(t, "DownloadContent", expectedFullURL, testToken)
	mockIOUtils.AssertNotCalled(t, "ReadBytesFromFile", mock.AnythingOfType("string"))
}

func TestGetDevPortalDataInvalidURL(t *testing.T) {
	// Arrange
	testToken := "testToken"
	testURL := "%%%invalid"
	mockIOUtils := new(MockIOUtils)
	mockIOUtils.On("DownloadContent", mock.Anything, testToken).Return([]byte(testJSON), nil)

	testSubject := devportaldata.Downloader{
		BuildAPIToken:     testToken,
		BuildURL:          testURL,
		DownloadContent:   mockIOUtils.DownloadContent,
		ReadBytesFromFile: mockIOUtils.ReadBytesFromFile,
	}

	// Act
	_, err := testSubject.GetDevPortalData()

	// Assert
	assert.NotNil(t, err, "error should not be nil")
}

func TestGetDevPortalDataNetworkFailure(t *testing.T) {
	// Arrange
	testToken := "testToken"
	testURL := "https:///test"
	testError := fmt.Errorf("test error")
	mockIOUtils := new(MockIOUtils)
	mockIOUtils.On("DownloadContent", mock.Anything, testToken).Return([]byte(nil), testError)

	testSubject := devportaldata.Downloader{
		BuildAPIToken:     testToken,
		BuildURL:          testURL,
		DownloadContent:   mockIOUtils.DownloadContent,
		ReadBytesFromFile: mockIOUtils.ReadBytesFromFile,
	}

	// Act
	_, err := testSubject.GetDevPortalData()

	// Assert
	assert.NotNil(t, err, "error should not be nil")
	assert.Equal(t, testError, err)
}
