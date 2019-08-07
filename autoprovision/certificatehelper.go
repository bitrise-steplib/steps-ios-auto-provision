package autoprovision

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

// APICertificate is certificate present on Apple App Store Connect API, could match a local certificate
type APICertificate struct {
	Certificate certificateutil.CertificateInfoModel
	ID          string
}

// CertificateSource ...
type CertificateSource struct {
	client                       *appstoreconnect.Client
	queryCertificateBySerialFunc func(*appstoreconnect.Client, *big.Int) (APICertificate, error)
	queryAllCertificatesFunc     func(*appstoreconnect.Client) (map[appstoreconnect.CertificateType][]APICertificate, error)
}

// APIClient ...
func APIClient(client *appstoreconnect.Client) CertificateSource {
	return CertificateSource{
		client:                       client,
		queryCertificateBySerialFunc: queryCertificateBySerial,
		queryAllCertificatesFunc:     queryAllIOSCertificates,
	}
}

func (c *CertificateSource) queryCertificateBySerial(serial *big.Int) (APICertificate, error) {
	return c.queryCertificateBySerialFunc(c.client, serial)
}

func (c *CertificateSource) queryAllCertificates() (map[appstoreconnect.CertificateType][]APICertificate, error) {
	return c.queryAllCertificatesFunc(c.client)
}

// queryAllIOSCertificates returns all iOS certificates from App Store Connect API
func queryAllIOSCertificates(client *appstoreconnect.Client) (map[appstoreconnect.CertificateType][]APICertificate, error) {
	typeToCertificates := map[appstoreconnect.CertificateType][]APICertificate{}

	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
		certs, err := queryCertificatesByType(client, certType)
		if err != nil {
			return map[appstoreconnect.CertificateType][]APICertificate{}, fmt.Errorf("failed to query certificates, error: %s", err)
		}
		typeToCertificates[certType] = certs
	}

	return typeToCertificates, nil
}

func queryCertificatesByType(client *appstoreconnect.Client, certificateType appstoreconnect.CertificateType) ([]APICertificate, error) {
	nextPageURL := ""
	var responseCertificates []appstoreconnect.Certificate
	for {
		response, err := client.Provisioning.ListCertificates(&appstoreconnect.ListCertificatesOptions{
			FilterCertificateType: certificateType,
			Limit:                 10,
			Next:                  nextPageURL,
		})
		if err != nil {
			return nil, err
		}
		responseCertificates = append(responseCertificates, response.Data...)

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			return parseCertificatesResponse(responseCertificates)
		}
	}
}

func queryCertificateBySerial(client *appstoreconnect.Client, serial *big.Int) (APICertificate, error) {
	response, err := client.Provisioning.FetchCertificate(serial.Text(16))
	if err != nil {
		return APICertificate{}, err
	}

	certs, err := parseCertificatesResponse([]appstoreconnect.Certificate{response})
	if err != nil {
		return APICertificate{}, err
	}
	return certs[0], nil
}

func parseCertificatesResponse(response []appstoreconnect.Certificate) ([]APICertificate, error) {
	var certifacteInfos []APICertificate
	for _, connectCertResponse := range response {
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

			certifacteInfos = append(certifacteInfos, APICertificate{
				Certificate: certInfo,
				ID:          connectCertResponse.ID,
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

// GetValidCertificates ...
func GetValidCertificates(localCertificates []certificateutil.CertificateInfoModel, client CertificateSource, requiredCertificateTypes map[appstoreconnect.CertificateType]bool, teamID string, logAllAPICerts bool) (map[appstoreconnect.CertificateType][]APICertificate, error) {
	typeToLocalCerts, err := GetValidLocalCertificates(localCertificates, teamID)
	if err != nil {
		return nil, err
	}

	log.Debugf("Certificate types required for Development: %t; Distribution: %t", requiredCertificateTypes[appstoreconnect.IOSDevelopment], requiredCertificateTypes[appstoreconnect.IOSDistribution])

	for certificateType, requried := range requiredCertificateTypes {
		if !requried {
			continue
		}
		if len(typeToLocalCerts[certificateType]) > 1 {
			log.Warnf(`Multiple %s type certificates with Team ID "%s": %s`,
				certificateType, teamID, certsToString(typeToLocalCerts[certificateType]))
		} else if len(typeToLocalCerts[certificateType]) == 0 {
			log.Warnf("Maybe you forgot to provide a %s type certificate.\n", certificateType)
			log.Warnf("Upload a %s type certificate (.p12) on the Code Signing tab of the Workflow Editor.", certificateType)
			return map[appstoreconnect.CertificateType][]APICertificate{}, fmt.Errorf("no valid %s type certificates uploaded with Team ID (%s)", certificateType, teamID)
		}
	}

	if logAllAPICerts {
		if err := LogAllAPICertificates(client, typeToLocalCerts); err != nil {
			return nil, fmt.Errorf("failed to log all Developer Portal certificates, error: %s", err)
		}
	}

	log.Debugf("")
	log.Debugf("Querying Apple Developer Portal for matching certificates.")

	validAPICertificates := map[appstoreconnect.CertificateType][]APICertificate{}
	for certificateType, validLocalCertificates := range typeToLocalCerts {
		matchingCertificates, err := MatchLocalToAPICertificates(client, certificateType, validLocalCertificates)
		if err != nil {
			return nil, err
		}

		log.Debugf("Certificates type %s having matches on Developer Portal:", certificateType)
		for _, cert := range matchingCertificates {
			log.Debugf("%s", cert.Certificate)
		}

		if requiredCertificateTypes[certificateType] && len(matchingCertificates) == 0 {
			if !logAllAPICerts {
				if err := LogAllAPICertificates(client, typeToLocalCerts); err != nil {
					log.Errorf("failed to log all Developer Portal certificates, error: %s", err)
				}
			}

			return nil, fmt.Errorf("not found any of the following %s certificates uploaded to Bitrise on Developer Portal: %s", certificateType, localCertificates)
		}

		validAPICertificates[certificateType] = matchingCertificates
	}

	return validAPICertificates, nil
}

// GetValidLocalCertificates returns validated and deduplicated local certificates
func GetValidLocalCertificates(certificates []certificateutil.CertificateInfoModel, teamID string) (map[appstoreconnect.CertificateType][]certificateutil.CertificateInfoModel, error) {
	log.Debugf("")
	log.Debugf("Filtering out invalid or duplicated name certificates.")

	preFilteredCerts := certificateutil.FilterValidCertificateInfos(certificates)

	if len(preFilteredCerts.InvalidCertificates) != 0 {
		log.Debugf("Ignoring expired or not yet valid certificates: %s", preFilteredCerts.InvalidCertificates)
	}
	if len(preFilteredCerts.DuplicatedCertificates) != 0 {
		log.Warnf("Ignoring duplicated certificates with the same name: %s", preFilteredCerts.DuplicatedCertificates)
	}
	log.Debugf("Valid and deduplicated common name certificates: %s", certsToString(preFilteredCerts.ValidCertificates))

	log.Debugf("")
	log.Debugf(`Filtering certificates for developer Team ID (%s)`, teamID)

	localCertificates := map[appstoreconnect.CertificateType][]certificateutil.CertificateInfoModel{}
	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
		localCertificates[certType] = filterCertificates(preFilteredCerts.ValidCertificates, certType, teamID)
	}

	return localCertificates, nil
}

// MatchLocalToAPICertificates ...
func MatchLocalToAPICertificates(client CertificateSource, certificateType appstoreconnect.CertificateType, localCertificates []certificateutil.CertificateInfoModel) ([]APICertificate, error) {
	var matchingCertificates []APICertificate

	for _, localCert := range localCertificates {
		log.Debugf("Looking for certificate on Developer Portal: %s", localCert)

		cert, err := client.queryCertificateBySerial(localCert.Certificate.SerialNumber)
		if err != nil {
			log.Warnf("Certificate not found on Developer Portal, %s", err)
			continue
		}

		log.Debugf("Found. ID: %s, %s ", cert.ID, cert.Certificate)
		matchingCertificates = append(matchingCertificates, cert)
	}

	return matchingCertificates, nil
}

// LogAllAPICertificates ...
func LogAllAPICertificates(client CertificateSource, localCertificates map[appstoreconnect.CertificateType][]certificateutil.CertificateInfoModel) error {
	certificates, err := client.queryAllCertificates()
	if err != nil {
		return fmt.Errorf("failed to query certificates on Developer Portal: %s", err)
	}

	for certType, certs := range certificates {
		log.Debugf("Developer Portal %s certificates:", certType)
		for _, cert := range certs {
			log.Debugf("%s", cert.Certificate)
		}
	}

	for _, certType := range []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution} {
		logUpdatedAPICertificates(localCertificates[certType], certificates[certType])
	}

	return nil
}

func logUpdatedAPICertificates(localCertificates []certificateutil.CertificateInfoModel, APIcertificates []APICertificate) bool {
	nameToAPICertificates := map[string][]APICertificate{}
	for _, cert := range APIcertificates {
		nameToAPICertificates[cert.Certificate.CommonName] = append(nameToAPICertificates[cert.Certificate.CommonName], cert)
	}

	existUpdated := false
	for _, localCert := range localCertificates {
		if len(nameToAPICertificates[localCert.CommonName]) == 0 {
			continue
		}

		var latestAPICert *APICertificate
		for _, APICert := range nameToAPICertificates[localCert.CommonName] {
			if APICert.Certificate.EndDate.After(localCert.EndDate) &&
				(latestAPICert == nil || APICert.Certificate.EndDate.After(latestAPICert.Certificate.EndDate)) {
				latestAPICert = &APICert
			}
		}

		if latestAPICert != nil {
			existUpdated = true
			log.Warnf("Provided an older version of certificate %s", localCert)
			log.Warnf("The most recent version of the certificate found on Developer Portal: expiry date: %s, serial: %s", latestAPICert.Certificate.EndDate, latestAPICert.Certificate.Serial)
			log.Warnf("Please upload this version to Bitrise.")
		}
	}
	return existUpdated
}

// filterCertificates returns the certificates matching to the given common name, developer team ID, and distribution type.
func filterCertificates(certificates []certificateutil.CertificateInfoModel, certificateType appstoreconnect.CertificateType, teamID string) []certificateutil.CertificateInfoModel {
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

	log.Debugf("Valid certificates with type %s, Team ID: (%s) %s ", certificateType, teamID, certsToString(filteredCertificates))

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
