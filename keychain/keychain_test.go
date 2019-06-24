package keychain

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCreateKeychain(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-create-keychain")
	if err != nil {
		t.Errorf("setup: create temp dir for keychain: %s", err)
	}
	path := filepath.Join(dir, "testkeychain")
	_, err = createKeychain(path, "randompassword")

	if err != nil {
		t.Errorf("error creating keychain: %s", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("keychain not created")
	}
}

func TestImportCertificate(t *testing.T) {
	const (
		testCertPassword     = "challenge"
		testCertFilename     = "testcert.crt"
		testKeychainFilename = "testkeychain"
		testKeychainPassword = "password"
	)

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("setup: faliled to get working dir, error: %s", err)
	}
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

	kchain := Keychain{Path: pathTesting, Password: testKeychainPassword}

	if err := kchain.importCertificate(pathCert, testCertPassword); err != nil {
		t.Errorf("could not import cert: %s", err)
	}

}
