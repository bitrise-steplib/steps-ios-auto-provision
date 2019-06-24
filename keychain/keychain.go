package keychain

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/hashicorp/go-version"
)

// Keychain descritbes a macOS Keychain
type Keychain struct {
	Path     string
	Password stepconf.Secret
}

// New ...
func New(pth string, pass stepconf.Secret) (*Keychain, error) {
	if exist, err := pathutil.IsPathExists(pth); err != nil {
		return nil, err
	} else if exist {
		return &Keychain{
			Path:     pth,
			Password: stepconf.Secret(pass),
		}, nil
	}

	p := pth + "-db"
	if exist, err := pathutil.IsPathExists(p); err != nil {
		return nil, err
	} else if exist {
		return &Keychain{
			Path:     pth,
			Password: pass,
		}, nil
	}

	return createKeychain(pth, pass)
}

// InstallCertificate ...
func (k Keychain) InstallCertificate(cert certificateutil.CertificateInfoModel, pass stepconf.Secret) error {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("keychain")
	if err != nil {
		return err
	}
	pth := filepath.Join(tmpDir, "Certificate.p12")
	_, err = cert.EncodeToP12(string(pass))
	if err != nil {
		return err
	}

	if err := k.importCertificate(pth, pass); err != nil {
		return err
	}

	if needed, err := isKeyPartitionListNeeded(); err != nil {
		return err
	} else if needed {
		if err := k.setKeyPartitionList(); err != nil {
			return err
		}
	}

	if err := k.setLockSettings(); err != nil {
		return err
	}

	if err := k.addToSearchPath(); err != nil {
		return err
	}

	if err := k.setAsDefault(); err != nil {
		return err
	}

	return k.unlock()
}

// listKeychains returns the paths of available keychains
func listKeychains() ([]string, error) {
	outbuf := bytes.NewBuffer([]byte{})
	cmd := command.New("security", "list-keychain").SetStdout(outbuf).SetStderr(outbuf)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		log.Errorf(outbuf.String())
		return nil, fmt.Errorf("list keychain command failed: %s", err)
	}

	out := outbuf.String()
	keychains := []string{}
	for _, path := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(path)
		trimmed = strings.Trim(trimmed, `"`)
		keychains = append(keychains, trimmed)
	}

	return keychains, nil
}

// createKeychain creates a new keychain file at
// path, protected by password. Returns an error
// if the keychain could not be created, otherwise
// a Keychain object representing the created
// keychain is returned.
func createKeychain(path string, password stepconf.Secret) (*Keychain, error) {
	params := []string{"-v", "create-keychain", "-p", "*****", path}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[3] = string(password)

	cmd := command.New("security", params...)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return nil, fmt.Errorf("create-keychain failed: %s", out)
		}
		return nil, fmt.Errorf("create-keychain failed: %s", err)
	}

	return &Keychain{
		Path:     path,
		Password: password,
	}, nil
}

// importCertificate adds the certificate at path, protected by
// passphrase to the k keychain.
func (k Keychain) importCertificate(path string, passphrase stepconf.Secret) error {
	params := []string{"import", path, "-k", k.Path, "-P", "*****", "-A"}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[5] = string(passphrase)

	cmd := command.New("security", params...)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("import failed: %s", out)
		}
		return fmt.Errorf("import failed: %s", err)
	}

	return nil
}

// setKeyPartitionList sets the partition list
// for the keychain to allow access for tools.
func (k Keychain) setKeyPartitionList() error {
	params := []string{"set-key-partition-list", "-S", "apple-tool:,apple:", "-k", "*****", k.Path}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[4] = string(k.Password)

	cmd := command.New("security", params...)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("set-key-partition-list failed: %s", out)
		}
		return fmt.Errorf("set-key-partition-list failed: %s", err)
	}

	return nil
}

// setLockSettings sets keychain autolocking.
func (k Keychain) setLockSettings() error {
	cmd := command.New("security", "-v", "set-keychain-settings", "-lut", "72000", k.Path)
	log.Debugf("$ %s", cmd.PrintableCommandArgs())

	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("set-keychain-settings failed: %s", out)
		}
		return fmt.Errorf("set-keychain-settings failed: %s", err)
	}

	return nil
}

// addToSearchPath registers the keychain
// in the systemwide search path
func (k Keychain) addToSearchPath() error {
	keychains, err := listKeychains()
	if err != nil {
		return fmt.Errorf("get keychain list: %s", err)
	}

	cmd := command.New("security", "-v", "list-keychains", "-s", strings.Join(keychains, " "))
	log.Debugf("$ %s", cmd.PrintableCommandArgs())

	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("list-keychains failed: %s", out)
		}
		return fmt.Errorf("list-keychains failed: %s", err)
	}

	return nil
}

// setAsDefault sets the keychain as the
// default keychain for the system.
func (k Keychain) setAsDefault() error {
	cmd := command.New("security", "-v", "default-keychain", "-s", k.Path)
	log.Debugf("$ %s", cmd.PrintableCommandArgs())

	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("default-keychain failed: %s", out)
		}
		return fmt.Errorf("default-keychain failed: %s", err)
	}

	return nil
}

// unlock unlocks the keychain
func (k Keychain) unlock() error {
	params := []string{"-v", "unlock-keychain", "-p", "*****", k.Path}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[3] = string(k.Password)

	cmd := command.New("security", params...)
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("unlock-keychain failed: %s", out)
		}
		return fmt.Errorf("unlock-keychain failed: %s", err)
	}

	return nil
}

// isKeyPartitionListNeeded determines whether
// key partition lists are used by the system.
func isKeyPartitionListNeeded() (bool, error) {
	outbuf := bytes.NewBuffer([]byte{})
	cmd := command.New("sw_vers", "-productVersion")
	log.Debugf("$ %s", cmd.PrintableCommandArgs())

	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return false, fmt.Errorf("sw_vers failed: %s", out)
		}
		return false, fmt.Errorf("sw_vers failed: %s", err)
	}

	const versionSierra = "10.12.0"
	sierra, err := version.NewVersion(versionSierra)
	if err != nil {
		return false, fmt.Errorf("invalid version (%s): %s", versionSierra, err)
	}

	current, err := version.NewVersion(outbuf.String())
	if err != nil {
		return false, fmt.Errorf("invalid version (%s): %s", current, err)
	}
	if current.GreaterThanOrEqual(sierra) {
		return true, nil
	}

	return false, nil
}
