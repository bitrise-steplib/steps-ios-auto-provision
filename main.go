package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-io/xcode-project/pretty"
	"github.com/bitrise-io/xcode-project/xcodeproj"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
	"github.com/bitrise-steplib/steps-ios-auto-provision/keychain"
)

// downloadCertificates downloads and parses a list of p12 files
func downloadCertificates(URLs []CertificateFileURL) ([]certificateutil.CertificateInfoModel, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	var certInfos []certificateutil.CertificateInfoModel

	for i, p12 := range URLs {
		log.Debugf("Downloading p12 file number %d from %s", i, p12.URL)

		p12CertInfos, err := downloadPKCS12(httpClient, p12.URL, p12.Passphrase)
		if err != nil {
			return nil, err
		}
		log.Debugf("Codesign identities included: %s", autoprovision.CertsToString(p12CertInfos))

		certInfos = append(certInfos, p12CertInfos...)
	}

	log.Debugf("%d certificates downloaded", len(certInfos))
	return certInfos, nil
}

// downloadPKCS12 downloads a pkcs12 format file and parses certificates and matching private keys.
func downloadPKCS12(httpClient *http.Client, certificateURL, passphrase string) ([]certificateutil.CertificateInfoModel, error) {
	contents, err := downloadFile(httpClient, certificateURL)
	if err != nil {
		return nil, err
	} else if contents == nil {
		return nil, fmt.Errorf("certificate (%s) is empty", certificateURL)
	}

	infos, err := certificateutil.CertificatesFromPKCS12Content(contents, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate (%s), err: %s", certificateURL, err)
	}

	return infos, nil
}

func downloadFile(httpClient *http.Client, src string) ([]byte, error) {
	url, err := url.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url (%s): %s", src, err)
	}

	// Local file
	if url.Scheme == "file" {
		src := strings.Replace(src, url.Scheme+"://", "", -1)

		return ioutil.ReadFile(src)
	}

	// Remote file
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	var contents []byte
	err = retry.Times(2).Wait(5 * time.Second).Try(func(attempt uint) error {
		log.Debugf("Downloading %s, attempt %d", src, attempt)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		req = req.WithContext(ctx)

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to download (%s): %s", src, err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Warnf("failed to close (%s) body: %s", src, err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download (%s) failed with status code (%d)", src, resp.StatusCode)
		}

		contents, err = ioutil.ReadAll(resp.Body)
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

func needToRegisterDevices(distrTypes []autoprovision.DistributionType) bool {
	for _, distrType := range distrTypes {
		if distrType == autoprovision.Development || distrType == autoprovision.AdHoc {
			return true
		}
	}
	return false
}

func failf(s string, args ...interface{}) {
	log.Errorf(s, args...)
	os.Exit(1)
}

func main() {
	var stepConf Config
	if err := stepconf.Parse(&stepConf); err != nil {
		failf(err.Error())
	}
	stepconf.Print(stepConf)

	log.SetEnableDebugLog(stepConf.VerboseLog)

	// Creating AppstoreConnectAPI client
	fmt.Println()
	log.Infof("Creating AppstoreConnectAPI client")
	privateKey, err := fileutil.ReadBytesFromFile(stepConf.PrivateKeyPth)
	if err != nil {
		failf(err.Error())
	}

	client, err := appstoreconnect.NewClient(string(stepConf.KeyID), string(stepConf.IssuerID), privateKey)
	if err != nil {
		failf(err.Error())
	}
	log.Donef("client created for: %s", client.BaseURL)

	// Analyzing project
	fmt.Println()
	log.Infof("Analyzing project")
	projHelper, config, err := autoprovision.NewProjectHelper(stepConf.ProjectPath, stepConf.Scheme, stepConf.Configuration)
	if err != nil {
		failf(err.Error())
	}
	log.Printf("configuration: %s", config)

	teamID, err := projHelper.ProjectTeamID(config)
	if err != nil {
		failf(err.Error())
	}
	log.Printf("team ID: %s", teamID)

	entitlementsByBundleID, err := projHelper.ArchivableTargetBundleIDToEntitlements()
	if err != nil {
		failf(err.Error())
	}
	log.Printf("bundle IDs: %s", pretty.Object(entitlementsByBundleID))

	platform, err := projHelper.Platform(config)
	if err != nil {
		failf(err.Error())
	}
	log.Printf("platform: %s", platform)

	// Downloading certificates
	fmt.Println()
	log.Infof("Downloading certificates")
	certURLs, err := stepConf.CertificateFileURLs()
	if err != nil {
		failf(err.Error())
	}
	certs, err := downloadCertificates(certURLs)
	if err != nil {
		failf(err.Error())
	}
	log.Printf("%d certificates downloaded:", len(certs))
	for _, cert := range certs {
		log.Printf("- %s", cert.CommonName)
	}

	certType, ok := autoprovision.CertificateTypeByDistribution[stepConf.DistributionType()]
	if !ok {
		failf("Invalid distribution type provided: %s", stepConf.DistributionType)
	}

	distrTypes := []autoprovision.DistributionType{stepConf.DistributionType()}
	requiredCertTypes := map[appstoreconnect.CertificateType]bool{certType: true}
	if stepConf.DistributionType() != autoprovision.Development {
		distrTypes = append(distrTypes, autoprovision.Development)
		requiredCertTypes[appstoreconnect.IOSDevelopment] = false
	}

	certClient := autoprovision.APIClient(client)
	certsByType, err := autoprovision.GetValidCertificates(certs, certClient, requiredCertTypes, teamID, false)
	if err != nil {
		failf(err.Error())
	}

	if len(certsByType) == 1 && stepConf.DistributionType() != autoprovision.Development {
		// remove development distribution if there is no development certificate uploaded
		distrTypes = []autoprovision.DistributionType{stepConf.DistributionType()}
	}
	log.Printf("distribution types: %s", distrTypes)

	// Ensure devices
	var devices []appstoreconnect.Device

	if needToRegisterDevices(distrTypes) {
		fmt.Println()
		log.Infof("Register %d Bitrise test devices", len(stepConf.DeviceIDs()))

		var err error
		devices, err = autoprovision.ListDevices(client, "", appstoreconnect.IOSDevice)
		if err != nil {
			failf(err.Error())
		}
		log.Printf("%d devices are already registered", len(devices))

		for _, id := range stepConf.DeviceIDs() {
			log.Printf("checking device: %s", id)

			found := false
			for _, device := range devices {
				if device.Attributes.UDID == id {
					found = true
					break
				}
			}

			if found {
				log.Printf("device already registered")
			} else {
				log.Printf("registering device")
				req := appstoreconnect.DeviceCreateRequest{
					Data: appstoreconnect.DeviceCreateRequestData{
						Attributes: appstoreconnect.DeviceCreateRequestDataAttributes{
							Name:     "Bitrise test device",
							Platform: appstoreconnect.IOS,
							UDID:     id,
						},
						Type: "devices",
					},
				}
				_, err := client.Provisioning.RegisterNewDevice(req)
				if err != nil {
					failf(err.Error())
				}
			}
		}
	}

	// Ensure Profiles
	type CodesignSettings struct {
		ProfilesByBundleID map[string]appstoreconnect.Profile
		Certificate        certificateutil.CertificateInfoModel
	}

	codesignSettingsByDistributionType := map[autoprovision.DistributionType]CodesignSettings{}

	bundleIDByBundleIDIdentifer := map[string]*appstoreconnect.BundleID{}

	for _, distrType := range distrTypes {
		fmt.Println()
		log.Infof("Checking %s provisioning profiles for %d bundle id(s)", distrType, len(entitlementsByBundleID))
		certType := autoprovision.CertificateTypeByDistribution[distrType]
		certs := certsByType[certType]
		if len(certs) == 0 {
			failf("no certificates for %s distribution", distrType)
		}

		codesignSettings := CodesignSettings{
			ProfilesByBundleID: map[string]appstoreconnect.Profile{},
			Certificate:        certs[0].Certificate,
		}

		var certIDs []string
		for _, cert := range certs {
			certIDs = append(certIDs, cert.ID)
		}

		platformProfileTypes, ok := autoprovision.PlatformToProfileTypeByDistribution[platform]
		if !ok {
			failf("unknown platform: %s, known platforms: %s, %s", platform, autoprovision.IOS, autoprovision.TVOS)
		}

		profileType := platformProfileTypes[distrType]
		log.Printf("  profile type: %s", profileType)

		var deviceIDs []string
		if needToRegisterDevices([]autoprovision.DistributionType{distrType}) {
			for _, d := range devices {
				if strings.HasPrefix(string(profileType), "TVOS") && d.Attributes.DeviceClass != "APPLE_TV" {
					continue
				} else if strings.HasPrefix(string(profileType), "IOS") &&
					string(d.Attributes.DeviceClass) != "IPHONE" && string(d.Attributes.DeviceClass) != "IPAD" && string(d.Attributes.DeviceClass) != "IPOD" {
					continue
				}
				deviceIDs = append(deviceIDs, d.ID)
			}
		}

		for bundleIDIdentifier, entitlements := range entitlementsByBundleID {
			fmt.Println()
			log.Infof("  Checking bundle id: %s", bundleIDIdentifier)
			log.Printf("  capabilities: %s", entitlements)

			// Search for Bitrise managed Profile
			profile, err := autoprovision.FindProfile(client, profileType, bundleIDIdentifier)
			if err != nil {
				failf(err.Error())
			}

			if profile != nil {
				log.Printf("  Bitrise managed profile found: %s", profile.Attributes.Name)

				// Check if Bitrise managed Profile is sync with the project
				if ok, err := autoprovision.CheckProfile(client, *profile, autoprovision.Entitlement(entitlements), deviceIDs, certIDs); err != nil {
					failf(err.Error())
				} else if ok {
					log.Donef("  profile is in sync with the project requirements")
					codesignSettings.ProfilesByBundleID[bundleIDIdentifier] = *profile
					codesignSettingsByDistributionType[distrType] = codesignSettings
					continue
				}

				// If not in sync, delete and regenerate
				log.Warnf("  profile is not in sync with the project requirements, regenerating ...")
				if err := autoprovision.DeleteProfile(client, profile.ID); err != nil {
					failf(err.Error())
				}
			} else {
				log.Warnf("  profile does not exist, generating...")
			}

			// Search for BundleID
			fmt.Println()
			log.Infof("  Searching for bundle ID: %s", bundleIDIdentifier)
			bundleID, ok := bundleIDByBundleIDIdentifer[bundleIDIdentifier]
			if !ok {
				var err error
				bundleID, err = autoprovision.FindBundleID(client, bundleIDIdentifier)
				if err != nil {
					failf(err.Error())
				}
			}

			if bundleID != nil {
				log.Printf("  bundle ID found: %s", bundleID.Attributes.Name)

				bundleIDByBundleIDIdentifer[bundleIDIdentifier] = bundleID

				// Check if BundleID is sync with the project
				if ok, err := autoprovision.CheckBundleIDEntitlements(client, *bundleID, autoprovision.Entitlement(entitlements)); err != nil {
					failf(err.Error())
				} else if !ok {
					log.Warnf("  bundle ID capabilities are not in sync with the project capabilities, synchronising...")
					if err := autoprovision.SyncBundleID(client, bundleID.ID, autoprovision.Entitlement(entitlements)); err != nil {
						failf(err.Error())
					}
				} else {
					log.Printf("  bundle ID capabilities are in sync with the project capabilities")
				}
			} else {
				// Create BundleID
				log.Warnf("  bundle ID not found, generating...")
				bundleID, err = autoprovision.CreateBundleID(client, bundleIDIdentifier, autoprovision.Entitlement(entitlements))
				if err := autoprovision.SyncBundleID(client, bundleID.ID, autoprovision.Entitlement(entitlements)); err != nil {
					failf(err.Error())
				}

				bundleIDByBundleIDIdentifer[bundleIDIdentifier] = bundleID
			}

			// Create Bitrise managed Profile
			fmt.Println()
			log.Infof("  Creating profile for bundle id: %s", bundleID.Attributes.Name)
			profile, err = autoprovision.CreateProfile(client, profileType, *bundleID, certIDs, deviceIDs)
			if err != nil {
				failf(err.Error())
			}

			log.Donef("  Profile created: %s", profile.Attributes.Name)
			codesignSettings.ProfilesByBundleID[bundleIDIdentifier] = *profile
			codesignSettingsByDistributionType[distrType] = codesignSettings
		}
	}

	// Force Codesign Settings
	log.Infof("Apply code sign setting in project")
	targets := append([]xcodeproj.Target{projHelper.MainTarget}, projHelper.MainTarget.DependentExecutableProductTargets(false)...)
	for _, target := range targets {
		codesignSettings, ok := codesignSettingsByDistributionType[autoprovision.Development]
		if !ok {
			failf("Failed to find development code sign settings")
		}
		teamID = codesignSettings.Certificate.TeamID

		targetBundleID, err := projHelper.TargetBundleID(target.Name, config)
		if err != nil {
			failf(err.Error())
		}
		profile, ok := codesignSettings.ProfilesByBundleID[targetBundleID]
		if !ok {
			failf("Failed to get profile for the bundleID %s", targetBundleID)
		}

		if err := projHelper.XcProj.ForceCodeSign(config, target.Name, teamID, codesignSettings.Certificate.CommonName, profile.Attributes.UUID); err != nil {
			failf("Failed to apply code sign settings for target (%s): %s", target.Name, err)
		}

		if err := projHelper.XcProj.Save(); err != nil {
			failf("Failed to save xcodeproj (%s): %s", projHelper.XcProj.Path, err)
		}

	}

	// Install certificates and profiles
	kc, err := keychain.New(stepConf.KeychainPath, stepConf.KeychainPassword)
	if err != nil {
		failf(err.Error())
	}

	for distrType, codesignSettings := range codesignSettingsByDistributionType {
		teamID = codesignSettings.Certificate.TeamID

		fmt.Println()
		log.Infof("Codesign settings for %s distribution", distrType)
		log.Printf("Development Team: %s(%s)", codesignSettings.Certificate.TeamName, teamID)
		log.Printf("Provisioning Profiles:")

		for bundleID, profile := range codesignSettings.ProfilesByBundleID {
			log.Printf("  %s: %s", bundleID, profile.Attributes.Name)
		}
		log.Printf("Certificate: %s", codesignSettings.Certificate.CommonName)

		if err := kc.InstallCertificate(codesignSettings.Certificate, ""); err != nil {
			failf(err.Error())
		}

		for _, profile := range codesignSettings.ProfilesByBundleID {
			if err := autoprovision.WriteProfile(profile); err != nil {
				failf(err.Error())
			}
		}
	}

	// Export output
	fmt.Println()
	log.Infof("Exporting outputs")
	outputs := map[string]string{
		"BITRISE_EXPORT_METHOD":  stepConf.Distribution,
		"BITRISE_DEVELOPER_TEAM": teamID,
	}

	settings, ok := codesignSettingsByDistributionType[autoprovision.Development]
	if ok {
		outputs["BITRISE_DEVELOPMENT_CODESIGN_IDENTITY"] = settings.Certificate.CommonName

		bundleID, err := projHelper.TargetBundleID(projHelper.MainTarget.Name, config)
		if err != nil {
			failf(err.Error())
		}
		profile, ok := settings.ProfilesByBundleID[bundleID]
		if !ok {
			failf("No provisioning profile generated for the main target: %s", projHelper.MainTarget.Name)
		}

		outputs["BITRISE_DEVELOPMENT_PROFILE"] = profile.Attributes.UUID
	}

	if stepConf.DistributionType() != autoprovision.Development {
		settings, ok := codesignSettingsByDistributionType[stepConf.DistributionType()]
		if !ok {
			failf("No codesign settings generated for the selected distribution type: %s", stepConf.DistributionType())
		}

		outputs["BITRISE_PRODUCTION_CODESIGN_IDENTITY"] = settings.Certificate.CommonName

		bundleID, err := projHelper.TargetBundleID(projHelper.MainTarget.Name, config)
		if err != nil {
			failf(err.Error())
		}
		profile, ok := settings.ProfilesByBundleID[bundleID]
		if !ok {
			failf("No provisioning profile generated for the main target: %s", projHelper.MainTarget.Name)
		}

		outputs["BITRISE_PRODUCTION_PROFILE"] = profile.Attributes.UUID
	}

	for k, v := range outputs {
		log.Donef("%s=%s", k, v)
		if err := tools.ExportEnvironmentWithEnvman(k, v); err != nil {
			failf("Failed to export %s=%s", k, v)
		}
	}

}
