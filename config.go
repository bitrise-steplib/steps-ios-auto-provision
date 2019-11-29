package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
)

// CertificateFileURL contains a p12 file URL and passphrase
type CertificateFileURL struct {
	URL, Passphrase string
}

// Config holds the step inputs
type Config struct {
	KeyID         stepconf.Secret `env:"key_id,required"`
	IssuerID      stepconf.Secret `env:"issuer_id,required"`
	PrivateKeyURL string          `env:"private_key,required"`

	ProjectPath   string `env:"project_path,dir"`
	Scheme        string `env:"scheme,required"`
	Configuration string `env:"configuration"`

	Distribution        string `env:"distribution_type,opt[development,app-store,ad-hoc,enterprise]"`
	Devices             string `env:"devices"`
	MinProfileDaysValid string `env:"min_profile_days_valid"`

	CertificateURLList        string          `env:"certificate_urls,required"`
	CertificatePassphraseList stepconf.Secret `env:"passphrases"`
	KeychainPath              string          `env:"keychain_path,required"`
	KeychainPassword          stepconf.Secret `env:"keychain_password,required"`

	VerboseLog bool `env:"verbose_log,opt[no,yes]"`
}

// PrivateKey ...
func (c Config) PrivateKey() ([]byte, error) {
	if strings.HasPrefix(c.PrivateKeyURL, "file://") {
		return fileutil.ReadBytesFromFile(strings.TrimPrefix(c.PrivateKeyURL, "file://"))
	}

	return downloadContent(c.PrivateKeyURL)
}

// DistributionType ...
func (c Config) DistributionType() autoprovision.DistributionType {
	return autoprovision.DistributionType(c.Distribution)
}

// DeviceIDs ...
func (c Config) DeviceIDs() []string {
	return split(c.Devices, ",", true)
}

// ValidateCertificates validates if the number of certificate URLs matches those of passphrases
func (c Config) ValidateCertificates() ([]string, []string, error) {
	pfxURLs := split(c.CertificateURLList, "|", true)
	passphrases := split(string(c.CertificatePassphraseList), "|", false)

	if len(pfxURLs) != len(passphrases) {
		return nil, nil, fmt.Errorf("certificates count (%d) and passphrases count (%d) should match", len(pfxURLs), len(passphrases))
	}

	return pfxURLs, passphrases, nil
}

// CertificateFileURLs returns an array of p12 file URLs and passphrases
func (c Config) CertificateFileURLs() ([]CertificateFileURL, error) {
	pfxURLs, passphrases, err := c.ValidateCertificates()
	if err != nil {
		return nil, err
	}

	files := make([]CertificateFileURL, len(pfxURLs))
	for i, pfxURL := range pfxURLs {
		files[i] = CertificateFileURL{
			URL:        pfxURL,
			Passphrase: passphrases[i],
		}
	}

	return files, nil
}

func split(list string, sep string, omitEmpty bool) (items []string) {
	for _, e := range strings.Split(list, sep) {
		if omitEmpty {
			e = strings.TrimSpace(e)
		}
		if !omitEmpty || len(e) > 0 {
			items = append(items, e)
		}
	}
	return
}

func downloadContent(url string) ([]byte, error) {
	var contentBytes []byte
	return contentBytes, retry.Times(2).Wait(time.Duration(3) * time.Second).Try(func(attempt uint) error {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download from %s: %s", url, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Warnf("failed to close (%s) body", url)
			}
		}()

		contentBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read received conent: %s", err)
		}
		return nil
	})
}
