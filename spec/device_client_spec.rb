require_relative '../lib/autoprovision/portal/device_client'
require_relative '../lib/autoprovision/test_device'
require_relative '../log/log'

RSpec.describe '.ensure_test_devices' do
  it 'it registers new device' do
    device = TestDevice.new(
      'device_identifier' => '123456',
      'title' => 'New Device'
    )

    fake_portal_device = double
    allow(fake_portal_device).to receive(:name).and_return(device.name)
    allow(fake_portal_device).to receive(:udid).and_return(device.udid)

    fake_portal_client = double
    allow(fake_portal_client).to receive(:all).and_return(nil)
    allow(fake_portal_client).to receive(:create!).and_return(fake_portal_device)

    Portal::DeviceClient.ensure_test_devices([device], fake_portal_client)
  end
end
