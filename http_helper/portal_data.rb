# PortalData
class PortalData
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

  attr_accessor :apple_id
  attr_accessor :password
  attr_accessor :session_cookies
  attr_accessor :test_devices

  def initialize(json)
    @apple_id = json['apple_id']
    @password = json['password']
    @session_cookies = json['session_cookies']

    @test_devices = []
    test_devices_json = json['test_devices']
    test_devices_json.each { |device_json| @test_devices.push(TestDevice.new(device_json)) } unless test_devices_json.to_s.empty?
  end

  def validate
    raise 'developer portal apple id not provided for this build' if @apple_id.to_s.empty?
    raise 'developer portal password not provided for this build' if @password.to_s.empty?
    @test_devices.each(&:validate)
  end
end

def get_developer_portal_data(build_url, build_api_token)
  url = "#{build_url}/apple_developer_portal_data.json"
  Log.debug("developer portal data url: #{url}")
  Log.debug("build_api_token: #{build_api_token}")
  uri = URI.parse(url)

  request = Net::HTTP::Get.new(uri)
  request['BUILD_API_TOKEN'] = build_api_token

  http_object = Net::HTTP.new(uri.host, uri.port)
  http_object.use_ssl = true

  response = http_object.start do |http|
    http.request(request)
  end

  raise 'failed to get response' unless response
  raise 'response has no body' unless response.body

  developer_portal_data = JSON.parse(response.body)

  unless response.code == '200'
    error_message = developer_portal_data['error_msg']
    error_message ||= printable_response(response)
    raise error_message
  end

  PortalData.new(developer_portal_data)
end
