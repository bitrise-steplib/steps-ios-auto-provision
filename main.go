package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
	"github.com/bitrise-steplib/steps-xcode-test/pretty"
)

func main() {
	// TODO: remove the whole content
	// only used the main function for quick testing the API endpoints
	// Set ISSUER, PRIVATE_KEY_ID and PRIVATE_KEY_PATH as secret for testing
	log.SetEnableDebugLog(true)

	// Authentication
	issuer := os.Getenv("ISSUER")
	kid := os.Getenv("PRIVATE_KEY_ID")
	signingKeyPth := os.Getenv("PRIVATE_KEY_PATH")

	signingKey, err := fileutil.ReadBytesFromFile(signingKeyPth)
	if err != nil {
		panic(err)
	}

	c, err := appstoreconnect.NewClient(kid, issuer, signingKey)
	if err != nil {
		panic(err)
	}

	profile, err := autoprovision.EnsureProfile(c, appstoreconnect.IOSAppDevelopment, "auto_provision.ios-simple-objc")
	if err != nil {
		panic(err)
	}
	fmt.Println(pretty.Object(profile))

	/*
		LIST CERTIFICATES

			opt := &appstoreconnect.ListCertificatesOptions{
				FilterCertificateType: appstoreconnect.MacInstallerDistribution,
				Limit:                 1,
			}
			r, err := c.Provisioning.ListCertificates(opt)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(r))
	*/

	/*
		CREATE BUNDLE ID

			body := appstoreconnect.BundleIDCreateRequest{
				Data: appstoreconnect.BundleIDCreateRequestData{
					Attributes: appstoreconnect.BundleIDCreateRequestDataAttributes{
						Identifier: "io.bitrise.appstoreconnecttest",
						Name:       "appstoreconnecttest",
						Platform:   appstoreconnect.IOS,
					},
					Type: "bundleIds",
				},
			}
			r, err := c.Provisioning.CreateBundleID(body)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(r))
	*/

	/*
		LIST BUNDLE IDS

			opt := &appstoreconnect.ListBundleIDsOptions{
				Include: "bundleIdCapabilities",
				Limit:   1,
			}
			r, err := c.Provisioning.ListBundleIDs(opt)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(r))
	*/

	/*
		CREATE DEVICE

			body := appstoreconnect.DeviceCreateRequest{
				Data: appstoreconnect.DeviceCreateRequestData{
					Type: "devices",
					Attributes: appstoreconnect.DeviceCreateRequestDataAttributes{
						Name:     "appstoreconnect test",
						UDID:     "invalid",
						Platform: appstoreconnect.IOS,
					},
				},
			}

			r, err := c.Provisioning.RegisterNewDevice(body)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(r))
	*/

	/*
		LIST DEVICES

			opt := &appstoreconnect.ListDevicesOptions{
				Limit: 1,
			}

			d, err := c.Provisioning.ListDevices(opt)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(d))
	*/

	/*
		CREATE PROFILE

			body := appstoreconnect.ProfileCreateRequest{
				Data: appstoreconnect.ProfileCreateRequestData{
					Type: "profiles",

					Attributes: appstoreconnect.ProfileCreateRequestDataAttributes{
						Name:        "AppstoreConnect API test",
						ProfileType: appstoreconnect.IOSAppDevelopment,
					},

					Relationships: appstoreconnect.ProfileCreateRequestDataRelationships{
						BundleID: appstoreconnect.ProfileCreateRequestDataRelationshipsBundleID{
							Data: appstoreconnect.ProfileCreateRequestDataRelationshipData{
								ID:   "",
								Type: "",
							},
						},

						Certificates: appstoreconnect.ProfileCreateRequestDataRelationshipsCertificates{
							Data: []appstoreconnect.ProfileCreateRequestDataRelationshipData{
								{
									ID:   "",
									Type: "",
								},
							},
						},

						Devices: appstoreconnect.ProfileCreateRequestDataRelationshipsDevices{
							Data: []appstoreconnect.ProfileCreateRequestDataRelationshipData{
								{
									ID:   "",
									Type: "",
								},
							},
						},
					},
				},
			}
			c.Provisioning.CreateProfile(body)
	*/

	/*
		LIST PROFILES

			opt := &appstoreconnect.ListProfilesOptions{
				Include:            "bundleId,certificates,devices",
				FilterProfileState: appstoreconnect.Active,
				FilterProfileType:  appstoreconnect.IOSAppDevelopment,
				Limit:              1,
			}

			r, err := c.Provisioning.ListProfiles(opt)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(r))

			req, err := c.NewRequest(http.MethodGet, r.Data[0].Relationships.BundleID.Links.Related, nil)
			if err != nil {
				panic(err)
			}
			var bundleID map[string]interface{}
			_, err = c.Do(req, &bundleID)
			if err != nil {
				panic(err)
			}
			fmt.Println(pretty.Object(bundleID))
	*/

	/*
		WRITE PROFILE IN FILE

			p := r.Data[0].Attributes.ProfileContent

			decoded, err := base64.StdEncoding.DecodeString(p)
			if err != nil {
				fmt.Println("decode error:", err)
				return
			}
			fmt.Println(string(decoded))
	*/
}
