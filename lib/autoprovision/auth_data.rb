require_relative 'test_device'

# AuthData
class AuthData
  attr_reader :apple_id
  attr_reader :password
  attr_reader :session_cookies
  attr_reader :test_devices

  def initialize(auth_data)
    @apple_id = auth_data['apple_id']
    @password = auth_data['password']
    @session_cookies = auth_data['session_cookies']

    @test_devices = []
    test_devices_json = auth_data['test_devices']
    test_devices_json.each { |device_data| @test_devices.push(TestDevice.new(device_data)) } unless test_devices_json.to_s.empty?
  end

  def validate
    raise 'developer portal apple id not provided for this build' if @apple_id.to_s.empty?
    raise 'developer portal password not provided for this build' if @password.to_s.empty?
    @test_devices.each(&:validate)
  end
end
