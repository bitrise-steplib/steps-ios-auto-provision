package autoprovision

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// CertificateType is an Apple code signing certifcate distribution type
type CertificateType int

// Development or Distribution
const (
	InvalidCertificateType CertificateType = iota
	DevelopmentCertificate
	DistributionCertificate
)

// String returns a string representation of CertificateType
func (t CertificateType) String() string {
	switch t {
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

// QueryAppStoreConnectCertificates returns certificates from App Store Connect API
func QueryAppStoreConnectCertificates(client *appstoreconnect.Client) (map[appstoreconnect.CertificateType][]AppStoreConnectCertificate, error) {
	typeToCertificates := map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{}

	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
		certs, err := queryAppStoreConnectCertificates(client, certType)
		if err != nil {
			return map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{}, fmt.Errorf("failed to get certificates from App Store Connect API, error: %s", err)
		}
		typeToCertificates[certType] = certs
	}

	return typeToCertificates, nil
}

func queryAppStoreConnectCertificates(client *appstoreconnect.Client, certificateType appstoreconnect.CertificateType) ([]AppStoreConnectCertificate, error) {
	response, err := client.Provisioning.ListCertificates(&appstoreconnect.ListCertificatesOptions{
		FilterCertificateType: certificateType,
	})
	if err != nil {
		return nil, err
	}

	var certifacteInfos []AppStoreConnectCertificate
	for _, connectCertResponse := range response.Data {
		if connectCertResponse.Type == "certificates" {
			certificateData, err := base64.StdEncoding.DecodeString(connectCertResponse.Attributes.CertificateContent)
			if err != nil {
				return nil, fmt.Errorf("failed to decode certificate connect, error: %s", err)
			}

			cert, err := x509.ParseCertificate(certificateData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate, error: %s", err)
			}

			certInfo := certificateutil.NewCertificateInfo(*cert, nil)

			certifacteInfos = append(certifacteInfos, AppStoreConnectCertificate{
				certificate:       certInfo,
				appStoreConnectID: connectCertResponse.ID,
			})
		}
	}

	return certifacteInfos, nil
}

func certsToString(certs []certificateutil.CertificateInfoModel) string {
	certInfo := "[\n"

	for _, cert := range certs {
		certInfo += cert.String() + "\n"
	}
	certInfo += "]"

	return certInfo
}

// GetMatchingCertificates returns validated and matching with App Store Connect API certificates
func GetMatchingCertificates(certificates []certificateutil.CertificateInfoModel, AppStoreConnectCertificates map[appstoreconnect.CertificateType][]AppStoreConnectCertificate, requiredCertificateTypes []appstoreconnect.CertificateType, typeToName map[appstoreconnect.CertificateType]string, teamID string) (map[appstoreconnect.CertificateType][]AppStoreConnectCertificate, error) {
	fmt.Println()
	log.Infof("Filtering out invalid or duplicated name certificates.")

	preFilteredCerts := certificateutil.FilterValidCertificateInfos(certificates)

	if len(preFilteredCerts.InvalidCertificates) != 0 {
		log.Debugf("Ignoring expired or not yet valid certificates: %s")
	}
	if len(preFilteredCerts.DuplicatedCertificates) != 0 {
		log.Warnf("Ignoring duplicated certificates with the same name: %s")
	}
	log.Infof("Valid and deduplicated common name certificates: %s", certsToString(preFilteredCerts.ValidCertificates))

	fmt.Println()
	log.Infof("Filtering certificates for required certificate types (%s), certificate name (development: %s; distribution: %s) and developer Team ID (%s)", requiredCertificateTypes, typeToName[appstoreconnect.IOSDevelopment], typeToName[appstoreconnect.IOSDistribution], teamID)

	localCertificates := map[appstoreconnect.CertificateType][]certificateutil.CertificateInfoModel{}
	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
		localCertificates[certType] = filterCertificates(preFilteredCerts.ValidCertificates, certType, typeToName[certType], teamID)
	}

	for _, certificateType := range requiredCertificateTypes {
		if len(localCertificates[certificateType]) > 1 {
			log.Warnf("Multiple %s type certificates with name (%s) and Team ID (%s):", certificateType, typeToName[certificateType], teamID)
			for i, cert := range localCertificates[certificateType] {
				log.Warnf("certificate number %s, name: %s, serial: %s, expiry date: %s", i, cert.CommonName, cert.Serial, cert.EndDate)
			}
		} else if len(localCertificates[certificateType]) == 0 {
			log.Warnf("Maybe you forgot to provide a %s type certificate.\n", certificateType)
			log.Warnf("Upload a %s type certificate (.p12) on the Code Signing tab of the Workflow Editor.", certificateType)
			return map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{}, fmt.Errorf("no valid %s type certificates uploaded with Team ID (%s), name (%s)", certificateType, teamID, typeToName[certificateType])
		}
	}

	for certType, certs := range AppStoreConnectCertificates {
		log.Debugf("App Store Connect %s certificates:", certType)
		for _, cert := range certs {
			log.Debugf("%s", cert.certificate)
		}
	}

	fmt.Println()
	log.Infof("Matching certificates present with App Store Connect certificates")

	matchingCerts := map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{}
	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
		matchingCerts[certType] = matchLocalCertificatesToConnectCertificates(localCertificates[certType], AppStoreConnectCertificates[certType])
		log.Debugf("Certificates type %s having matches on App Store Connect", certType)
		for _, cert := range matchingCerts[certType] {
			log.Debugf("%s", cert.certificate)
		}
	}

	for _, certType := range requiredCertificateTypes {
		if len(matchingCerts[certType]) == 0 {
			return nil, fmt.Errorf("not found any of the following %s certificates uploaded to Bitrise on App Store Connect: %s", certType, localCertificates[certType])
		}
	}

	return matchingCerts, nil
}

// filterCertificates returns the certificates matching to the given common name, developer team ID, and distribution type.
func filterCertificates(certificates []certificateutil.CertificateInfoModel, certificateType appstoreconnect.CertificateType, commonName, teamID string) []certificateutil.CertificateInfoModel {
	// filter by distribution type
	var filteredCertificates []certificateutil.CertificateInfoModel
	for _, certificate := range certificates {
		if certificateType == appstoreconnect.IOSDistribution && isDistributionCertificate(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		} else if certificateType == appstoreconnect.IOSDevelopment && !isDistributionCertificate(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		}
	}

	log.Debugf("Valid certificates with type %s: %s", certificateType, certsToString(filteredCertificates))

	if len(filteredCertificates) == 0 {
		return nil
	}

	// filter by team
	if teamID != "" {
		certsByTeam := mapCertsToTeams(filteredCertificates)
		filteredCertificates = certsByTeam[teamID]
	}

	log.Debugf("Valid certificates with type %s, Team ID: (%s): %s", certificateType, teamID, certsToString(filteredCertificates))

	if len(filteredCertificates) == 0 {
		return nil
	}

	// filter by name
	if commonName != "" {
		nameToCertificates := mapCertsToNames(filteredCertificates)

		var matchingNameCerts []certificateutil.CertificateInfoModel
		for name, nameCerts := range nameToCertificates {
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(commonName)) {
				matchingNameCerts = append(matchingNameCerts, nameCerts...)
			}
		}
		filteredCertificates = matchingNameCerts
	}

	log.Debugf("Valid certificates with type %s, Team ID: (%s), Name: (%s) %s ", certificateType, teamID, commonName, certsToString(filteredCertificates))

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
	return strings.HasPrefix(strings.ToLower(cert.CommonName), strings.ToLower("iPhone Distribution")) ||
		strings.HasPrefix(strings.ToLower(cert.CommonName), strings.ToLower("Apple Distribution"))
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
		connectCerts := nameToConnectCertificates[localCert.CommonName]
		if len(connectCerts) == 0 {
			continue
		}

		var latestConnectCert *AppStoreConnectCertificate
		for _, connectCert := range nameToConnectCertificates[localCert.CommonName] {
			if connectCert.certificate.EndDate.After(localCert.EndDate) &&
				(latestConnectCert == nil || connectCert.certificate.EndDate.After(latestConnectCert.certificate.EndDate)) {
				latestConnectCert = &connectCert
			}
		}

		if latestConnectCert != nil {
			log.Warnf("Provided an older version of certificate $s", localCert)
			log.Warnf("The most recent version of the certificate found on Apple Developer Portal: expiry date: %s, serial: %s", latestConnectCert.certificate.EndDate, latestConnectCert.certificate.Serial)
			log.Warnf("Please upload this version to Bitrise.")
		}
	}

	var matchingCertificates []AppStoreConnectCertificate
	for _, localCert := range localCertificates {
		connectCert, found := hashToConnectCertificates[localCert.SHA1Fingerprint]
		if !found {
			log.Warnf("Certificate not found on Apple Developer Portal: %s", localCert)
			continue
		}
		matchingCertificates = append(matchingCertificates, connectCert)
	}

	return matchingCertificates
}
