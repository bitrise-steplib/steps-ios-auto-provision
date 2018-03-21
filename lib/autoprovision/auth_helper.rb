require 'json'

require_relative 'common'
require_relative 'auth_data'
require_relative 'portal/auth_client'

# AuthHelper ...
class AuthHelper
  DES_COOKIE_TEMPLATE = '---
  - !ruby/object:HTTP::Cookie
    name: <DES_NAME>
    value: <DES_VALUE>
    domain: idmsa.apple.com
    for_domain: true
    path: "/"
    secure: true
    httponly: true
    expires:
    max_age: 2592000
  '.freeze

  attr_reader :test_devices

  def login(build_url, build_api_token, team_id)
    portal_data = get_developer_portal_data(build_url, build_api_token)
    portal_data.validate

    @test_devices = portal_data.test_devices

    two_factor_session = convert_des_cookie(portal_data.session_cookies)
    Portal::AuthClient.login(portal_data.apple_id, portal_data.password, two_factor_session, team_id)
  end

  private

  def get_developer_portal_data(build_url, build_api_token)
    portal_data_json = ENV['BITRISE_PORTAL_DATA_JSON']
    unless portal_data_json.nil?
      developer_portal_data = JSON.parse(portal_data_json)
      return AuthData.new(developer_portal_data)
    end

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

    AuthData.new(developer_portal_data)
  end

  def convert_des_cookie(cookies_json_str)
    Log.debug("session cookie: #{cookies_json_str}")

    cookies_json_str.each_value do |cookies|
      cookies.each do |cookie|
        name = cookie['name']
        value = cookie['value']

        next unless name.start_with?('DES')

        converted_cookie = DES_COOKIE_TEMPLATE.sub('<DES_NAME>', name).sub('<DES_VALUE>', value).gsub!("\n", '\n')
        Log.debug("converted session cookie: #{converted_cookie}")

        return converted_cookie
      end
    end
    nil
  end
end
