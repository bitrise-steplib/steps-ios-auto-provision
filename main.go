package main

import (
	"fmt"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/xcode-project/pretty"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
	"os"
)

func failf(s string, args ...interface{}) {
	log.Errorf(s, args...)
	os.Exit(1)
}

func main() {
	stepConf, err := autoprovision.ParseConfig()
	if err != nil {
		log.Warnf(err.Error())
	} else {
		stepconf.Print(stepConf)
	}

	log.SetEnableDebugLog(stepConf.VerboseLog == "yes")

	//
	fmt.Println()
	log.Infof("Creating AppstoreConnectAPI client")
	privateKey, err := fileutil.ReadBytesFromFile(stepConf.PrivateKeyPth)
	if err != nil {
		failf(err.Error())
	}

	client, err := appstoreconnect.NewClient(stepConf.KeyID, stepConf.IssuerID, privateKey)
	if err != nil {
		failf(err.Error())
	}
	log.Donef("client created for: %s", client.BaseURL)

	//
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

	platform := projHelper.Platform
	log.Printf("platform: %s", platform)

	//
	fmt.Println()
	log.Infof("Downloading certificates")
	certURLs, err := stepConf.CertificateFileURLs()
	if err != nil {
		failf(err.Error())
	}
	certs, err := autoprovision.DownloadLocalCertificates(certURLs)
	if err != nil {
		failf(err.Error())
	}
	log.Printf("%d certificates downloaded:", len(certs))
	for _, cert := range certs {
		log.Printf("- %s", cert.CommonName)
	}

	certType, ok := autoprovision.CertificateTypeByDistribution[stepConf.DistributionType]
	if !ok {
		failf("Invalid distribution type provided: %s", stepConf.DistributionType)
	}

	distrTypes := []autoprovision.DistributionType{stepConf.DistributionType}
	requiredCertTypes := map[appstoreconnect.CertificateType]bool{certType: true}
	certTypesByName := map[appstoreconnect.CertificateType]string{certType: ""}
	if stepConf.DistributionType != autoprovision.Development {
		distrTypes = append(distrTypes, autoprovision.Development)
		requiredCertTypes[appstoreconnect.IOSDevelopment] = false
		certTypesByName[appstoreconnect.IOSDevelopment] = ""
	}
	log.Printf("distribution types: %s", distrTypes)

	certClient := autoprovision.APIClient(client)
	certsByType, err := autoprovision.GetValidCertificates(certs, certClient, requiredCertTypes, certTypesByName, teamID, false)
	if err != nil {
		failf(err.Error())
	}

	if len(certsByType) == 1 && stepConf.DistributionType != autoprovision.Development {
		// remove development distribution if there is no development certificate uploaded
		distrTypes = []autoprovision.DistributionType{stepConf.DistributionType}
	}

	// Ensure devices
	fmt.Println()
	log.Infof("Register %d Bitrise test devices", len(stepConf.DeviceIDs()))
	var devices []appstoreconnect.Device
	for _, id := range stepConf.DeviceIDs() {
		log.Printf("checking device: %s", id)
		r, err := client.Provisioning.ListDevices(&appstoreconnect.ListDevicesOptions{
			FilterUDID: id,
		})
		if err != nil {
			failf(err.Error())
		}
		if len(r.Data) > 0 {
			log.Printf("device already registered", id)
			devices = append(devices, r.Data[0])
		} else {
			log.Printf("registering device", id)
			req := appstoreconnect.DeviceCreateRequest{
				Data: appstoreconnect.DeviceCreateRequestData{
					Attributes: appstoreconnect.DeviceCreateRequestDataAttributes{
						Name:     "Bitrise test device",
						Platform: appstoreconnect.IOS,
						UDID:     id,
					},
					Type: "",
				},
			}
			r, err := client.Provisioning.RegisterNewDevice(req)
			if err != nil {
				failf(err.Error())
			}
			devices = append(devices, r.Data...)
		}
	}

	r, err := client.Provisioning.ListDevices(nil)
	if err != nil {
		failf(err.Error())
	}
	log.Printf("%d devices are registered", len(r.Data))

	var deviceIDs []string
	for _, device := range r.Data {
		deviceIDs = append(deviceIDs, device.ID)
	}

	fmt.Println()
	log.Infof("Checking provisioning profiles for %d bundle id(s)", len(entitlementsByBundleID))
	profilesByBundleID := map[string][]autoprovision.Profile{}
	for bundleIDIdentifier, entitlements := range entitlementsByBundleID {
		fmt.Println()
		log.Infof("  Checking bundle id: %s", bundleIDIdentifier)
		log.Printf("  capabilities: %s", entitlements)

		for _, distrType := range distrTypes {
			log.Printf("  distribution type: %s", distrType)

			// Search for Bitrise managed Profile
			platformProfileTypes, ok := autoprovision.PlatformToProfileTypeByDistribution[platform]
			if !ok {
				failf("unknown platform: %s", platform)
			}
			profileType := platformProfileTypes[distrType]
			log.Printf("  profile type: %s", profileType)

			profile, err := autoprovision.FindProfile(client, profileType, bundleIDIdentifier)
			if err != nil {
				failf(err.Error())
			}

			if profile != nil {
				log.Printf("  Bitrise managed profile found: %s", profile.Attributes.Name)

				// Check if Bitrise managed Profile is sync with the project
				if ok, err := autoprovision.CheckProfile(*profile, autoprovision.Entitlement(entitlements), nil, nil); err != nil {
					failf(err.Error())
				} else if ok {
					log.Donef("  profile capabilities are in sync with the project capabilities")
					profiles := profilesByBundleID[bundleIDIdentifier]
					profiles = append(profiles, *profile)
					profilesByBundleID[bundleIDIdentifier] = profiles
					continue
				}

				// If not in sync, delete and re generate
				log.Warnf("  profile capabilities are not in sync with the project capabilities, re generating ...")
				if err := autoprovision.DeleteProfile(client, *profile); err != nil {
					failf(err.Error())
				}
			} else {
				log.Warnf("  profile does not exist, generating...")
			}

			// Search for BundleID
			fmt.Println()
			log.Infof("  Searching for bundle ID: %s", bundleIDIdentifier)
			bundleID, err := autoprovision.FindBundleID(client, bundleIDIdentifier)
			if err != nil {
				failf(err.Error())
			}

			if bundleID != nil {
				log.Printf("  bundle ID found: %s", bundleID.Attributes.Name)
				// Check if BundleID is sync with the project
				if ok, err := autoprovision.CheckBundleID(*bundleID, autoprovision.Entitlement(entitlements)); err != nil {
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
			}

			// Create Bitrise managed Profile
			fmt.Println()
			log.Infof("  Creating profile for bundle id: %s", bundleID.Attributes.Name)
			certType := autoprovision.CertificateTypeByDistribution[distrType]
			certs := certsByType[certType]
			var certIDs []string
			for _, cert := range certs {
				certIDs = append(certIDs, cert.ID)
			}

			profile, err = autoprovision.CreateProfile(client, profileType, *bundleID, certIDs, deviceIDs)
			if err != nil {
				failf(err.Error())
			}
			log.Donef("  Profile created: %s", profile.Attributes.Name)

			profiles := profilesByBundleID[bundleIDIdentifier]
			profiles = append(profiles, *profile)
			profilesByBundleID[bundleIDIdentifier] = profiles
		}
	}
}
