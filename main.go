package main

import (
	"os"

	"github.com/bitrise-io/go-utils/log"
)

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
func (d *ProfileType) String() string {
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
	teamID := ""
	distribution := Development

	var URLs []p12URL
	for i, passphrase := range config.CertificatePassphrases() {
		URLs = append(URLs, p12URL{
			URL:        config.CertificateURLs()[i],
			Passphrase: passphrase,
		})
	}

	// client, err := initAppStoreConnectClient()
	// if err != nil {
	// 	failf("%s", err)
	// }

	localCertificates, err := DownloadLocalCertificates(URLs)
	if err != nil {
		failf(err.Error())
	}

	appStoreConnectCertificates, err := QueryAppStoreConnectCertificates(nil)
	if err != nil {
		failf(err.Error())
	}

	_, err = GetMatchingCertificates(localCertificates, appStoreConnectCertificates, distribution, commonName, teamID)
	if err != nil {
		failf("%s", err)
	}
}
