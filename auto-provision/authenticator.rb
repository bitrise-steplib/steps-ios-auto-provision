require 'fastlane'

def convert_tfa_cookies(tfa_cookies)
  session_cookies_str = '---
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
'

  des_name = ''
  des_value = ''

  tfa_cookies.each_value do |cookies|
    cookies.each do |cookie|
      name = cookie['name']
      value = cookie['value']

      if name.start_with? 'DES'
        des_name = name
        des_value = value
      end
    end
  end

  raise 'invalid cookies: missing DES... name' if des_name.empty?
  raise 'invalid cookies: missing DES... value' if des_value.empty?

  session_cookies_str.sub('<DES_NAME>', des_name).sub('<DES_VALUE>', des_value).gsub!("\n", '\n')
end

def developer_portal_authentication(username, password, two_factor_session = nil, team_id = nil)
  ENV['FASTLANE_PASSWORD'] = password
  ENV['FASTLANE_SESSION'] = two_factor_session unless two_factor_session.to_s.empty?

  client = Spaceship::Portal.login(username, password)

  if team_id.to_s.empty?
    teams = client.teams
    raise 'Your developer portal account belongs to multiple teams, please provide the team id to sign in' if teams.to_a.size > 1
  else
    client.team_id = team_id
  end
end
