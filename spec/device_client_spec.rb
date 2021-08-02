require_relative '../lib/autoprovision/portal/device_client'
require_relative '../lib/autoprovision/device'
require_relative '../log/log'

RSpec.describe '.ensure_test_devices' do
  it 'returns empty array for empty input' do
    fake_portal_client = double
    allow(fake_portal_client).to receive(:all).and_return(nil)

    dev_portal_devices = Portal::DeviceClient.ensure_test_devices(true, [], :ios, fake_portal_client)

    expect(dev_portal_devices).to eq([])
  end

  it 'it registers new device' do
    device = Device.new(
      'device_identifier' => '123456',
      'title' => 'New Device'
    )

    fake_portal_device = double
    allow(fake_portal_device).to receive(:name).and_return(device.name)
    allow(fake_portal_device).to receive(:udid).and_return(device.udid)
    allow(fake_portal_device).to receive(:device_type).and_return('iphone')

    fake_portal_client = double
    allow(fake_portal_client).to receive(:all).and_return(nil)
    allow(fake_portal_client).to receive(:create!).and_return(fake_portal_device)

    dev_portal_devices = Portal::DeviceClient.ensure_test_devices(true, [device], :ios, fake_portal_client)
    dev_portal_device_udids = dev_portal_devices.map(&:udid)
    test_device_udids = [device].map(&:udid)

    expect(dev_portal_device_udids).to eq(test_device_udids)
  end

  it 'suppresses error due to invalid or mac device UDID' do
    existing_device = Device.new(
      'device_identifier' => '123456',
      'title' => 'Existing Device'
    )
    invalid_device = Device.new(
      'device_identifier' => 'invalid-udid',
      'title' => 'Invalid Device'
    )

    fake_portal_device = double
    allow(fake_portal_device).to receive(:name).and_return(existing_device.name)
    allow(fake_portal_device).to receive(:udid).and_return(existing_device.udid)
    allow(fake_portal_device).to receive(:device_type).and_return('iphone')

    fake_portal_client = double
    allow(fake_portal_client).to receive(:all).and_return([fake_portal_device])
    allow(fake_portal_client).to receive(:create!).and_raise('error')

    dev_portal_devices = Portal::DeviceClient.ensure_test_devices(true, [existing_device, invalid_device], :ios, fake_portal_client)
    dev_portal_device_udids = dev_portal_devices.map(&:udid)
    test_device_udids = [existing_device].map(&:udid)

    expect(dev_portal_device_udids).to eq(test_device_udids)
  end

  [
    [:ios, 'watch', 1],
    [:ios, 'ipad', 1],
    [:ios, 'iphone', 1],
    [:ios, 'ipod', 1],
    [:ios, 'tvOS', 0],

    [:osx, 'watch', 0],
    [:osx, 'ipad', 0],
    [:osx, 'iphone', 0],
    [:osx, 'ipod', 0],
    [:osx, 'tvOS', 0],

    [:tvos, 'watch', 0],
    [:tvos, 'ipad', 0],
    [:tvos, 'iphone', 0],
    [:tvos, 'ipod', 0],
    [:tvos, 'tvOS', 1],

    [:watchos, 'watch', 1],
    [:watchos, 'ipad', 1],
    [:watchos, 'iphone', 1],
    [:watchos, 'ipod', 1],
    [:watchos, 'tvOS', 0],
  ].each do |platform, device_type, len|
    it "on #{platform} platform with #{device_type} device valid devices length should be #{len}" do
      device = Device.new(
        'device_identifier' => '123456',
        'title' => 'New Device'
      )

      fake_portal_device = double
      allow(fake_portal_device).to receive(:name).and_return(device.name)
      allow(fake_portal_device).to receive(:udid).and_return(device.udid)
      allow(fake_portal_device).to receive(:device_type).and_return(device_type)
  
      fake_portal_client = double
      allow(fake_portal_client).to receive(:all).and_return(nil)
      allow(fake_portal_client).to receive(:create!).and_return(fake_portal_device)

      dev_portal_devices = Portal::DeviceClient.ensure_test_devices(true, [device], platform, fake_portal_client)

      expect(dev_portal_devices.length).to eq(len)
    end
  end
end
