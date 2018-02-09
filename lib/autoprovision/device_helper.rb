require_relative 'portal/device'

# TestDevice
class TestDevice
  attr_accessor :udid
  attr_accessor :name

  def initialize(json)
    @udid = json['device_identifier'] || ''
    @name = json['title'] || ''
  end

  def validate
    raise 'device udid not porvided this build' if @udid.empty?
    raise 'device title not provided for this build' if @name.empty?
  end
end

# DeviceHelper ...
class DeviceHelper
  def ensure_test_devices(test_devices)
    Portal::DeviceHelper.ensure_test_devices(test_devices)
  end
end
