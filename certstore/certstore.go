package certstore

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-xcode/certificateutil"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// Store represents a certificate storage.
type Store struct {
	Certificates []certificateutil.CertificateInfoModel
}

// New creates a Store.
func New() (*Store, error) {
	return &Store{}, nil
}

// Download downloads a certificate and opens
func (s Store) Download(certificateURL, passphrase string) ([]certificateutil.CertificateInfoModel, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("certstore")
	if err != nil {
		return nil, err
	}
	pth := filepath.Join(tmpDir, "Certificate.p12")
	if err := download(certificateURL, pth); err != nil {
		return nil, err
	}

	certs, err := certificateutil.NewCertificateInfosFromPKCS12(pth, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate (%s), err: %s", certificateURL, err)
	}

	return certs, err
}

// Add adds Certificate to the Store
func (s Store) Add(cert certificateutil.CertificateInfoModel) {
	for i, c := range s.Certificates {
		if c.CommonName == cert.CommonName {
			if cert.EndDate.After(c.EndDate) {
				s.Certificates[i] = cert
			}
		}
	}
}

// ToP12 creates a p12 file from a certificate
func (s Store) ToP12(cert certificateutil.CertificateInfoModel) (string, error) {
	b, err := pkcs12.Encode(rand.Reader, nil, &cert.Certificate, nil, "")
	if err != nil {
		return "", err
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("certstore")
	if err != nil {
		return "", err
	}
	pth := filepath.Join(tmpDir, "Certificate.p12")
	if err := fileutil.WriteBytesToFile(pth, b); err != nil {
		return "", err
	}
	return pth, err
}

// Find ...
func (s Store) Find(name, team string, distribution bool) ([]certificateutil.CertificateInfoModel, error) {
	certs := s.Certificates

	// filter by distribution type
	var certsByDistribution []certificateutil.CertificateInfoModel
	for _, c := range certs {
		isDistribution := isDistributionCertificate(c)
		if distribution && isDistribution {
			certsByDistribution = append(certsByDistribution, c)
		} else if !distribution && !isDistribution {
			certsByDistribution = append(certsByDistribution, c)
		}
	}

	if len(certs) == 0 {
		if distribution {
			return nil, errors.New("no development certificate found")
		} else {
			return nil, errors.New("no distribution certificate found")
		}
	}

	// filter by team
	if team != "" {
		certsByTeam := mapCertsToTeams(s.Certificates)
		certs = certsByTeam[team]

		if len(certs) == 0 {
			return nil, fmt.Errorf("no certificate found for team: %s", team)
		}
	}

	// filter by name
	if name != "" {
		certsByName := mapCertsToNames(s.Certificates)
		certs = certsByName[team]

		if len(certs) == 0 {
			return nil, fmt.Errorf("no certificate found for name: %s", name)
		}
	}

	return certs, nil
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
	return strings.Contains(strings.ToLower(cert.CommonName), strings.ToLower("iPhone Distribution"))
}

func download(src, dest string) error {
	url, err := url.Parse(src)
	if err != nil {
		return fmt.Errorf("failed to parse url (%s), error: %s", src, err)
	}

	// Local file
	if url.Scheme == "file" {
		src := strings.Replace(src, url.Scheme+"://", "", -1)

		if err := command.CopyFile(src, dest); err != nil {
			return fmt.Errorf("failed to copy (%s) to (%s)", src, dest)
		}
		return nil
	}

	// Remote file
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create (%s), error: %s", dest, err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Warnf("failed to close (%s), error: %s", dest, err)
		}
	}()

	resp, err := http.Get(src)
	if err != nil {
		return fmt.Errorf("failed to download (%s), error: %s", src, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("failed to close (%s) body, error: %s", src, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download (%s) failed with status code (%d)", src, resp.StatusCode)
	}

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response (%s), error: %s", src, err)
	}

	// ioutil.WriteFile truncates a destination file if exists
	if err := ioutil.WriteFile(out.Name(), buff, 0644); err != nil {
		return fmt.Errorf("failed to write response (%s) body to file, error: %s", src, err)
	}

	return nil
}
