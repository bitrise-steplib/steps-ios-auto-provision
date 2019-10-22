package autoprovision

import "github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"

// ListDevices returns the reigestered devices on the Apple Developer portal
func ListDevices(client *appstoreconnect.Client, opt *appstoreconnect.ListDevicesOptions) ([]appstoreconnect.Device, error) {
	var filterUDID string
	if opt == nil {
		filterUDID = ""
	} else {
		filterUDID = opt.FilterUDID
	}

	var nextPageURL string
	var devices []appstoreconnect.Device
	for {
		response, err := client.Provisioning.ListDevices(&appstoreconnect.ListDevicesOptions{
			FilterUDID: filterUDID,
			Limit:      20,
			Next:       nextPageURL,
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
