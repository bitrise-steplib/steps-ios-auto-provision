require 'openssl'

require_relative 'log/log'
require_relative 'project_helper/project_helper'
require_relative 'http_helper/http_helper'
require_relative 'http_helper/portal_data'
require_relative 'auto-provision/authenticator'
require_relative 'auto-provision/generator'
require_relative 'auto-provision/app_services'
require_relative 'keychain/keychain'

# CertificateInfo
class CertificateInfo
  attr_accessor :path
  attr_accessor :passphrase
  attr_accessor :certificate
  attr_accessor :portal_certificate

  @path = nil
  @passphrase = nil
  @certificate = nil
  @portal_certificate = nil
end

# ProfileInfo
class ProfileInfo
  attr_accessor :path
  attr_accessor :profile
  attr_accessor :portal_profile

  @path = nil
  @profile = nil
  @portal_profile = nil
end

# CodesignSettings
class CodesignSettings
  attr_accessor :team_id
  attr_accessor :development_certificate_info
  attr_accessor :production_certificate_info
  attr_accessor :bundle_id_development_profile
  attr_accessor :bundle_id_production_profile

  @team_id = nil
  @development_certificates = []
  @production_certificates = []
  @bundle_id_development_profile = {}
  @bundle_id_production_profile = {}
end

# Params
class Params
  attr_accessor :build_url
  attr_accessor :build_api_token
  attr_accessor :team_id
  attr_accessor :certificate_urls
  attr_accessor :passphrases
  attr_accessor :distributon_type
  attr_accessor :project_path
  attr_accessor :keychain_path
  attr_accessor :keychain_password
  attr_accessor :verbose_log

  def initialize
    @build_url = ENV['build_url'] || ''
    @build_api_token = ENV['build_api_token'] || ''
    @team_id = ENV['team_id'] || ''
    @certificate_urls = ENV['certificate_urls'] || ''
    @passphrases = ENV['passphrases'] || ''
    @distributon_type = ENV['distributon_type'] || ''
    @project_path = ENV['project_path'] || ''
    @keychain_path = ENV['keychain_path'] || ''
    @keychain_password = ENV['keychain_password'] || ''
    @verbose_log = ENV['verbose_log'] || ''
  end

  def print
    log_info('Params:')
    log_details("team_id: #{@team_id}")
    log_details("certificate_urls: #{secure_value(@certificate_urls)}")
    log_details("passphrases: #{secure_value(@passphrases)}")
    log_details("distributon_type: #{@distributon_type}")
    log_details("project_path: #{@project_path}")
    log_details("build_url: #{@build_url}")
    log_details("build_api_token: #{secure_value(@build_api_token)}")
    log_details("keychain_path: #{@keychain_path}")
    log_details("keychain_password: #{secure_value(@keychain_password)}")
    log_details("verbose_log: #{@verbose_log}")
  end

  def validate
    raise 'missing: build_url' if @build_url.empty?
    raise 'missing: build_api_token' if @build_api_token.empty?
    raise 'missing: team_id' if @team_id.empty?
    raise 'missing: certificate_urls' if @certificate_urls.empty?
    raise 'missing: distributon_type' if @distributon_type.empty?
    raise 'missing: project_path' if @project_path.empty?
    raise 'missing: keychain_path' if @keychain_path.empty?
    raise 'missing: keychain_password' if @keychain_password.empty?
    raise 'missing: verbose_log' if @verbose_log.empty?
  end
end

def secure_value(value)
  return '' if value.empty?
  '***'
end

def split_pipe_separated_list(list)
  separator_count = list.count('|')
  char_count = list.length

  return [''] if char_count.zero?
  return [list] unless list.include?('|')
  return Array.new(separator_count + 1, '') if separator_count == char_count
  list.split('|').map(&:strip)
end

def certificate_common_name(certificate)
  common_name = certificate.subject.to_a.find { |name, _, _| name == 'CN' }[1]
  common_name = common_name.force_encoding('UTF-8')
  common_name
end

def certificate_team_id(certificate)
  certificate.subject.to_a.find { |name, _, _| name == 'OU' }[1]
end

def find_certificate_info_by_identity(identity, certificate_infos)
  certificate_infos.each do |certificate_info|
    common_name = certificate_common_name(certificate_info.certificate)
    return certificate_info if common_name.downcase.include?(identity.downcase)
  end
  nil
end

def find_certificate_infos_by_team_id(team_id, certificate_infos)
  matching_certificate_infos = []
  certificate_infos.each do |certificate_info|
    org_unit = certificate_team_id(certificate_info.certificate)
    matching_certificate_infos.push(certificate_info) if org_unit.downcase.include?(team_id.downcase)
  end
  matching_certificate_infos
end

def find_matching_codesign_identity_info(identity_name, team_id, certificate_infos)
  if identity_name
    certificate_info = find_certificate_info_by_identity(identity_name, certificate_infos)
    return certificate_info if certificate_info
  end

  team_certificate_infos = find_certificate_infos_by_team_id(team_id, certificate_infos)
  return team_certificate_infos[0] if team_certificate_infos.to_a.length == 1
  log_details('no development certificate found') if team_certificate_infos.to_a.empty?
  log_warning("#{team_certificate_infos.length} development certificate found") if team_certificate_infos.to_a.length > 1
end

begin
  # Params
  params = Params.new
  params.print
  params.validate

  DEBUG_LOG = (params.verbose_log == 'yes')
  ###

  # Anlyzing project
  log_info('Anlyzing project')

  project_helper = ProjectHelper.new(params.project_path)

  project_target_bundle_id_map = project_helper.project_target_bundle_id_map
  raise 'no targets found' if project_target_bundle_id_map.to_a.empty?

  project_target_entitlements_map = project_helper.project_target_entitlements_map
  raise 'no targets found' if project_target_entitlements_map.to_a.empty?
  raise 'analyzer failed' unless project_target_bundle_id_map.to_a.length == project_target_entitlements_map.to_a.length

  project_codesign_identity_map = {}
  project_team_id_map = {}
  project_target_bundle_id_map.each do |project_path, target_bundle_id|
    log_done("project: #{project_path}")

    idx = 0
    target_bundle_id.each do |target, bundle_id|
      idx += 1
      entitlements_count = (project_target_entitlements_map[project_path][target] || []).length

      log_details("target ##{idx}: #{target} (#{bundle_id}) with #{entitlements_count} services")
    end

    codesign_identity = project_helper.codesign_identity(project_path)
    raise 'failed to determine project codesign identity' unless codesign_identity

    team_id = project_helper.team_id(project_path)
    raise 'failed to determine project team id' unless team_id

    if !params.team_id.to_s.empty? && params.team_id != team_id
      log_warning("different team id defined: #{params.team_id} then the project's one: #{team_id}")
      log_warning("using defined team id: #{params.team_id}")
      log_warning("droping project codesign identity: #{codesign_identity}")

      team_id = params.team_id
      codesign_identity = nil
    end

    project_codesign_identity_map[project_path] = codesign_identity if codesign_identity
    project_team_id_map[project_path] = team_id
  end
  ###

  # Developer Portal authentication
  log_info('Developer Portal authentication')

  portal_data = get_developer_portal_data(params.build_url, params.build_api_token)
  portal_data.validate

  log_debug("session cookie: #{portal_data.session_cookies}\n")
  session = convert_tfa_cookies(portal_data.session_cookies)
  log_debug("converted session cookie: #{session}\n")

  developer_portal_authentication(portal_data.apple_id, portal_data.password, session, params.team_id)

  log_done('authenticated')
  ###

  # Download certificates
  log_info('Downloading Certificates')

  certificate_urls = split_pipe_separated_list(params.certificate_urls).reject(&:empty?)
  raise 'no certificates provider' if certificate_urls.to_a.empty?

  passphrases = split_pipe_separated_list(params.passphrases)
  raise "certificates count (#{certificate_urls.length}) and passphrases count (#{passphrases.length}) should match" unless certificate_urls.length == passphrases.length

  certificate_infos = []
  certificate_urls.each_with_index do |url, idx|
    log_debug("downloading certificate ##{idx + 1}")
    path = download_to_tmp_file(url, "Certrificate#{idx}.p12")
    log_debug("certificate path: #{path}")

    certificate_info = CertificateInfo.new
    certificate_info.path = path
    certificate_info.passphrase = passphrases[idx]
    certificate_infos.push(certificate_info)
  end
  log_done("#{certificate_infos.length} certificates downloaded")
  ###

  # Identify Certificates on developer Portal
  log_info('Identify Certificates on developer Portal')

  development_certificate_infos = []
  production_certificate_infos = []
  certificate_infos.each do |certificate_info|
    certificate_content = File.read(certificate_info.path)
    log_debug("searching for Certificate (#{certificate_info.path})")
    raise "Invalid certificate file #{certificate_info.path}: empty" if certificate_content.to_s.empty?

    p12 = OpenSSL::PKCS12.new(certificate_content, certificate_info.passphrase)

    portal_certificate = find_development_portal_certificate(p12.certificate)
    if portal_certificate
      log_done("development Certificate found: #{portal_certificate.name}")
      raise 'multiple development certificates provided: step can handle only one development (and only one production) certificate' unless development_certificate_infos.empty?

      certificate_info.certificate = p12.certificate
      certificate_info.portal_certificate = portal_certificate
      development_certificate_infos.push(certificate_info)
    end

    portal_certificate = find_production_portal_certificate(p12.certificate)
    next unless portal_certificate

    log_done("production Certificate found: #{portal_certificate.name}")
    raise 'multiple production certificates provided: step can handle only one production (and only one development) certificate' unless production_certificate_infos.empty?

    certificate_info.certificate = p12.certificate
    certificate_info.portal_certificate = portal_certificate
    production_certificate_infos.push(certificate_info)
  end
  raise 'no development nor production certificate identified on development portal' if development_certificate_infos.empty? && production_certificate_infos.empty?
  ###

  # Matching project codesign identity with the uploaded certificates
  log_info('Matching project codesign identity with the uploaded certificates')

  project_codesign_settings = {}
  project_target_bundle_id_map.each_key do |path|
    log_details("checking project: #{path}")
    codesign_settings = CodesignSettings.new

    identity_name = project_codesign_identity_map[path]
    team_id = project_team_id_map[path]

    certificate_info = find_matching_codesign_identity_info(identity_name, team_id, development_certificate_infos)
    if certificate_info
      codesign_settings.team_id = certificate_team_id(certificate_info.certificate)
      codesign_settings.development_certificate_info = certificate_info
      project_codesign_settings[path] = codesign_settings
      log_done("using: #{certificate_common_name(certificate_info.certificate)}")
      next
    end

    certificate_info = find_matching_codesign_identity_info(identity_name, team_id, production_certificate_infos)
    if certificate_info
      codesign_settings.team_id = certificate_team_id(certificate_info.certificate)
      codesign_settings.production_certificate_info = certificate_info
      project_codesign_settings[path] = codesign_settings
      log_done("using: #{certificate_common_name(certificate_info.certificate)}")
      next
    end

    raise 'Failed to find desired certificate in uploaded certificates'
  end
  ###

  # Ensure certificate for defined distribution type
  log_info('Ensure certificate for defined distribution type')

  project_codesign_settings.each do |path, codesign_settings|
    log_details("distribution type: #{params.distributon_type}")

    if params.distributon_type == 'development'
      raise "development distribution defined but no uploaded identity found in team: #{codesign_settings.team_id}" if codesign_settings.development_certificate_info.nil?
      log_done('certificate already found')
      next
    end

    unless codesign_settings.production_certificate_info.nil?
      log_done("#{params.distributon_type} certificate already found")
      next
    end

    certificate_info = find_matching_codesign_identity_info(nil, codesign_settings.team_id, production_certificate_infos)
    raise "#{params.distributon_type} distribution defined but no uploaded identity found in team: #{codesign_settings.team_id}" unless certificate_info

    codesign_settings.production_certificate_info = certificate_info
    log_done("using: #{certificate_common_name(certificate_info.certificate)}")
    project_codesign_settings[path] = codesign_settings
  end
  ###

  # Ensure test devices
  if ['development', 'ad-hoc'].include?(params.distributon_type)
    log_info('Ensure test devices on Developer Portal')
    ensure_test_devices(portal_data.test_devices)
  end
  ###

  # Ensure App IDs and Provisioning Profiles on Developer Portal
  log_info('Ensure App IDs and Provisioning Profiles on Developer Portal')

  project_target_bundle_id_map.each do |path, target_bundle_id|
    log_details("checking project: #{path}")
    codesign_settings = project_codesign_settings[path]
    development_certificate_info = codesign_settings.development_certificate_info
    production_certificate_info = codesign_settings.production_certificate_info

    bundle_id_development_profile = {}
    bundle_id_production_profile = {}

    target_entitlements = project_target_entitlements_map[path]
    target_bundle_id.each do |target, bundle_id|
      entitlements = target_entitlements[target]
      puts
      log_done("checking target: #{target} (#{bundle_id}) with #{entitlements.length} services")

      log_details("ensure App ID (#{bundle_id}) on Developer Portal")
      app = ensure_app(bundle_id)

      log_details("sync App ID (#{bundle_id}) Services")
      app = sync_app_services(app, entitlements)

      if development_certificate_info
        log_details('ensure Development Provisioning Profile on Developer Portal')
        portal_profile = ensure_provisioning_profile(development_certificate_info.portal_certificate, app, 'development')

        log_done("downloading development profile: #{portal_profile.name}")
        profile_path = download_profile(portal_profile)

        log_debug("profile path: #{profile_path}")

        profile_info = ProfileInfo.new
        profile_info.path = profile_path
        profile_info.portal_profile = portal_profile
        bundle_id_development_profile[bundle_id] = profile_info
      end

      next if params.distributon_type == 'development'
      next unless production_certificate_info

      log_details('ensure Production Provisioning Profile on Developer Portal')
      portal_profile = ensure_provisioning_profile(production_certificate_info.portal_certificate, app, params.distributon_type)

      log_done("downloading #{params.distributon_type} profile: #{portal_profile.name}")
      profile_path = download_profile(portal_profile)

      log_debug("profile path: #{profile_path}")

      profile_info = ProfileInfo.new
      profile_info.path = profile_path
      profile_info.portal_profile = portal_profile
      bundle_id_production_profile[bundle_id] = profile_info
    end

    codesign_settings.bundle_id_development_profile = bundle_id_development_profile
    codesign_settings.bundle_id_production_profile = bundle_id_production_profile
    project_codesign_settings[path] = codesign_settings
  end
  ###

  # Apply code sign setting in project
  log_info('Apply code sign setting in project')

  project_target_bundle_id_map.each do |path, target_bundle_id|
    log_details("checking project: #{path}")
    codesign_settings = project_codesign_settings[path]

    target_bundle_id.each do |target, bundle_id|
      puts
      log_done("checking target: #{target} (#{bundle_id})")

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
        raise "no codesign settings generated for target: #{target} (#{bundle_id})"
      end

      project_helper.force_code_sign_properties(path, target, team_id, code_sign_identity, provisioning_profile)
    end
  end
  ###

  # Install certificates
  log_info('Install certificates')

  keychain_helper = KeychainHelper.new(params.keychain_path, params.keychain_password)

  certificate_infos.each do |certificate_info|
    keychain_helper.import_certificate(certificate_info.path, certificate_info.passphrase)
  end

  keychain_helper.set_key_partition_list_if_needed
  keychain_helper.set_keychain_settings_default_lock
  keychain_helper.add_to_keychain_search_path
  keychain_helper.set_default_keychain
  keychain_helper.unlock_keychain

  log_done("#{certificate_infos.length} certificates installed")
  ###
rescue => ex
  puts
  log_error(ex.to_s)
  log_error(ex.backtrace.join("\n").to_s) if DEBUG_LOG
  exit 1
end
