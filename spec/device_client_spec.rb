require_relative '../lib/autoprovision/portal/device_client'
require_relative '../lib/autoprovision/device'
require_relative '../log/log'

RSpec.describe '.ensure_test_devices' do
  it 'it registers new device' do
    device = Device.new(
      'device_identifier' => '123456',
      'title' => 'New Device'
    )

    fake_portal_device = double
    allow(fake_portal_device).to receive(:name).and_return(device.name)
    allow(fake_portal_device).to receive(:udid).and_return(device.udid)

    fake_portal_client = double
    allow(fake_portal_client).to receive(:all).and_return(nil)
    allow(fake_portal_client).to receive(:create!).and_return(fake_portal_device)

    valid_devices = Portal::DeviceClient.ensure_test_devices([device], fake_portal_client)

    expect(valid_devices).to eq([device])
  end

  it 'supresses error due to invalid od mac device UDID' do
    device = Device.new(
      'device_identifier' => '123456',
      'title' => 'New Device'
    )

    fake_portal_device = double
    allow(fake_portal_device).to receive(:name).and_return(device.name)
    allow(fake_portal_device).to receive(:udid).and_return(device.udid)

    fake_portal_client = double
    allow(fake_portal_client).to receive(:all).and_return(nil)
    allow(fake_portal_client).to receive(:create!).and_raise('error')

    valid_devices = Portal::DeviceClient.ensure_test_devices([device], fake_portal_client)

    expect(valid_devices).to eq([])
  end
end
