package keychain

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/hashicorp/go-version"
)

// Keychain descritbes a macOS Keychain
type Keychain struct {
	Path     string
	Password stepconf.Secret
}

// ListKeychains returns the paths of available keychains
func ListKeychains() ([]string, error) {
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

// CreateKeychain creates a new keychain file at
// path, protected by password. Returns an error
// if the keychain could not be created, otherwise
// a Keychain object representing the created
// keychain is returned.
func CreateKeychain(path string, password stepconf.Secret, out, errout io.Writer) (*Keychain, error) {
	params := []string{"-v", "create-keychain", "-p", "*****", path}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[3] = string(password)

	cmd := command.New("security", params...).SetStdout(out).SetStderr(errout)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("create keychain command failed: %s", err)
	}

	return &Keychain{
		Path:     path,
		Password: password,
	}, nil
}

// ImportCertificate adds the certificate at path, protected by
// passphrase to the kc keychain.
func (kc Keychain) ImportCertificate(path string, passphrase stepconf.Secret) error {
	params := []string{"import", path, "-k", kc.Path, "-P", "*****", "-A"}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[5] = string(passphrase)

	cmd := command.New("security", params...).SetStdout(os.Stdout).SetStderr(os.Stderr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("import certificate command: %s", err)
	}

	return nil
}

// SetKeyPartitionList sets the partition list
// for the keychain to allow access for tools.
func (kc Keychain) SetKeyPartitionList() error {
	params := []string{"set-key-partition-list", "-S", "apple-tool:,apple:", "-k", "*****", kc.Path}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[4] = string(kc.Password)

	cmd := command.New("security", params...).SetStdout(os.Stdout).SetStderr(os.Stderr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set partition list command failed: %s", err)
	}

	return nil
}

// SetLockSettings sets keychain autolocking.
func (kc Keychain) SetLockSettings() error {
	cmd := command.New("security", "-v", "set-keychain-settings", "-lut", "72000", kc.Path).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set keychain lock settings command: %s", err)
	}

	return nil
}

// AddToSearchPath registers the keychain
// in the systemwide search path
func (kc Keychain) AddToSearchPath() error {
	keychains, err := ListKeychains()
	if err != nil {
		return fmt.Errorf("get keychain list: %s", err)
	}

	cmd := command.New("security", "-v", "list-keychains", "-s", strings.Join(keychains, " ")).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("add keychain to search path failed: %s", err)
	}

	return nil
}

// SetAsDefault sets the keychain as the
// default keychain for the system.
func (kc Keychain) SetAsDefault() error {
	cmd := command.New("security", "-v", "default-keychain", "-s", kc.Path).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("set keychain as default command failed: %s", err)
	}

	return nil
}

// Unlock unlocks the keychain
func (kc Keychain) Unlock() error {
	params := []string{"-v", "unlock-keychain", "-p", "*****", kc.Path}
	log.Debugf("$ %s", command.New("security", params...).PrintableCommandArgs())
	params[3] = string(kc.Password)

	cmd := command.New("security", params...).SetStdout(os.Stdout).SetStderr(os.Stderr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unlock keychain command failed: %s", err)
	}

	return nil
}

// IsKeyPartitionListNeeded determines whether
// key partition lists are used by the system.
func IsKeyPartitionListNeeded() (bool, error) {
	outbuf := bytes.NewBuffer([]byte{})
	cmd := command.New("sw_vers", "-productVersion").SetStdout(outbuf).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("get OS version: %s", err)
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
