package main

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/certificateutil"
)

// CertificateType is an Apple code signing certifcate distribution type
type CertificateType int

// Development or Distribution
const (
	InvalidCertificateType CertificateType = iota
	DevelopmentCertificate
	DistributionCertificate
)

// ToString returns a string representation of CertificateType
func (t *CertificateType) String() string {
	switch *t {
	case DevelopmentCertificate:
		return "Development"
	case DistributionCertificate:
		return "Distribution"
	default:
		return "invalid"
	}
}

// AppStoreConnectCertificate is certificate present on Apple App Store Connect API, could match a local certificate
type AppStoreConnectCertificate struct {
	certificate       certificateutil.CertificateInfoModel
	appStoreConnectID string
}

// GetMatchingCertificates returns validated and matching with App Store Connect API certificates
func GetMatchingCertificates(certificates []certificateutil.CertificateInfoModel, AppStoreConnectCertificates map[CertificateType][]AppStoreConnectCertificate, distribution ProfileType, commonName, teamID string) (map[CertificateType][]AppStoreConnectCertificate, error) {
	fmt.Println()
	log.Infof("Filtering out invalid or duplicated common name certificates.")

	preFilteredCerts := certificateutil.FilterValidCertificateInfos(certificates)

	log.Debugf("Ignoring expired or not yet valid certificates: %s", preFilteredCerts.InvalidCertificates)
	log.Warnf("Ignoring duplicated certificates with the same common name: %s", preFilteredCerts.DuplicatedCertificates)
	log.Infof("Valid and deduplicated common name certificates: %s", preFilteredCerts.ValidCertificates)

	fmt.Println()
	log.Infof("Filtering certificates for selected distribution method (%s), certificate name (%s) and developer Team ID (%s)", distribution, commonName, teamID)

	localCertificates := map[CertificateType][]certificateutil.CertificateInfoModel{}
	for _, certType := range []CertificateType{DevelopmentCertificate, DistributionCertificate} {
		localCertificates[certType] = filterCertificates(preFilteredCerts.ValidCertificates, certType, commonName, teamID)
	}

	requiredCertificateTypes := []CertificateType{DevelopmentCertificate}
	if distribution != Development {
		log.Infof("Selected app distribution type %s requires both Development and Distribution certificate types.", distribution)
		requiredCertificateTypes = append(requiredCertificateTypes, DistributionCertificate)
	}

	for _, certificateType := range requiredCertificateTypes {
		if len(localCertificates[certificateType]) > 1 {
			log.Warnf("Multiple %s type certificates with name (%s) and Team ID (%s):", certificateType.String(), commonName, teamID)
			for i, cert := range localCertificates[certificateType] {
				log.Warnf("certificate number %s, name: %s, serial: %s, expiry date: %s", i, cert.CommonName, cert.Serial, cert.EndDate)
			}
		} else if len(localCertificates[certificateType]) == 0 {
			log.Warnf(fmt.Sprintf("Maybe you forgot to provide a %s type certificate.\n", certificateType.String()) +
				fmt.Sprintf("Upload a %s type certificate (.p12) on the Code Signing tab of the Workflow Editor.", certificateType.String()))
			return map[CertificateType][]AppStoreConnectCertificate{}, fmt.Errorf("no valid %s type certificates uploaded with Team ID (%s), name (%s)", certificateType.String(), teamID, commonName)
		}
	}

	for certType, certs := range AppStoreConnectCertificates {
		log.Debugf("App Store Connect %s certificates:", certType)
		for _, cert := range certs {
			log.Debugf("certificate Serial: %s, Name: %s, Team ID: %s, Team: %s, Expiration: %s", cert.certificate.Serial, cert.certificate.CommonName, cert.certificate.TeamID, cert.certificate.TeamName, cert.certificate.EndDate)
		}
	}

	fmt.Println()
	log.Infof("Matching local certificates with App Store Connect Certificates:")

	matchingCerts := map[CertificateType][]AppStoreConnectCertificate{}
	for _, certType := range []CertificateType{DevelopmentCertificate, DistributionCertificate} {
		matchingCerts[certType] = matchLocalCertificatesToConnectCertificates(localCertificates[certType], AppStoreConnectCertificates[certType])
	}

	return matchingCerts, nil
}

// filterCertificates returns the certificates matching to the given common name, developer team ID, and distribution type.
func filterCertificates(certificates []certificateutil.CertificateInfoModel, certificateType CertificateType, commonName, teamID string) []certificateutil.CertificateInfoModel {
	// filter by distribution type
	var filteredCertificates []certificateutil.CertificateInfoModel
	for _, certificate := range certificates {
		if certificateType == DistributionCertificate && isDistributionCertificate(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		} else if certificateType != DistributionCertificate && !isDistributionCertificate(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		}
	}

	log.Debugf("Valid certificates with type %s:", certificateType)
	for _, cert := range filteredCertificates {
		log.Debugf("certificate Serial: %s, Name: %s, Team ID: %s, Team: %s, Expiration: %s", cert.Serial, cert.CommonName, cert.TeamID, cert.TeamName, cert.EndDate)
	}

	if len(filteredCertificates) == 0 {
		return nil
	}

	// filter by team
	if teamID != "" {
		certsByTeam := mapCertsToTeams(filteredCertificates)
		filteredCertificates = certsByTeam[teamID]
	}

	log.Debugf("Valid certificates with type %s, Team ID: %s", certificateType, teamID)
	for _, cert := range filteredCertificates {
		log.Debugf("certificate Serial: %s, Name: %s, Team ID: %s, Team: %s, Expiration: %s", cert.Serial, cert.CommonName, cert.TeamID, cert.TeamName, cert.EndDate)
	}

	if len(filteredCertificates) == 0 {
		return nil
	}

	// filter by name
	if commonName != "" {
		nameToCertificates := mapCertsToNames(certificates)

		var matchingNameCerts []certificateutil.CertificateInfoModel
		for name, certificates := range nameToCertificates {
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(commonName)) {
				matchingNameCerts = append(matchingNameCerts, certificates...)
			}
		}
		filteredCertificates = matchingNameCerts
	}

	log.Debugf("Valid certificates with type %s, Team ID: %s, Name: ", certificateType, teamID, commonName)
	for _, cert := range filteredCertificates {
		log.Debugf("certificate Serial: %s, Name: %s, Team ID: %s, Team: %s, Expiration: %s", cert.Serial, cert.CommonName, cert.TeamID, cert.TeamName, cert.EndDate)
	}

	return filteredCertificates
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

func matchLocalCertificatesToConnectCertificates(localCertificates []certificateutil.CertificateInfoModel, connectCertificates []AppStoreConnectCertificate) []AppStoreConnectCertificate {
	hashToConnectCertificates := map[string]AppStoreConnectCertificate{}
	for _, cert := range connectCertificates {
		hashToConnectCertificates[cert.certificate.SHA1Fingerprint] = cert
	}

	nameToConnectCertificates := map[string][]AppStoreConnectCertificate{}
	for _, cert := range connectCertificates {
		nameToConnectCertificates[cert.certificate.CommonName] = append(nameToConnectCertificates[cert.certificate.CommonName], cert)
	}

	for _, localCert := range localCertificates {
		for _, connectCert := range nameToConnectCertificates[localCert.CommonName] {
			var latestConnectCert *AppStoreConnectCertificate
			if connectCert.certificate.EndDate.After(localCert.EndDate) {
				if latestConnectCert == nil {
					latestConnectCert = &connectCert
				} else if connectCert.certificate.EndDate.After(latestConnectCert.certificate.EndDate) {
					latestConnectCert = &connectCert
				}
			}
			if latestConnectCert != nil {
				log.Warnf("Provided an older version of certificate with name (%s), expiry date: %s, serial: %s", localCert.CommonName, localCert.EndDate, localCert.Serial)
				log.Warnf("The most recent version of the certificate found on Apple Developer Portal: expiry date: %s, serial: %s", latestConnectCert.certificate.EndDate, latestConnectCert.certificate.Serial)
				log.Warnf("Please upload this version to Bitrise.")
			}
		}
	}

	var matchingCertificates []AppStoreConnectCertificate
	for _, localCert := range localCertificates {
		connectCert, found := hashToConnectCertificates[localCert.SHA1Fingerprint]
		if !found {
			log.Warnf("Certificate (name: %s, serial: %s, expiry date: %s) not found on Apple Developer Portal.", localCert.CommonName, localCert.Serial, localCert.EndDate)
			continue
		}
		matchingCertificates = append(matchingCertificates, connectCert)
	}

	return matchingCertificates
}
