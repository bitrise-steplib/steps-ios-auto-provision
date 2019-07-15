package keychain_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/keychain"
)

const (
	testCertPassword     = "challenge"
	testCertFilename     = "testcert.crt"
	testKeychainFilename = "testkeychain"
	testKeychainPassword = "password"
)

func TestCreateKeychain(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-create-keychain")
	if err != nil {
		t.Errorf("setup: create temp dir for keychain: %s", err)
	}
	path := filepath.Join(dir, "testkeychain")
	outbuf := bytes.NewBuffer([]byte{})
	errbuf := bytes.NewBuffer([]byte{})
	_, err = keychain.CreateKeychain(path, "randompassword", outbuf, errbuf)

	if err != nil {
		t.Log(outbuf.String(), errbuf.String())
		t.Errorf("error creating keychain: %s", err)
	}
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Log(outbuf.String())
		t.Log(errbuf.String())
		t.Errorf("keychain not created")
	}
}

func TestImportCertificate(t *testing.T) {
	cwd, err := os.Getwd()
	dirTest := filepath.Join(cwd, "..", "test")
	pathGolden := filepath.Join(dirTest, "testkeychain")
	dirTmp, err := ioutil.TempDir("", "test-import-certificate")
	if err != nil {
		t.Errorf("setup: create temp dir for keychain: %s", err)
	}
	
	pathTesting := filepath.Join(dirTmp, testKeychainFilename)

	cmd := exec.Command("cp", pathGolden, pathTesting)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Errorf("setup: copy golden file: %s", err)
	}
	pathCert := filepath.Join(dirTest, testCertFilename)

	kchain := keychain.Keychain{Path: pathTesting, Password: testKeychainPassword}

	if err := kchain.ImportCertificate(pathCert, testCertPassword); err != nil {
		t.Errorf("could not import cert: %s", err)
	}

}