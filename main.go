package main

import (
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-ios-auto-provision/certstore"
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

	var URLs []certstore.P12URL
	for i, passphrase := range config.CertificatePassphrases() {
		URLs = append(URLs, certstore.P12URL{
			URL:        config.CertificateURLs()[i],
			Passphrase: passphrase,
		})
	}

	_, err = certstore.Download(URLs)

}
