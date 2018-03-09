# AuthData
class AuthData
  attr_reader :apple_id
  attr_reader :password
  attr_reader :session_cookies
  attr_reader :test_devices

  def initialize(json_data)
    @apple_id = json_data['apple_id']
    @password = json_data['password']
    @session_cookies = json_data['session_cookies']

    @test_devices = []
    test_devices_json = json_data['test_devices']
    test_devices_json.each { |device_json| @test_devices.push(TestDevice.new(device_json)) } unless test_devices_json.to_s.empty?
  end

  def validate
    raise 'developer portal apple id not provided for this build' if @apple_id.to_s.empty?
    raise 'developer portal password not provided for this build' if @password.to_s.empty?
    @test_devices.each(&:validate)
  end
end
