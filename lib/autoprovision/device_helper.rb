require_relative 'portal/device'

# DeviceHelper ...
class DeviceHelper
  def ensure_test_devices(test_devices)
    Portal::DeviceHelper.ensure_test_devices(test_devices)
  end
end
