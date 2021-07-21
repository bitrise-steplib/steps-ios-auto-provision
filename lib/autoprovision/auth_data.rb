require_relative 'device'

# AuthData
class AuthData
  attr_reader :apple_id, :password, :session_cookies, :test_devices

  def initialize(auth_data)
    @apple_id = auth_data['apple_id']
    @password = auth_data['password']
    @session_cookies = auth_data['session_cookies']

    @test_devices = []
    test_devices_json = auth_data['test_devices']
    unless test_devices_json.to_s.empty?
      test_devices_json.each do |device_data|
        @test_devices.push(Device.new(device_data))
      end
    end
  end

  def validate
    raise 'developer portal apple id not provided for this build' if @apple_id.to_s.empty?
    raise 'developer portal password not provided for this build' if @password.to_s.empty?

    @test_devices.each(&:validate)
  end
end
