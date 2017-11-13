# PortalData
class PortalData
  # TestDevice
  class TestDevice
    attr_accessor :uuid
    attr_accessor :name

    def initialize(json)
      @uuid = json['device_identifier'] || ''
      @name = json['title'] || ''
    end

    def validate
      raise 'device uuid not porvided this build' if @uuid.empty?
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
    test_devices_json.each do |device_json|
      @test_devices.push(TestDevice.new(device_json))
    end
  end

  def validate
    raise 'developer portal apple id not provided for this build' if @apple_id.empty?
    raise 'developer portal password not provided for this build' if @password.empty?
  end
end

def get_developer_portal_data(build_url, build_api_token)
  url = "#{build_url}/apple_developer_portal_data.json"
  log_debug("developer portal data url: #{url}")
  uri = URI.parse(url)

  request = Net::HTTP::Get.new(uri)
  request['BUILD_API_TOKEN'] = build_api_token

  http_object = Net::HTTP.new(uri.host, uri.port)
  http_object.use_ssl = (uri.scheme == 'https')

  response = http_object.start do |http|
    http.request(request)
  end

  # log_debug(printable_response(response))

  developer_portal_data = JSON.parse(response.body) if response.body
  error_message = developer_portal_data['error_msg'] if developer_portal_data
  error_message ||= printable_response(response)
  raise error_message unless response.code == '200'

  PortalData.new(developer_portal_data)
end
