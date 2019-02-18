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

  def ==(other)
    substituted_udid = @udid.sub(/[^0-9A-Za-z]/, '')
    other_substituted_udid = other.udid.sub(/[^0-9A-Za-z]/, '')
    substituted_udid == other_substituted_udid
  end
end
