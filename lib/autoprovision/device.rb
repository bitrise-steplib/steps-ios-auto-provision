# Device
class Device
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

  def eql?(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def ===(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def ==(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end

  def self.filter_duplicated_devices(devices)
    return devices if devices.to_a.empty?
    devices.uniq { |device| device.udid.sub(/[^0-9A-Za-z]/, '') }
  end

  def self.duplicated_device_groups(devices)
    return devices if devices.to_a.empty?
    groups = devices.group_by { |device| device.udid.sub(/[^0-9A-Za-z]/, '') }.values.select { |a| a.length > 1 }
    groups
  end
end
