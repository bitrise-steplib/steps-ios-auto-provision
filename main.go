package main

import (
	"os"

	"github.com/bitrise-io/go-utils/log"
)

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
