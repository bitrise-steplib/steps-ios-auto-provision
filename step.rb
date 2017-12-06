require_relative 'log/log'
require_relative 'project_helper/project_helper'
require_relative 'http_helper/http_helper'
require_relative 'http_helper/portal_data'
require_relative 'auto-provision/authenticator'
require_relative 'auto-provision/generator'
require_relative 'auto-provision/app_services'
require_relative 'keychain/keychain'
require_relative 'certificate_helper/certificate_helper'

# CertificateInfo
class CertificateInfo
  attr_accessor :path
  attr_accessor :passphrase
  attr_accessor :certificate
  attr_accessor :portal_certificate

  def initialize(path, passphrase, certificate)
    @path = path
    @passphrase = passphrase
    @certificate = certificate
  end
end

# ProfileInfo
class ProfileInfo
  attr_accessor :path
  attr_accessor :portal_profile

  def initialize(path, portal_profile)
    @path = path
    @portal_profile = portal_profile
  end
end

# CodesignSettings
class CodesignSettings
  attr_accessor :team_id
  attr_accessor :development_certificate_info
  attr_accessor :production_certificate_info
  attr_accessor :bundle_id_development_profile
  attr_accessor :bundle_id_production_profile
end

# Params
class Params
  attr_accessor :build_url
  attr_accessor :build_api_token
  attr_accessor :team_id
  attr_accessor :certificate_urls_str
  attr_accessor :passphrases_str
  attr_accessor :distribution_type
  attr_accessor :project_path
  attr_accessor :scheme
  attr_accessor :configuration
  attr_accessor :keychain_path
  attr_accessor :keychain_password
  attr_accessor :verbose_log

  attr_accessor :certificate_urls
  attr_accessor :passphrases

  def initialize
    @build_url = ENV['build_url']
    @build_api_token = ENV['build_api_token']
    @team_id = ENV['team_id']
    @certificate_urls_str = ENV['certificate_urls']
    @passphrases_str = ENV['passphrases']
    @distribution_type = ENV['distribution_type'] || ENV['distributon_type']
    @project_path = ENV['project_path']
    @scheme = ENV['scheme']
    @configuration = ENV['configuration']
    @keychain_path = ENV['keychain_path']
    @keychain_password = ENV['keychain_password']
    @verbose_log = ENV['verbose_log']

    @certificate_urls = split_pipe_separated_list(@certificate_urls_str)
    @passphrases = split_pipe_separated_list(@passphrases_str)
  end

  def print
    Log.info('Params:')
    Log.print("team_id: #{@team_id}")
    Log.print("certificate_urls: #{Log.secure_value(@certificate_urls_str)}")
    Log.print("passphrases: #{Log.secure_value(@passphrases_str)}")
    Log.print("distribution_type: #{@distribution_type}")
    Log.print("project_path: #{@project_path}")
    Log.print("scheme: #{@scheme}")
    Log.print("configuration: #{@configuration}")
    Log.print("build_url: #{@build_url}")
    Log.print("build_api_token: #{Log.secure_value(@build_api_token)}")
    Log.print("keychain_path: #{@keychain_path}")
    Log.print("keychain_password: #{Log.secure_value(@keychain_password)}")
    Log.print("verbose_log: #{@verbose_log}")

    Log.error("\n'distributon_type' input is deprecated please use 'distribution_type'") if ENV['distributon_type']
  end

  def validate
    raise 'missing: build_url' if @build_url.nil?
    raise 'missing: build_api_token' if @build_api_token.nil?
    raise 'missing: certificate_urls' if @certificate_urls_str.nil?
    raise 'missing: distribution_type' if @distribution_type.nil?
    raise 'missing: project_path' if @project_path.nil?
    raise 'missing: scheme' if @scheme.nil?
    raise 'missing: keychain_path' if @keychain_path.nil?
    raise 'missing: keychain_password' if @keychain_password.nil?
    raise 'missing: verbose_log' if @verbose_log.nil?
  end

  private

  def split_pipe_separated_list(list)
    return [''] if list.to_s.empty?
    list.split('|', -1)
  end
end

begin
  # Params
  params = Params.new
  params.print
  params.validate

  Log.verbose = (params.verbose_log == 'yes')
  ###

  # Developer Portal authentication
  Log.info('Developer Portal authentication')

  portal_data = get_developer_portal_data(params.build_url, params.build_api_token)
  portal_data.validate

  session = nil
  unless portal_data.session_cookies.to_s.empty?
    Log.debug("session cookie: #{portal_data.session_cookies}\n")
    session = convert_tfa_cookies(portal_data.session_cookies)
    Log.debug("converted session cookie: #{session}\n")
  end

  developer_portal_authentication(portal_data.apple_id, portal_data.password, session, params.team_id)

  Log.success('authenticated')
  ###

  # Download certificates
  Log.info('Downloading Certificates')

  certificate_urls = params.certificate_urls.reject(&:empty?)
  raise 'no certificates provider' if certificate_urls.to_a.empty?

  passphrases = params.passphrases
  raise "certificates count (#{certificate_urls.length}) and passphrases count (#{passphrases.length}) should match" unless certificate_urls.length == passphrases.length

  certificate_infos = []
  certificate_urls.each_with_index do |url, idx|
    Log.debug("downloading certificate ##{idx + 1}")
    path = download_or_create_local_path(url, "Certrificate#{idx}.p12")
    Log.debug("certificate path: #{path}")

    certificates = read_certificates(path, passphrases[idx])
    certificates.each do |certificate|
      certificate_info = CertificateInfo.new(path, passphrases[idx], certificate)
      certificate_infos = append_if_latest_certificate(certificate_info, certificate_infos)
    end
  end
  Log.success("#{certificate_urls.length} certificate files downloaded, #{certificate_infos.length} codesign identities included")
  ###

  # Identify Certificates on developer Portal
  Log.info('Identify Certificates on developer Portal')

  development_certificate_infos = []
  production_certificate_infos = []
  certificate_infos.each do |certificate_info|
    Log.debug("searching for Certificate (#{certificate_common_name(certificate_info.certificate)})")

    portal_certificate = find_development_portal_certificate(certificate_info.certificate)
    if portal_certificate
      Log.success("development Certificate identified: #{portal_certificate.name}")
      certificate_info.portal_certificate = portal_certificate
      development_certificate_infos.push(certificate_info)
    end

    portal_certificate = find_production_portal_certificate(certificate_info.certificate)
    next unless portal_certificate

    Log.success("production Certificate identified: #{portal_certificate.name}")
    certificate_info.portal_certificate = portal_certificate
    production_certificate_infos.push(certificate_info)
  end
  raise 'no development nor production certificate identified on development portal' if development_certificate_infos.empty? && production_certificate_infos.empty?
  ###

  # Anlyzing project
  Log.info('Analyzing project')

  project_helper = ProjectHelper.new(params.project_path, params.scheme, params.configuration)

  codesign_identity = project_helper.project_codesign_identity
  Log.print("project codesign identity: #{codesign_identity}")

  team_id = project_helper.project_team_id
  Log.print("project team id: #{team_id}")

  targets = project_helper.targets.collect(&:name)
  targets.each_with_index do |target_name, idx|
    bundle_id = project_helper.target_bundle_id(target_name)
    entitlements = project_helper.target_entitlements(target_name) || {}

    Log.print("target ##{idx}: #{target_name} (#{bundle_id}) with #{entitlements.length} services")
  end

  if !params.team_id.to_s.empty? && params.team_id != team_id
    Log.warn("different team id defined: #{params.team_id} than the project's one: #{team_id}")
    Log.warn("using defined team id: #{params.team_id}")
    Log.warn("dropping project codesign identity: #{codesign_identity}")

    team_id = params.team_id
    codesign_identity = nil
  end
  ###

  # Matching project codesign identity with the uploaded certificates
  Log.info('Matching project codesign identity with the uploaded certificates')

  team_development_certificate_infos = map_certificates_infos_by_team_id(development_certificate_infos)[team_id] || []
  team_production_certificate_infos = map_certificates_infos_by_team_id(production_certificate_infos)[team_id] || []

  if team_development_certificate_infos.empty? && team_production_certificate_infos.empty?
    raise "no certificate uploaded for the desired team: #{team_id}"
  end

  codesign_settings = CodesignSettings.new
  codesign_settings.team_id = team_id

  if codesign_identity
    filtered_team_development_certificate_infos = team_development_certificate_infos.select do |certificate_info|
      common_name = certificate_common_name(certificate_info.certificate)
      common_name.downcase.include?(codesign_identity.downcase)
    end
    team_development_certificate_infos = filtered_team_development_certificate_infos unless filtered_team_development_certificate_infos.empty?

    filtered_team_production_certificate_infos = team_production_certificate_infos.select do |certificate_info|
      common_name = certificate_common_name(certificate_info.certificate)
      common_name.downcase.include?(codesign_identity.downcase)
    end
    team_production_certificate_infos = filtered_team_production_certificate_infos unless filtered_team_production_certificate_infos.empty?
  end

  if team_development_certificate_infos.length > 1
    msg = "Multiple Development certificates mathes to development team: #{team_id}"
    msg += " and name: #{codesign_identity}" if codesign_identity
    Log.warn(msg)
    team_development_certificate_infos.each { |info| Log.warn(" - #{certificate_common_name(info.certificate)}") }
  end

  unless team_development_certificate_infos.empty?
    certificate_info = team_development_certificate_infos[0]
    Log.success("using: #{certificate_common_name(certificate_info.certificate)}")
    codesign_settings.development_certificate_info = certificate_info
  end

  if team_production_certificate_infos.length > 1
    msg = "Multiple Distribution certificates mathes to development team: #{team_id}"
    msg += " and name: #{codesign_identity}" if codesign_identity
    Log.warn(msg)
    team_production_certificate_infos.each { |info| Log.warn(" - #{certificate_common_name(info.certificate)}") }
  end

  unless team_production_certificate_infos.empty?
    certificate_info = team_production_certificate_infos[0]
    Log.success("using: #{certificate_common_name(certificate_info.certificate)}")
    codesign_settings.production_certificate_info = certificate_info
  end

  if params.distribution_type == 'development' && codesign_settings.development_certificate_info.nil?
    raise 'Selected distribution type: development, but forgot to provide a Development type certificate.' \
"Don't worry, it's really simple to fix! :)" \
"Simply provide a Development type certificate (.p12) and we'll be building in no time!"
  end

  if params.distribution_type != 'development' && codesign_settings.production_certificate_info.nil?
    raise "Selected distribution type: #{params.distribution_type}, but forgot to provide a Distribution type certificate." \
"Don't worry, it's really simple to fix! :)" \
"Simply provide a Distribution type certificate (.p12) and we'll be building in no time!"
  end
  ###

  # Ensure test devices
  if ['development', 'ad-hoc'].include?(params.distribution_type)
    Log.info('Ensure test devices on Developer Portal')
    ensure_test_devices(portal_data.test_devices)
  end
  ###

  # Ensure App IDs and Provisioning Profiles on Developer Portal
  Log.info('Ensure App IDs and Provisioning Profiles on Developer Portal')

  bundle_id_development_profile = {}
  bundle_id_production_profile = {}

  development_certificate_info = codesign_settings.development_certificate_info
  production_certificate_info = codesign_settings.production_certificate_info

  targets.each do |target_name|
    bundle_id = project_helper.target_bundle_id(target_name)
    entitlements = project_helper.target_entitlements(target_name) || {}

    puts
    Log.success("checking target: #{target_name} (#{bundle_id}) with #{entitlements.length} services")

    Log.print("ensure App ID (#{bundle_id}) on Developer Portal")
    app = ensure_app(bundle_id)

    Log.print("sync App ID (#{bundle_id}) Services")
    app = sync_app_services(app, entitlements)

    if codesign_settings.development_certificate_info
      Log.print('ensure Development Provisioning Profile on Developer Portal')
      portal_profile = ensure_provisioning_profile(development_certificate_info.portal_certificate, app, 'development')

      Log.success("downloading development profile: #{portal_profile.name}")
      profile_path = write_profile(portal_profile)

      Log.debug("profile path: #{profile_path}")

      profile_info = ProfileInfo.new(profile_path, portal_profile)
      bundle_id_development_profile[bundle_id] = profile_info
    end

    next if params.distribution_type == 'development'
    next unless production_certificate_info

    Log.print('ensure Production Provisioning Profile on Developer Portal')
    portal_profile = ensure_provisioning_profile(production_certificate_info.portal_certificate, app, params.distribution_type)

    Log.success("downloading #{params.distribution_type} profile: #{portal_profile.name}")
    profile_path = write_profile(portal_profile)

    Log.debug("profile path: #{profile_path}")

    profile_info = ProfileInfo.new(profile_path, portal_profile)
    bundle_id_production_profile[bundle_id] = profile_info

    codesign_settings.bundle_id_development_profile = bundle_id_development_profile
    codesign_settings.bundle_id_production_profile = bundle_id_production_profile
  end
  ###

  # Apply code sign setting in project
  Log.info('Apply code sign setting in project')

  targets.each do |target_name|
    bundle_id = project_helper.target_bundle_id(target_name)

    puts
    Log.success("configure target: #{target_name} (#{bundle_id})")

    team_id = codesign_settings.team_id
    code_sign_identity = nil
    provisioning_profile = nil

    if codesign_settings.development_certificate_info
      code_sign_identity = certificate_common_name(codesign_settings.development_certificate_info.certificate)
      provisioning_profile = codesign_settings.bundle_id_development_profile[bundle_id].portal_profile.uuid
    elsif codesign_settings.production_certificate_info
      code_sign_identity = certificate_common_name(codesign_settings.production_certificate_info.certificate)
      provisioning_profile = codesign_settings.bundle_id_production_profile[bundle_id].portal_profile.uuid
    else
      raise "no codesign settings generated for target: #{target_name} (#{bundle_id})"
    end

    project_helper.force_code_sign_properties(target_name, team_id, code_sign_identity, provisioning_profile)
  end
  ###

  # Install certificates
  Log.info('Install certificates')

  keychain_helper = KeychainHelper.new(params.keychain_path, params.keychain_password)

  certificate_path_passphrase_map = Hash[certificate_infos.map { |info| [info.path, info.passphrase] }]

  keychain_helper.install_certificates(certificate_path_passphrase_map)

  Log.success("#{certificate_path_passphrase_map.length} certificates installed")
  ###
rescue => ex
  puts
  Log.error(ex.to_s)
  puts
  Log.error('Stacktrace:')
  Log.error(ex.backtrace.join("\n").to_s)
  exit 1
end
