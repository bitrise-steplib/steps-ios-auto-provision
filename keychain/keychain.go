package keychain

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/hashicorp/go-version"
)

type Keychain struct {
	Path     string
	Password string
}


func ListKeychains() ([]string, error) {
	outbuf := bytes.NewBuffer([]byte{})
	cmd := command.New("security", "list-keychain").SetStdout(outbuf).SetStderr(outbuf)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		log.Errorf(outbuf.String())
		return nil, fmt.Errorf("run command: %s", err)
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
func CreateKeychain(path, password string, out, errout io.Writer) (*Keychain, error) {
	cmd := command.New("security", "-v", "create-keychain", "-p", password, path).SetStdout(out).SetStderr(errout)

	// log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run command: %s", err)
	}

	return &Keychain{
		Path:     path,
		Password: password,
	}, nil
}

// ImportCertificate adds the certificate at path, protected by
// passphrase to the kc keychain.
func (kc Keychain) ImportCertificate(path, passphrase string) error {
	cmd := command.New("security", "import", path, "-k", kc.Path, "-P", passphrase, "-A").SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run command: %s", err)
	}

	return nil
}

// SetKeyPartitionList sets the partition list
// for the keychain to allow access for tools.
func (kc Keychain) SetKeyPartitionList() error {
	cmd := command.New("security", "set-key-partition-list", "-S", "apple-tool:,apple:", "-k", kc.Password, kc.Path).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run command: %s", err)
	}

	return nil
}

// SetLockSettings sets keychain autolocking.
func (kc Keychain) SetLockSettings() error {
	cmd := command.New("security", "-v", "set-keychain-settings", "-lut", "72000", kc.Path).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run command: %s", err)
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
		return fmt.Errorf("run command: %s", err)
	}

	return nil
}

// SetAsDefault sets the keychain as the
// default keychain for the system.
func (kc Keychain) SetAsDefault() error {
	cmd := command.New("security", "-v", "default-keychain", "-s", kc.Path).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run command: %s", err)
	}

	return nil
}

// Unlock unlocks the keychain
func (kc Keychain) Unlock() error {
	cmd := command.New("security", "-v", "unlock-keychain", "-p", kc.Password, kc.Path).SetStdout(os.Stdout).SetStderr(os.Stderr)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run command: %s", err)
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
