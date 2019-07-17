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
	Development
	Distribution
)

// String returns a string representation of CertificateType
func (t CertificateType) String() string {
	switch t {
	case Development:
		return "Development"
	case Distribution:
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

// QueryAllIOSCertificates returns all iOS certificates from App Store Connect API
func QueryAllIOSCertificates(client *appstoreconnect.Client) (map[CertificateType][]AppStoreConnectCertificate, error) {
	typeToCertificates := map[CertificateType][]AppStoreConnectCertificate{}

	for _, certType := range []CertificateType{Development, Distribution} {
		var APIcertType appstoreconnect.CertificateType
		switch certType {
		case Development:
			APIcertType = appstoreconnect.IOSDevelopment
		case Distribution:
			APIcertType = appstoreconnect.IOSDistribution
		default:
			return nil, fmt.Errorf("invalid certifiate type")
		}

		certs, err := queryCertificatesByType(client, APIcertType)
		if err != nil {
			return map[CertificateType][]AppStoreConnectCertificate{}, fmt.Errorf("failed to get certificates from App Store Connect API, error: %s", err)
		}
		typeToCertificates[certType] = certs
	}

	return typeToCertificates, nil
}

func queryCertificatesByType(client *appstoreconnect.Client, certificateType appstoreconnect.CertificateType) ([]AppStoreConnectCertificate, error) {
	response, err := client.Provisioning.ListCertificates(&appstoreconnect.ListCertificatesOptions{
		FilterCertificateType: certificateType,
	})
	if err != nil {
		return nil, err
	}

	return parseCertificatesResponse(response)
}

func queryCertificatesBySerial(client *appstoreconnect.Client, serial string) ([]AppStoreConnectCertificate, error) {
	response, err := client.Provisioning.ListCertificates(&appstoreconnect.ListCertificatesOptions{
		FilterSerialNumber: serial,
	})
	if err != nil {
		return nil, err
	}

	return parseCertificatesResponse(response)
}

func parseCertificatesResponse(response *appstoreconnect.CertificatesResponse) ([]AppStoreConnectCertificate, error) {
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

// GetLocalCertificates returns validated and deduplicated local certificates
func GetLocalCertificates(requiredCertificateTypes []CertificateType, certificates []certificateutil.CertificateInfoModel, typeToName map[CertificateType]string, teamID string) (map[CertificateType][]certificateutil.CertificateInfoModel, error) {
	fmt.Println()
	log.Infof("Filtering out invalid or duplicated name certificates.")

	preFilteredCerts := certificateutil.FilterValidCertificateInfos(certificates)

	if len(preFilteredCerts.InvalidCertificates) != 0 {
		log.Debugf("Ignoring expired or not yet valid certificates: %s", preFilteredCerts.InvalidCertificates)
	}
	if len(preFilteredCerts.DuplicatedCertificates) != 0 {
		log.Warnf("Ignoring duplicated certificates with the same name: %s", preFilteredCerts.DuplicatedCertificates)
	}
	log.Infof("Valid and deduplicated common name certificates: %s", certsToString(preFilteredCerts.ValidCertificates))

	fmt.Println()
	log.Infof("Filtering certificates for required certificate types (%s), certificate name (development: %s; distribution: %s) and developer Team ID (%s)", requiredCertificateTypes, typeToName[Development], typeToName[Distribution], teamID)

	localCertificates := map[CertificateType][]certificateutil.CertificateInfoModel{}
	for _, certType := range []CertificateType{Development, Distribution} {
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
			return map[CertificateType][]certificateutil.CertificateInfoModel{}, fmt.Errorf("no valid %s type certificates uploaded with Team ID (%s), name (%s)", certificateType, teamID, typeToName[certificateType])
		}
	}

	return localCertificates, nil
}

func MatchLocalToAPICertificates(client *appstoreconnect.Client, requiredCertificateTypes []CertificateType, localCertificates map[CertificateType][]certificateutil.CertificateInfoModel) (map[CertificateType][]AppStoreConnectCertificate, error) {
	fmt.Println()
	log.Infof("Matching certificates present with Developer Portal certificates")

	matchingCerts := map[CertificateType][]AppStoreConnectCertificate{}
	for _, certType := range []CertificateType{Development, Distribution} {
		var matchingCertificates []AppStoreConnectCertificate

		for _, localCert := range localCertificates[certType] {
			APIcerts, err := queryCertificatesBySerial(client, localCert.Serial)
			if err != nil {
				return nil, fmt.Errorf("failed to query certificate by serial, error: %s", err)
			}
			if len(APIcerts) == 0 {
				log.Warnf("Certificate not found on Developer Portal: %s", localCert)
				continue
			} else if len(APIcerts) > 1 {
				return nil, fmt.Errorf("more than one certificate with serial %s found on Developer Portal", localCert.Serial)
			}

			matchingCertificates = append(matchingCertificates, APIcerts[0])
		}

		log.Debugf("Certificates type %s having matches on Apple Developer Portal:", certType)
		for _, cert := range matchingCerts[certType] {
			log.Debugf("%s", cert.certificate)
		}
	}

	for _, certType := range requiredCertificateTypes {
		if len(matchingCerts[certType]) == 0 {
			return nil, fmt.Errorf("not found any of the following %s certificates uploaded to Bitrise on Developer Portal: %s", certType, localCertificates[certType])
		}
	}

	return matchingCerts, nil
}

func LogAllCertificates(client *appstoreconnect.Client, localCertificates map[CertificateType][]certificateutil.CertificateInfoModel) error {
	APIcertificates, err := QueryAllIOSCertificates(client)
	if err != nil {
		return fmt.Errorf("failed to query certificates on Developer Portal: %s", err)
	}

	for certType, certs := range APIcertificates {
		log.Debugf("Developer Portal %s certificates:", certType)
		for _, cert := range certs {
			log.Debugf("%s", cert.certificate)
		}
	}

	for _, certType := range []CertificateType{Development, Distribution} {
		nameToAPICertificates := map[string][]AppStoreConnectCertificate{}
		for _, cert := range APIcertificates[certType] {
			nameToAPICertificates[cert.certificate.CommonName] = append(nameToAPICertificates[cert.certificate.CommonName], cert)
		}

		for _, localCert := range localCertificates[certType] {
			connectCerts := nameToAPICertificates[localCert.CommonName]
			if len(connectCerts) == 0 {
				continue
			}

			var latestConnectCert *AppStoreConnectCertificate
			for _, connectCert := range nameToAPICertificates[localCert.CommonName] {
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
	}

	return nil
}

// filterCertificates returns the certificates matching to the given common name, developer team ID, and distribution type.
func filterCertificates(certificates []certificateutil.CertificateInfoModel, certificateType CertificateType, commonName, teamID string) []certificateutil.CertificateInfoModel {
	// filter by distribution type
	var filteredCertificates []certificateutil.CertificateInfoModel
	for _, certificate := range certificates {
		if certificateType == Distribution && isDistributionCertificate(certificate) {
			filteredCertificates = append(filteredCertificates, certificate)
		} else if certificateType == Development && !isDistributionCertificate(certificate) {
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
