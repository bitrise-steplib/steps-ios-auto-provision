package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
)

// CertificateFileURL contains a p12 file URL and passphrase
type CertificateFileURL struct {
	URL, Passphrase string
}

// DeviceData ...
type DeviceData struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	DeviceID   string `json:"device_identifier"`
	Title      string `json:"title"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	DeviceType string `json:"device_type"`
}

// DevPortalData ...
type DevPortalData struct {
	KeyID       string       `json:"key_id"`
	IssuerID    string       `json:"issuer_id"`
	PrivateKey  string       `json:"private_key"`
	TestDevices []DeviceData `json:"test_devices"`
}

// PrivateKeyWithHeader adds header and footer if needed
func (d DevPortalData) PrivateKeyWithHeader() string {
	if strings.HasPrefix(d.PrivateKey, "-----BEGIN PRIVATE KEY----") {
		return d.PrivateKey
	}

	return fmt.Sprint(
		"-----BEGIN PRIVATE KEY-----\n",
		d.PrivateKey,
		"\n-----END PRIVATE KEY-----",
	)
}

// Config holds the step inputs
type Config struct {
	BuildAPIToken string `env:"build_api_token,required"`
	BuildURL      string `env:"build_url,required"`

	ProjectPath   string `env:"project_path,dir"`
	Scheme        string `env:"scheme,required"`
	Configuration string `env:"configuration"`

	Distribution        string `env:"distribution_type,opt[development,app-store,ad-hoc,enterprise]"`
	MinProfileDaysValid string `env:"min_profile_days_valid"`

	CertificateURLList        string          `env:"certificate_urls,required"`
	CertificatePassphraseList stepconf.Secret `env:"passphrases"`
	KeychainPath              string          `env:"keychain_path,required"`
	KeychainPassword          stepconf.Secret `env:"keychain_password,required"`

	VerboseLog bool `env:"verbose_log,opt[no,yes]"`
}

// DevPortalData ...
func (c Config) DevPortalData() (devPortalData DevPortalData, err error) {
	var data []byte

	if strings.HasPrefix(c.BuildURL, "file://") {
		data, err = fileutil.ReadBytesFromFile(strings.TrimPrefix(c.BuildURL, "file://"))
	} else {
		var u *url.URL
		u, err = u.Parse(c.BuildURL)
		if err != nil {
			return
		}
		u.Path = path.Join(u.Path, "apple_developer_portal_data.json")
		data, err = downloadContent(u.String(), c.BuildAPIToken)
	}

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &devPortalData)
	return
}

// DistributionType ...
func (c Config) DistributionType() autoprovision.DistributionType {
	return autoprovision.DistributionType(c.Distribution)
}

// ValidateCertificates validates if the number of certificate URLs matches those of passphrases
func (c Config) ValidateCertificates() ([]string, []string, error) {
	pfxURLs := splitAndClean(c.CertificateURLList, "|", true)
	passphrases := splitAndClean(string(c.CertificatePassphraseList), "|", false)

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

// SplitAndClean ...
func splitAndClean(list string, sep string, omitEmpty bool) (items []string) {
	return sliceutil.CleanWhitespace(strings.Split(list, sep), omitEmpty)
}

func downloadContent(url string, buildAPIToken string) ([]byte, error) {
	var contentBytes []byte
	return contentBytes, retry.Times(2).Wait(time.Duration(3) * time.Second).Try(func(attempt uint) error {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for url %s: %s", url, err)
		}

		if buildAPIToken != "" {
			req.Header.Add("BUILD_API_TOKEN", buildAPIToken)
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to download from %s: %s", url, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Warnf("failed to close (%s) body", url)
			}
		}()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("request failed with status HTTP%d", resp.StatusCode)
		}

		contentBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read received conent: %s", err)
		}
		return nil
	})
}
