package keychain

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-steputils/stepconf"
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

func TestKeychain_importCertificate(t *testing.T) {
	const (
		// #nosec: G101  Potential hardcoded credentials (gosec)
		testCertPassword     = "xGG}!Tk3/L'f-w){9pAD(tHKusK}?om$"
		testCertFilename     = "TestCert.p12"
		testKeychainFilename = "testkeychain"
		testKeychainPassword = "password"
	)

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("setup: faliled to get working dir: %s", err)
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
	if err := kchain.unlock(); err != nil {
		t.Errorf("failed to unlock keychain: %s", err)
	}

	type fields struct {
		Path     string
		Password stepconf.Secret
	}
	type args struct {
		path       string
		passphrase stepconf.Secret
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Good password",
			fields: fields{
				Path:     pathTesting,
				Password: testKeychainPassword,
			},
			args: args{
				path:       pathCert,
				passphrase: testCertPassword,
			},
			wantErr: false,
		},
		{
			name: "Incorrect password",
			fields: fields{
				Path:     pathTesting,
				Password: testKeychainPassword,
			},
			args: args{
				path:       pathCert,
				passphrase: "Incorrect password",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := Keychain{
				Path:     tt.fields.Path,
				Password: tt.fields.Password,
			}
			err := k.importCertificate(tt.args.path, tt.args.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keychain.importCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Logf("Keychain.importCertificate() error = %v", err)
			}
		})
	}
}
