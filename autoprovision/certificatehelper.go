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

// APICertificate is certificate present on Apple App Store Connect API, could match a local certificate
type APICertificate struct {
	Certificate certificateutil.CertificateInfoModel
	ID          string
}

type getCertificateBySerial func(*appstoreconnect.Client, *big.Int) (APICertificate, error)

type getAllCertificates func(*appstoreconnect.Client) (map[CertificateType][]APICertificate, error)

type certificateSource struct {
	client                       *appstoreconnect.Client
	queryCertificateBySerialFunc getCertificateBySerial
	queryAllCertificatesFunc     getAllCertificates
}

func APIClient(client *appstoreconnect.Client) certificateSource {
	return certificateSource{
		client:                       client,
		queryCertificateBySerialFunc: queryCertificateBySerial,
		queryAllCertificatesFunc:     queryAllIOSCertificates,
	}
}

func (c *certificateSource) queryCertificateBySerial(serial *big.Int) (APICertificate, error) {
	return c.queryCertificateBySerialFunc(c.client, serial)
}

func (c *certificateSource) queryAllCertificates() (map[CertificateType][]APICertificate, error) {
	return c.queryAllCertificatesFunc(c.client)
}

// queryAllIOSCertificates returns all iOS certificates from App Store Connect API
func queryAllIOSCertificates(client *appstoreconnect.Client) (map[CertificateType][]APICertificate, error) {
	typeToCertificates := map[CertificateType][]APICertificate{}

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
			return map[CertificateType][]APICertificate{}, fmt.Errorf("failed to query certificates, error: %s", err)
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

func GetValidCertificates(localCertificates []certificateutil.CertificateInfoModel, client certificateSource, requiredCertificateTypes map[CertificateType]bool, typeToName map[CertificateType]string, teamID string, logAllCerts bool) (map[CertificateType][]APICertificate, error) {
	typeToCerts, err := GetValidLocalCertificates(localCertificates, requiredCertificateTypes, typeToName, teamID)
	if err != nil {
		return nil, err
	}

	if logAllCerts {
		if err := LogAllAPICertificates(client, typeToCerts); err != nil {
			return nil, fmt.Errorf("failed to log all Developer Portal certificates, error: %s", err)
		}
	}

	validAPICertificates := map[CertificateType][]APICertificate{}
	for certificateType, validLocalCertificates := range typeToCerts {
		matchingCertificates, err := MatchLocalToAPICertificates(client, certificateType, validLocalCertificates)
		if err != nil {
			return nil, err
		}

		log.Debugf("Certificates type %s having matches on Developer Portal:", certificateType)
		for _, cert := range matchingCertificates {
			log.Debugf("%s", cert.Certificate)
		}

		if requiredCertificateTypes[certificateType] && len(matchingCertificates) == 0 {
			if !logAllCerts {
				if err := LogAllAPICertificates(client, typeToCerts); err != nil {
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
func GetValidLocalCertificates(certificates []certificateutil.CertificateInfoModel, requiredCertificateTypes map[CertificateType]bool, typeToName map[CertificateType]string, teamID string) (map[CertificateType][]certificateutil.CertificateInfoModel, error) {
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

	for certificateType, requried := range requiredCertificateTypes {
		if !requried {
			continue
		}
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

func MatchLocalToAPICertificates(client certificateSource, certificateType CertificateType, localCertificates []certificateutil.CertificateInfoModel) ([]APICertificate, error) {
	var matchingCertificates []APICertificate

	for _, localCert := range localCertificates {
		cert, err := client.queryCertificateBySerial(localCert.Certificate.SerialNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to query certificate by serial, error: %s", err)
		}

		matchingCertificates = append(matchingCertificates, cert)
	}

	return matchingCertificates, nil
}

func LogAllAPICertificates(client certificateSource, localCertificates map[CertificateType][]certificateutil.CertificateInfoModel) error {
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

	for _, certType := range []CertificateType{Development, Distribution} {
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
			log.Warnf("Provided an older version of certificate $s", localCert)
			log.Warnf("The most recent version of the certificate found on Developer Portal: expiry date: %s, serial: %s", latestAPICert.Certificate.EndDate, latestAPICert.Certificate.Serial)
			log.Warnf("Please upload this version to Bitrise.")
		}
	}
	return existUpdated
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
