package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
)

// ClientInterface for http client
type ClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// CertificateDownloader provides methods for p12 file downloads
type CertificateDownloader interface {
	DownloadCertificates(URLs []CertificateFileURL) ([]certificateutil.CertificateInfoModel, error)
}

// CertificateDownloaderImpl provides methods for p12 file downloads
type CertificateDownloaderImpl struct {
	httpClient      ClientInterface
	debugLogger     func(string, ...interface{})
	warnLogger      func(string, ...interface{})
	pkcs12Converter func(content []byte, password string) ([]certificateutil.CertificateInfoModel, error)
	readFile        func(filename string) ([]byte, error)
	readAll         func(r io.Reader) ([]byte, error)
}

// DownloadCertificates downloads and parses a list of p12 files
func (c CertificateDownloaderImpl) DownloadCertificates(URLs []CertificateFileURL) ([]certificateutil.CertificateInfoModel, error) {
	var certInfos []certificateutil.CertificateInfoModel

	for i, p12 := range URLs {
		c.debugLogger("Downloading p12 file number %d from %s", i, p12.URL)

		p12CertInfos, err := c.downloadPKCS12(p12.URL, p12.Passphrase)
		if err != nil {
			return nil, err
		}
		c.debugLogger("Codesign identities included:\n%s", autoprovision.CertsToString(p12CertInfos))

		certInfos = append(certInfos, p12CertInfos...)
	}

	return certInfos, nil
}

// downloadPKCS12 downloads a pkcs12 format file and parses certificates and matching private keys.
func (c CertificateDownloaderImpl) downloadPKCS12(certificateURL, passphrase string) ([]certificateutil.CertificateInfoModel, error) {
	contents, err := c.downloadFile(certificateURL)
	if err != nil {
		return nil, err
	} else if contents == nil {
		return nil, fmt.Errorf("certificate (%s) is empty", certificateURL)
	}

	infos, err := c.pkcs12Converter(contents, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate (%s), err: %s", certificateURL, err)
	}

	return infos, nil
}

func (c CertificateDownloaderImpl) downloadFile(src string) ([]byte, error) {
	url, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url (%s): %s", src, err)
	}

	// Local file
	if url.Scheme == "file" {
		src := strings.Replace(src, url.Scheme+"://", "", -1)

		return c.readFile(src)
	}

	// Remote file
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	var contents []byte
	err = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
		c.debugLogger("Downloading %s, attempt %d", src, attempt)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		req = req.WithContext(ctx)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to download (%s): %s", src, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.warnLogger("failed to close (%s) body: %s", src, err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download (%s) failed with status code (%d)", src, resp.StatusCode)
		}

		contents, err = c.readAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response (%s): %s", src, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return contents, nil
}
