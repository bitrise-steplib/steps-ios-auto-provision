package autoprovision

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

// CertificateFileURL contains a p12 file URL and passphrase
type CertificateFileURL struct {
	URL, Passphrase string
}

// Config holds the step inputs
type Config struct {
	ProjectPath   string `env:"project_path,dir"`
	Scheme        string `env:"scheme,required"`
	Configuration string `env:"configuration"`

	distributionType string `env:"distribution_type,required"`
	TeamID           string `env:"team_id"`

	GenerateProfiles    string `env:"generate_profiles,opt[no,yes]"`
	MinProfileDaysValid string `env:"min_profile_days_valid"`
	VerboseLog          string `env:"verbose_log,opt[no,yes]"`

	certificateURLList        string `env:"certificate_urls,required"`
	certificatePassphraseList string `env:"passphrases,required"`

	KeychainPath     string `env:"keychain_path,required"`
	KeychainPassword string `env:"keychain_password,required"`

	BuildURL      string `env:"build_url,required"`
	BuildAPIToken string `env:"build_api_token,required"`
}

// ParseConfig expands the step inputs from the current environment
func ParseConfig() (c Config, err error) {
	err = stepconf.Parse(&c)
	return
}

func stringToDistribution(distribution string) appstoreconnect.ProfileType {
	switch distribution {
	case "development":
		return appstoreconnect.IOSAppDevelopment
	case "app-store":
		return appstoreconnect.IOSAppStore
	case "ad-hoc":
		return appstoreconnect.IOSAppAdHoc
	case "enterprise":
		return appstoreconnect.IOSAppInHouse
	default:
		return appstoreconnect.InvalidProfileType
	}
}

// Distribution returns a distribution type
func (c Config) Distribution() appstoreconnect.ProfileType {
	return stringToDistribution(c.distributionType)
}

// ValidateCertifacates validates if the number of certificate URLs matches those of passphrases
func (c Config) ValidateCertifacates() ([]string, []string, error) {
	pfxURLs := splitByPipe(c.certificateURLList, true)
	passphrases := splitByPipe(c.certificatePassphraseList, false)

	if len(pfxURLs) != len(passphrases) {
		return nil, nil, fmt.Errorf("certificates count (%d) and passphrases count (%d) should match", len(pfxURLs), len(passphrases))
	}

	return pfxURLs, passphrases, nil
}

// CertificateFileURLs returns an array of p12 file URLs and passphrases
func (c Config) CertificateFileURLs() ([]CertificateFileURL, error) {
	pfxURLs, passphrases, err := c.ValidateCertifacates()
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

func splitByPipe(list string, omitEmpty bool) (items []string) {
	for _, e := range strings.Split(list, "|") {
		if omitEmpty {
			e = strings.TrimSpace(e)
		}
		if !omitEmpty || len(e) > 0 {
			items = append(items, e)
		}
	}
	return
}

// DownloadLocalCertificates downloads and parses a list of p12 files
func DownloadLocalCertificates(URLs []CertificateFileURL) ([]certificateutil.CertificateInfoModel, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	var certInfos []certificateutil.CertificateInfoModel

	for i, p12 := range URLs {
		log.Debugf("Downloading p12 file number %d from %s", i, p12.URL)

		p12CertInfos, err := downloadPKCS12(httpClient, p12.URL, p12.Passphrase)
		if err != nil {
			return nil, err
		}
		log.Debugf("Codesign identities included: %s", certsToString(p12CertInfos))

		certInfos = append(certInfos, p12CertInfos...)
	}

	log.Debugf("%d certificates downloaded", len(certInfos))
	return certInfos, nil
}

// downloadPKCS12 downloads a pkcs12 format file and parses certificates and matching private keys.
func downloadPKCS12(httpClient *http.Client, certificateURL, passphrase string) ([]certificateutil.CertificateInfoModel, error) {
	contents, err := downloadFile(httpClient, certificateURL)
	if err != nil {
		return nil, err
	} else if contents == nil {
		return nil, fmt.Errorf("certificate (%s) is empty", certificateURL)
	}

	infos, err := certificateutil.CertificatesFromPKCS12Content(contents, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate (%s), err: %s", certificateURL, err)
	}

	return infos, nil
}

func downloadFile(httpClient *http.Client, src string) ([]byte, error) {
	url, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url (%s), error: %s", src, err)
	}

	// Local file
	if url.Scheme == "file" {
		src := strings.Replace(src, url.Scheme+"://", "", -1)

		return ioutil.ReadFile(src)
	}

	// Remote file
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request, error: %s", err)
	}

	var contents []byte
	err = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
		log.Debugf("Downloading %s, attempt %d", src, attempt)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		req = req.WithContext(ctx)

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to download (%s), error: %s", src, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Warnf("failed to close (%s) body, error: %s", src, err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download (%s) failed with status code (%d)", src, resp.StatusCode)
		}

		contents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response (%s), error: %s", src, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return contents, nil
}
