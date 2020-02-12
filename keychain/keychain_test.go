package keychain

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommandFactory struct {
	mock.Mock
}

type MockCommand struct {
	mock.Mock
}

func (m *MockCommandFactory) New(name string, args ...string) CanRunAndReturnTrimmedCombinedOutput {
	mockargs := m.Called(name, args)
	return mockargs.Get(0).(*MockCommand)
}

func (m *MockCommand) RunAndReturnTrimmedCombinedOutput() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockCommand) PrintableCommandArgs() string {
	args := m.Called()
	return args.String(0)
}

func TestCreateKeychainSucceeds(t *testing.T) {
	// Arrange
	testPath := "testPath"
	testPass := "strongpass"
	testSecret := stepconf.Secret(testPass)
	testOutput := "testOutput"
	mockCommandFactory := new(MockCommandFactory)
	mockCommand := new(MockCommand)

	mockCommandFactory.On("New", "security", []string{"-v", "create-keychain", "-p", testPass, testPath}).Return(mockCommand)
	mockCommand.On("RunAndReturnTrimmedCombinedOutput").Return(testOutput, nil)

	// Act
	keychain, err := createKeychain(mockCommandFactory.New, testPath, testSecret)

	// Assert
	assert.Nil(t, err, "error creating keychain: %s", err)
	assert.NotNil(t, keychain, "created keychain should not be nil")
	assert.Equal(t, testPath, keychain.path)
	assert.Equal(t, testSecret, keychain.password)
	mockCommandFactory.AssertExpectations(t)
}

func TestCreateKeychainFails(t *testing.T) {
	// Arrange
	testPath := "testPath"
	testPass := "strongpass"
	testSecret := stepconf.Secret(testPass)
	testError := fmt.Errorf("testError")
	mockCommandFactory := new(MockCommandFactory)
	mockCommand := new(MockCommand)

	mockCommandFactory.On("New", "security", []string{"-v", "create-keychain", "-p", testPass, testPath}).Return(mockCommand)
	mockCommand.On("RunAndReturnTrimmedCombinedOutput").Return("", testError)

	// Act
	keychain, err := createKeychain(mockCommandFactory.New, testPath, testSecret)

	// Assert
	assert.NotNil(t, err, "call should fail")
	assert.Contains(t, err.Error(), testError.Error(), "unexpected error")
	assert.Nil(t, keychain, "created keychain should not be nil")
	mockCommandFactory.AssertExpectations(t)
}

func TestImportCertificateSucceeds(t *testing.T) {
	// Arrange
	testPath := "testPath"
	testPass := "strongpass"
	testImportPath := "testImportPath"
	testOutput := "testOutput"
	testSecret := stepconf.Secret(testPass)
	mockCommandFactory := new(MockCommandFactory)
	mockCommand := new(MockCommand)

	mockCommandFactory.On("New", "security", []string{"import", testImportPath, "-k", testPath, "-P", testPass, "-A"}).Return(mockCommand)
	mockCommand.On("RunAndReturnTrimmedCombinedOutput").Return(testOutput, nil)

	// Act
	k := Keychain{
		path:                    testPath,
		password:                testSecret,
		commandFactory:          mockCommandFactory.New,
		normalizedOSTempDirPath: func(s string) (string, error) { return s, nil },
		writeBytesToFile:        func(pth string, fileCont []byte) error { return nil },
	}
	err := k.importCertificate(testImportPath, testSecret)

	// Assert
	assert.Nil(t, err, "error creating keychain: %s", err)
}

func TestImportCertificateCommandFails(t *testing.T) {
	// Arrange
	testPath := "testPath"
	testPass := "strongpass"
	testImportPath := "testImportPath"
	testError := fmt.Errorf("testError")
	testSecret := stepconf.Secret(testPass)
	mockCommandFactory := new(MockCommandFactory)
	mockCommand := new(MockCommand)
	fakeNormalizedOSTempDirPath := func(s string) (string, error) { return s, nil }
	fakeWriteBytesToFile := func(pth string, fileCont []byte) error { return nil }

	mockCommandFactory.On("New", "security", []string{"import", testImportPath, "-k", testPath, "-P", testPass, "-A"}).Return(mockCommand)
	mockCommand.On("RunAndReturnTrimmedCombinedOutput").Return("", testError)

	// Act
	k := Keychain{
		path:                    testPath,
		password:                testSecret,
		commandFactory:          mockCommandFactory.New,
		normalizedOSTempDirPath: fakeNormalizedOSTempDirPath,
		writeBytesToFile:        fakeWriteBytesToFile,
	}
	err := k.importCertificate(testImportPath, testSecret)

	// Assert
	assert.NotNil(t, err, "call should fail")
}
