package main

import (
	"fmt"
	"strings"

	"github.com/bitrise-tools/go-steputils/stepconf"
)

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

// Print prints the config
func (c Config) Print() {
	// TODO: update stepconf.Print to receive the output writer
	// and write test for this method
	stepconf.Print(c)
}

// ProfileType is an iOS app distribution type
type ProfileType int

// Development or ProfileType
const (
	Invalid ProfileType = iota
	Development
	AdHoc
	Enterprise
	AppStore
	Unsupported
)

// String returns a string representation of ProfileType
func (p ProfileType) String() string {
	switch p {
	case Development:
		return "Development"
	case AdHoc:
		return "Ad Hoc"
	case Enterprise:
		return "Enterprise"
	case AppStore:
		return "App Store"
	default:
		return "unsupported"
	}
}

func stringToDistribution(distribution string) ProfileType {
	switch distribution {
	case "development":
		return Development
	case "app-store":
		return AppStore
	case "ad-hoc":
		return AdHoc
	case "enterprise":
		return Enterprise
	default:
		return Unsupported
	}
}

// Distribution returns a distribution type
func (c Config) Distribution() ProfileType {
	return stringToDistribution(c.distributionType)
}

// CertificateURLs returns a list of certificate urls
func (c Config) CertificateURLs() []string {
	return splitByPipe(c.certificateURLList, true)
}

// CertificatePassphrases returns a list of passphrases
func (c Config) CertificatePassphrases() []string {
	return splitByPipe(c.certificatePassphraseList, false)
}

// ValidateCertificateAndPassphraseCount returns an error if the number of certificates does not equal to the number of passphrases
func (c Config) ValidateCertificateAndPassphraseCount() error {
	certCount, passCount := len(c.CertificateURLs()), len(c.CertificatePassphrases())
	if certCount != passCount {
		return fmt.Errorf("certificates count (%d) and passphrases count (%d) should match", certCount, passCount)
	}
	return nil
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
