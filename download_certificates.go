package main

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
	"github.com/bitrise-tools/go-xcode/certificateutil"
)

type p12URL struct {
	URL, Passphrase string
}

// DownloadLocalCertificates downloads and parses a list of p12 files
func DownloadLocalCertificates(URLs []p12URL) ([]certificateutil.CertificateInfoModel, error) {
	var certInfos []certificateutil.CertificateInfoModel
	for i, p12 := range URLs {
		log.Debugf("Downloading p12 file number %d from %s", i, p12.URL)
		p12CertInfos, err := downloadPKCS12(p12.URL, p12.Passphrase)
		if err != nil {
			return nil, err
		}

		log.Debugf("Codesign identities included:")
		for _, cert := range p12CertInfos {
			log.Debugf("certificate Serial: %s, Name: %s, Team ID: %s, Team: %s, Expiration: %s", cert.Serial, cert.CommonName, cert.TeamID, cert.TeamName, cert.EndDate)
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

// QueryAppStoreConnectCertificates returns certificates from App Store Connect API
func QueryAppStoreConnectCertificates(client *appstoreconnect.Client) (map[CertificateType][]AppStoreConnectCertificate, error) {
	typeToCertificates := map[CertificateType][]AppStoreConnectCertificate{}

	for _, certType := range []CertificateType{DevelopmentCertificate, DistributionCertificate} {
		certs, err := queryAppStoreConnectCertificates(client, certType)
		if err != nil {
			return map[CertificateType][]AppStoreConnectCertificate{}, fmt.Errorf("failed to get certificates from App Store Connect API, error: %s", err)
		}
		typeToCertificates[certType] = certs
	}

	return typeToCertificates, nil
}

func queryAppStoreConnectCertificates(client *appstoreconnect.Client, certificatesType CertificateType) ([]AppStoreConnectCertificate, error) {
	var certTypeFilter appstoreconnect.CertificateType
	switch certificatesType {
	case DevelopmentCertificate:
		certTypeFilter = appstoreconnect.IOSDevelopment
	case DistributionCertificate:
		certTypeFilter = appstoreconnect.IOSDistribution
	default:
		return nil, errors.New("unsupported certificate type provided")
	}

	response, err := client.Provisioning.ListCertificates(&appstoreconnect.ListCertificatesOptions{
		FilterCertificateType: certTypeFilter,
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
