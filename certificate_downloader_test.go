package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHttpClient mocking http client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (mock *MockLogger) log(m string, args ...interface{}) {
	mock.Called(m, args)
}

type MockCertificateConverter struct {
	mock.Mock
}

func (m *MockCertificateConverter) CertificatesFromPKCS12Content(content []byte, password string) ([]certificateutil.CertificateInfoModel, error) {
	args := m.Called(content, password)
	return args.Get(0).([]certificateutil.CertificateInfoModel), args.Error(1)
}

type MockIOUtils struct {
	mock.Mock
}

func (m *MockIOUtils) ReadFile(filename string) ([]byte, error) {
	args := m.Called(filename)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockIOUtils) ReadAll(r io.Reader) ([]byte, error) {
	args := m.Called(r)
	return args.Get(0).([]byte), args.Error(1)
}

type MockCloser struct {
	mock.Mock
}

func (m *MockCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCloser) Read(p []byte) (n int, err error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func TestDownloadCertificates_HttpClientWorking_DownloadSucceeds(t *testing.T) {
	// Arrange
	testPass := "testpass"
	testURLList := []CertificateFileURL{
		CertificateFileURL{
			URL:        "http://test.test",
			Passphrase: testPass,
		},
	}
	mockCloser := new(MockCloser)
	testHTTPResponse := http.Response{
		Status:     "200",
		StatusCode: 200,
		Body:       mockCloser,
	}
	testCertInfo := certificateutil.CertificateInfoModel{}
	mockHTTPClient := new(MockHTTPClient)
	mockCertificateConverter := new(MockCertificateConverter)
	mockIOUtils := new(MockIOUtils)
	mockLogger := new(MockLogger)
	certificateDownloader := CertificateDownloaderImpl{
		httpClient:      mockHTTPClient,
		debugLogger:     mockLogger.log,
		warnLogger:      mockLogger.log,
		pkcs12Converter: mockCertificateConverter.CertificatesFromPKCS12Content,
		readFile:        mockIOUtils.ReadFile,
		readAll:         mockIOUtils.ReadAll,
	}

	mockHTTPClient.On("Do", mock.Anything).Return(&testHTTPResponse, nil)
	mockLogger.On("log", mock.AnythingOfType("string"), mock.Anything)
	mockCertificateConverter.On("CertificatesFromPKCS12Content", mock.Anything, testPass).Return([]certificateutil.CertificateInfoModel{testCertInfo}, nil)
	mockCloser.On("Close").Return(nil)
	mockIOUtils.On("ReadAll", mockCloser).Return([]byte{1, 2, 3}, nil)

	// Act
	modelList, error := certificateDownloader.DownloadCertificates(testURLList)

	// Assert
	assert.Nil(t, error, "Error should be nil!")
	assert.NotNil(t, modelList, "Model list should not be nil!")
	mockIOUtils.AssertNotCalled(t, "ReadFile", mock.AnythingOfType("string"))
}
