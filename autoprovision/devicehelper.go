package autoprovision

import "github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"

// ListDevices returns the reigestered devices on the Apple Developer portal
func ListDevices(client *appstoreconnect.Client, udid string, platform appstoreconnect.DevicePlatform) ([]appstoreconnect.Device, error) {
	var nextPageURL string
	var devices []appstoreconnect.Device
	for {
		response, err := client.Provisioning.ListDevices(&appstoreconnect.ListDevicesOptions{
			FilterUDID:     udid,
			FilterPlatform: platform,
			Limit:          20,
			Next:           nextPageURL,
		})
		if err != nil {
			return nil, err
		}

		devices = append(devices, response.Data...)

		nextPageURL = response.Links.Next
		if nextPageURL == "" {
			return devices, nil
		}
	}
}
