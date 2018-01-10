require 'openssl'
require 'spaceship'

def result_string(ex)
  result = ex.preferred_error_info
  return nil unless result
  result.join(' ')
end

def run_and_handle_portal_function
  yield
rescue Spaceship::Client::UnexpectedResponse => ex
  message = result_string(ex)
  raise ex unless message
  raise message
end

def ensure_app(bundle_id)
  app = Spaceship::Portal.app.find(bundle_id)
  if app.nil?
    normalized_bundle_id = bundle_id.tr('.', ' ')
    name = "Bitrise - (#{normalized_bundle_id})"
    Log.success("registering app: #{name} with bundle id: (#{bundle_id})")

    app = nil
    begin
      run_and_handle_portal_function { app = Spaceship::Portal.app.create!(bundle_id: bundle_id, name: name) }
    rescue => ex
      message = ex.to_s
      if message =~ /An App ID with Identifier .* is not available/i
        raise message + "\nPossible solutions: https://stackoverflow.com/search?q=An+App+ID+with+Identifier+is+not+available"
      end
      raise ex
    end
  else
    Log.success("app already registered: #{app.name} with bundle id: #{app.bundle_id}")
  end

  raise "failed to find or create app with bundle id: #{bundle_id}" unless app
  app
end

def certificate_matches(certificate1, certificate2)
  common_name_match = certificate_common_name(certificate1) == certificate_common_name(certificate2)
  serial_match = certificate1.serial == certificate2.serial

  if common_name_match && !serial_match
    Log.warn("provided an older version of #{certificate_common_name(certificate1)} certificate, please provide the most recent version of the certificate")
  end

  serial_match
end

def find_development_portal_certificate(local_certificate)
  portal_development_certificates = nil
  run_and_handle_portal_function { portal_development_certificates = Spaceship::Portal.certificate.development.all }
  Log.debug('no development Certificates belongs to the account in this team') if portal_development_certificates.to_a.empty?

  portal_development_certificates.each do |cert|
    unless cert.can_download
      Log.debug("development Certificate: #{cert.name} is not downloadable, skipping...")
      next
    end
    portal_certificate = cert.download
    return cert if certificate_matches(local_certificate, portal_certificate)
  end

  Log.debug('no development Certificates matches')
  nil
end

def find_production_portal_certificate(local_certificate)
  portal_production_certificates = nil
  run_and_handle_portal_function { portal_production_certificates = Spaceship::Portal.certificate.production.all }

  if portal_production_certificates.to_a.empty?
    run_and_handle_portal_function { portal_production_certificates = Spaceship::Portal.certificate.in_house.all }
  end

  Log.debug('no production Certificates belongs to the account in this team') if portal_production_certificates.to_a.empty?

  portal_production_certificates.each do |cert|
    unless cert.can_download
      Log.debug("production Certificate: #{cert.name} is not downloadable, skipping...")
      next
    end
    portal_certificate = cert.download
    return cert if certificate_matches(local_certificate, portal_certificate)
  end

  Log.debug('no production Certificates matches')
  nil
end

def ensure_test_devices(test_devices)
  if test_devices.to_a.empty?
    Log.success('no test devices registered on bitrise')
    return
  end

  portal_devices = nil
  run_and_handle_portal_function { portal_devices = Spaceship::Portal.device.all(mac: false, include_disabled: true) || [] }
  test_devices.each do |test_device|
    registered_test_device = nil

    portal_devices.each do |portal_device|
      next unless portal_device.udid == test_device.udid

      registered_test_device = portal_device
      Log.success("test device #{registered_test_device.name} (#{registered_test_device.udid}) already registered")
      break
    end

    unless registered_test_device
      registered_test_device = nil
      run_and_handle_portal_function { registered_test_device = Spaceship::Portal.device.create!(name: test_device.name, udid: test_device.udid) }
      Log.success("registering test device #{registered_test_device.name} (#{registered_test_device.udid})")
    end

    raise 'failed to find or create device' unless registered_test_device

    registered_test_device.enable!
  end
end

def ensure_profile_certificate(profile, certificate)
  certificate_included = false

  certificates = profile.certificates
  certificates.each do |cert|
    if cert.id == certificate.id
      certificate_included = true
      break
    end
  end

  unless certificate_included
    Log.debug("#{certificate.name} not registered in profile: #{profile.name}")
    certificates.push(certificate)
    profile.certificates = certificates
  end

  profile
end

def ensure_provisioning_profile(certificate, app, distribution_type, allow_retry = true)
  portal_profile_class = nil
  case distribution_type
  when 'development'
    run_and_handle_portal_function { portal_profile_class = Spaceship::Portal.provisioning_profile.development }
  when 'app-store'
    run_and_handle_portal_function { portal_profile_class = Spaceship::Portal.provisioning_profile.app_store }
  when 'ad-hoc'
    run_and_handle_portal_function { portal_profile_class = Spaceship::Portal.provisioning_profile.ad_hoc }
  when 'enterprise'
    run_and_handle_portal_function { portal_profile_class = Spaceship::Portal.provisioning_profile.in_house }
  else
    raise "invalid distribution type provided: #{distribution_type}, available: [development, app-store, ad-hoc, enterprise]"
  end

  profiles = nil
  profile_name = "Bitrise #{distribution_type} - (#{app.bundle_id})"
  run_and_handle_portal_function { profiles = portal_profile_class.all.select { |profile| profile.app.bundle_id == app.bundle_id && profile.name == profile_name } }
  # Both app_store.all and ad_hoc.all return the same
  # This is the case since September 2016, since the API has changed
  # and there is no fast way to get the type when fetching the profiles
  # Distinguish between App Store and Ad Hoc profiles
  if distribution_type == 'app-store'
    profiles = profiles.reject(&:is_adhoc?)
  elsif distribution_type == 'ad-hoc'
    profiles = profiles.select(&:is_adhoc?)
  end

  if profiles.empty?
    Log.success("generating #{distribution_type} profile: #{profile_name}")
  else
    # it's easier to just create a new one, than to:
    # - add test devices
    # - add the certificate
    # - update profile
    # update seems to revoking the certificate, even if it is not neccessary
    # it has the same effects anyway, including a new UUID of the provisioning profile
    if profiles.count > 1
      Log.warn("multiple #{distribution_type} profiles found with name: #{profile_name}")
      profiles.each_with_index { |prof, index| Log.warn("#{index}. #{prof.name}") }
    end

    profiles.each do |profile|
      Log.warn("removing existing #{distribution_type} profile: #{profile.name}")
      profile.delete!
    end
  end

  profile = nil
  begin
    Log.warn("generating #{distribution_type} profile: #{profile_name}")
    run_and_handle_portal_function { profile = portal_profile_class.create!(bundle_id: app.bundle_id, certificate: certificate, name: profile_name) }
  rescue => ex
    # Failed to remove already existing managed profile, try it again!
    raise ex unless allow_retry
    raise ex unless ex.to_s =~ /Multiple profiles found with the name '(.*)'.\s*Please remove the duplicate profiles and try again./i

    Log.warn(ex.to_s)
    Log.warn('failed to regenerate the profile, retrying in 5 sec ...')
    sleep(5)
    ensure_provisioning_profile(certificate, app, distribution_type, false)
  end

  raise "failed to find or create provisioning profile for bundle id: #{app.bundle_id}" unless profile
  profile
end
