package certstore

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/certificateutil"
)

// P12URL ...
type P12URL struct {
	URL, Passphrase string
}

// Download ...
func Download(URLs []P12URL) ([]certificateutil.CertificateInfoModel, error) {
	var certInfos []certificateutil.CertificateInfoModel
	for i, p12 := range URLs {
		log.Debugf("Downloading p12 file number %d from %s", i, p12.URL)
		p12CertInfos, err := downloadPKCS12(p12.URL, p12.Passphrase)
		if err != nil {
			return nil, err
		}

		log.Debugf("Codesign identities included:")
		for i, certInfo := range p12CertInfos {
			log.Debugf("Certificate number %d: %s", i, certInfo)
		}

		certInfos = append(certInfos, p12CertInfos...)
	}

	log.Debugf("%d certificates downloaded", len(certInfos))
	return certInfos, nil
}

// downloadPKCS12 downloads a pkcs12 format file and parses certificates and matching private keys.
func downloadPKCS12(certificateURL, passphrase string) ([]certificateutil.CertificateInfoModel, error) {
	contents, err := downloadFile(certificateURL)
	if err != nil {
		return nil, err
	} else if contents == nil {
		return nil, fmt.Errorf("certificate (%s) is empty", certificateURL)
	}

	identities, err := certificateutil.CertificatesFromPKCS12Content(contents, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate (%s), err: %s", certificateURL, err)
	}

	infos := []certificateutil.CertificateInfoModel{}
	for _, identity := range identities {
		if identity.Certificate != nil {
			infos = append(infos, certificateutil.NewCertificateInfo(*identity.Certificate, identity.PrivateKey))
		}
	}

	return infos, nil
}

// FilterLatestValidCertificates ...
func FilterLatestValidCertificates(certificates []certificateutil.CertificateInfoModel) []certificateutil.CertificateInfoModel {
	filteredCertificates := certificateutil.FilterValidCertificateInfos(certificates)

	log.Debugf("Ignoring expired or not yet valid certificates: %s", filteredCertificates.InvalidCertificates)
	log.Warnf("Ignoring duplicated certificates with the same common name: %s", filteredCertificates.DuplicatedCertificates)
	log.Infof("Valid and deduplicated common name certificates: %s", filteredCertificates.ValidCertificates)

	return filteredCertificates.ValidCertificates
}

// CertificateDistributionTarget is an Apple code signing certiifcate distribution type
type CertificateDistributionTarget int

// Development or Distribution
const (
	Invalid CertificateDistributionTarget = iota
	Development
	Distribution
	Unsupported
)

// Find returns a Certificates matching to the given filters.
func Find(certificates []certificateutil.CertificateInfoModel, name, team string, distribution CertificateDistributionTarget) ([]certificateutil.CertificateInfoModel, error) {
	// filter by distribution type
	if distribution != Development && distribution != Distribution {
		return nil, errors.New("invalid certificate distribution type specified")
	}

	var filteredCertificates []certificateutil.CertificateInfoModel
	for _, certificate := range certificates {
		isDistribution := isDistributionCertificate(certificate)
		if distribution == Distribution && isDistribution {
			filteredCertificates = append(filteredCertificates, certificate)
		} else if distribution != Distribution && !isDistribution {
			filteredCertificates = append(filteredCertificates, certificate)
		}
	}

	if len(filteredCertificates) == 0 {
		if distribution == Development {
			return nil, errors.New("no development certificate found")
		}
		return nil, errors.New("no distribution certificate found")
	}

	// filter by team
	if team != "" {
		certsByTeam := mapCertsToTeams(certificates)
		filteredCertificates = certsByTeam[team]

		if len(filteredCertificates) == 0 {
			return nil, fmt.Errorf("no certificate found for team: %s", team)
		}
	}

	// filter by name
	if name != "" {
		certsByName := mapCertsToNames(certificates)
		filteredCertificates = certsByName[team]

		if len(filteredCertificates) == 0 {
			return nil, fmt.Errorf("no certificate found for name: %s", name)
		}
	}

	return filteredCertificates, nil
}

func mapCertsToTeams(certs []certificateutil.CertificateInfoModel) map[string][]certificateutil.CertificateInfoModel {
	m := map[string][]certificateutil.CertificateInfoModel{}
	for _, c := range certs {
		teamCerts := m[c.TeamID]
		m[c.TeamID] = append(teamCerts, c)
	}
	return m
}

func mapCertsToNames(certs []certificateutil.CertificateInfoModel) map[string][]certificateutil.CertificateInfoModel {
	m := map[string][]certificateutil.CertificateInfoModel{}
	for _, c := range certs {
		teamCerts := m[c.CommonName]
		m[c.CommonName] = append(teamCerts, c)
	}
	return m
}

func isDistributionCertificate(cert certificateutil.CertificateInfoModel) bool {
	// Apple certificate types: https://help.apple.com/xcode/mac/current/#/dev80c6204ec)
	return strings.Contains(strings.ToLower(cert.CommonName), strings.ToLower("iPhone Distribution")) ||
		strings.Contains(strings.ToLower(cert.CommonName), strings.ToLower("Apple Distribution"))
}

func downloadFile(src string) ([]byte, error) {
	url, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url (%s), error: %s", src, err)
	}

	// Local file
	if url.Scheme == "file" {
		src := strings.Replace(src, url.Scheme+"://", "", -1)

		return ioutil.ReadFile(src)
	}

	// ToDo: add timeout, retry
	// Remote file
	resp, err := http.Get(src)
	if err != nil {
		return nil, fmt.Errorf("failed to download (%s), error: %s", src, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body, error: %s", src, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download (%s) failed with status code (%d)", src, resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response (%s), error: %s", src, err)
	}
	return contents, nil
}
