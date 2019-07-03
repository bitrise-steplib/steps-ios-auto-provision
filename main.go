package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/log"
)

// Distribution is an iOS app distribution type
type Distribution int

// Development or Distribution
const (
	Invalid Distribution = iota
	Development
	AdHoc
	Enterprise
	AppStore
	Unsupported
)

// String returns a string representation of Distribution
func (d *Distribution) String() string {
	switch *d {
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

func failf(s string, args ...interface{}) {
	log.Errorf(s, args...)
	os.Exit(1)
}

func main() {
	config, err := ParseConfig()
	if err != nil {
		failf(err.Error())
	}
	commonName := ""
	distribution := Development

	var URLs []P12URL
	for i, passphrase := range config.CertificatePassphrases() {
		URLs = append(URLs, P12URL{
			URL:        config.CertificateURLs()[i],
			Passphrase: passphrase,
		})
	}

	certificates, err := DownloadCertificates(URLs)
	if err != nil {
		failf(err.Error())
	}

	certificateType := distribution.Category()
	log.Infof("Filtering certificates for selected distribution method (%s), certificate type (%s), certificate common name (%s) and developer Team ID (%s)", distribution, certificateType, commonName, config.TeamID)

	filteredCertificates, err := FilterCertificates(certificates, certificateType, "", "")
	if err != nil {
		failf("No valid certificates found, error: %s", err)
	}
	log.Debugf("Filtered certificates: %s", filteredCertificates)

	if len(filteredCertificates) > 1 {
		log.Warnf("Multiple certificates for selected distribution, common name and Team ID: %s", filteredCertificates)
	} else if len(filteredCertificates) == 0 {
		log.Infof("Selected app distribution type (%s) needs certificate type: %s", distribution, certificateType)
		log.Warnf(fmt.Sprintf("Maybe you forgot to provide a %s type certificate.\n", certificateType.String()) +
			"Upload a %s type certificate (.p12) on the workflow editor's CodeSign tab.")
		failf("No valid certificates found for distribution (%s), common name (%s) and Team ID (%s)", distribution, commonName, config.TeamID)
	}

}
