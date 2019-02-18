# TestDevice
class TestDevice
  attr_reader :udid
  attr_reader :name

  def initialize(device_data)
    @udid = device_data['device_identifier'] || ''
    @name = device_data['title'] || ''
  end

  def validate
    raise 'device udid not porvided this build' if @udid.empty?
    raise 'device title not provided for this build' if @name.empty?
  end

  def to_test_device(object)
    TestDevice.New(object.udid, object.name)
  end

  def ==(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def ===(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def eql?(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def equal?(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def self.uniq(test_devices)
    return test_devices if test_devices.to_a.empty? || test_devices.to_a.length == 1

    filtered_test_devices = [test_devices.to_a[0]]
    for i in 1..test_devices.to_a.length - 1 do
      test_device = test_devices.to_a[i]

      if filtered_test_devices.include?(test_device)
        same_test_device = filtered_test_devices.detect { |device| device == test_device }
        Log.debug("Device registered multiple times on Bitrise: #{test_device.name} with udid: #{test_device.udid}.\
 Same device: #{same_test_device.name} with udid: #{same_test_device.udid}")
        next
      end
      filtered_test_devices << test_device
    end
    filtered_test_devices
  end
end
