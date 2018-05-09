require_relative 'params'
require_relative 'log/log'
require_relative 'lib/autoprovision'

begin
  # Params
  params = Params.new
  params.print
  params.validate

  Log.verbose = (params.verbose_log == 'yes')
  ###

  # Developer Portal authentication
  Log.info('Developer Portal authentication')

  auth = AuthHelper.new
  auth.login(params.build_url, params.build_api_token, params.team_id)

  Log.success('authenticated')
  ###

  # Download certificates
  Log.info('Downloading Certificates')

  certificate_urls = params.certificate_urls.reject(&:empty?)
  raise 'no certificates provider' if certificate_urls.to_a.empty?

  cert_helper = CertificateHelper.new
  cert_helper.download_and_identify(certificate_urls, params.passphrases)
  ###

  # Anlyzing project
  Log.info('Analyzing project')

  project_helper = ProjectHelper.new(params.project_path, params.scheme, params.configuration)
  codesign_identity = project_helper.project_codesign_identity
  team_id = project_helper.project_team_id

  Log.print("project codesign identity: #{codesign_identity}")
  Log.print("project team id: #{team_id}")
  Log.print("uses xcode managed signing: #{project_helper.uses_xcode_auto_codesigning?}")
  Log.print("main target's platform: #{project_helper.platform}")

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

  raise 'failed to determine project development team' unless team_id
  ###

  # Matching project codesign identity with the uploaded certificates
  Log.info('Matching project codesign identity with the uploaded certificates')

  cert_helper.ensure_certificate(codesign_identity, team_id, params.distribution_type)
  ###

  # Ensure test devices
  if ['development', 'ad-hoc'].include?(params.distribution_type)
    Log.info('Ensure test devices on Developer Portal')
    Portal::DeviceClient.ensure_test_devices(auth.test_devices)
  end
  ###

  # Ensure Provisioning Profiles on Developer Portal
  Log.info('Ensure Provisioning Profiles on Developer Portal')

  profile_helper = ProfileHelper.new(project_helper, cert_helper)
  xcode_managed_signing = profile_helper.ensure_profiles(params.distribution_type, params.generate_profiles == 'yes')
  ###

  unless xcode_managed_signing
    # Apply code sign setting in project
    Log.info('Apply code sign setting in project')

    targets.each do |target_name|
      bundle_id = project_helper.target_bundle_id(target_name)

      puts
      Log.success("configure target: #{target_name} (#{bundle_id})")

      code_sign_identity = nil
      provisioning_profile = nil

      if cert_helper.development_certificate_info
        certificate = cert_helper.development_certificate_info.certificate
        code_sign_identity = certificate_common_name(certificate)

        portal_profile = profile_helper.profiles_by_bundle_id('development')[bundle_id].portal_profile
        provisioning_profile = portal_profile.uuid
      elsif cert_helper.production_certificate_info
        certificate = cert_helper.production_certificate_info.certificate
        code_sign_identity = certificate_common_name(certificate)

        portal_profile = profile_helper.profiles_by_bundle_id(params.distribution_type)[bundle_id].portal_profile
        provisioning_profile = portal_profile.uuid
      else
        raise "no codesign settings generated for target: #{target_name} (#{bundle_id})"
      end

      project_helper.force_code_sign_properties(target_name, team_id, code_sign_identity, provisioning_profile)
    end
    ###
  end

  # Install certificates
  Log.info('Install certificates')

  certificate_infos = []
  certificate_infos.push(cert_helper.development_certificate_info) if cert_helper.development_certificate_info
  certificate_infos.push(cert_helper.production_certificate_info) if cert_helper.production_certificate_info
  certificate_path_passphrase_map = Hash[certificate_infos.map { |info| [info.path, info.passphrase] }]

  keychain_helper = KeychainHelper.new(params.keychain_path, params.keychain_password)
  keychain_helper.install_certificates(certificate_path_passphrase_map)

  Log.success("#{certificate_path_passphrase_map.length} certificates installed")
  ###
rescue => ex
  puts
  Log.error('Error:')
  Log.error(ex.to_s)
  puts
  Log.error('Stacktrace (for debugging):')
  Log.error(ex.backtrace.join("\n").to_s)
  exit 1
end
