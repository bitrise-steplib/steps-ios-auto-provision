require 'openssl'
require 'fastlane'

def ensure_app(bundle_id)
  app = Spaceship::Portal.app.find(bundle_id)
  if app.nil?
    normalized_bundle_id = bundle_id.tr('.', ' ')
    name = "Bitrise - (#{normalized_bundle_id})"
    log_done("registering app: #{name} with bundle id: (#{bundle_id})")

    app = Spaceship::Portal.app.create!(bundle_id: bundle_id, name: name)
  else
    log_done("app already registered: #{app.name} with bundle id: #{app.bundle_id}")
  end

  raise "failed to find or create app with bundle id: #{bundle_id}" unless app
  app
end

def certificate_matches(certificate1, certificate2)
  certificate1.serial == certificate2.serial
end

def find_development_portal_certificate(local_certificate)
  portal_development_certificates = Spaceship::Portal.certificate.development.all
  log_debug('no development Certificates belongs to the account in this team') if portal_development_certificates.to_a.empty?
  portal_development_certificates.each do |cert|
    portal_certificate = cert.download
    return cert if certificate_matches(local_certificate, portal_certificate)
  end

  log_debug('no development Certificates matches')
  nil
end

def find_production_portal_certificate(local_certificate)
  portal_production_certificates = Spaceship::Portal.certificate.production.all
  log_debug('no production Certificates belongs to the account in this team') if portal_production_certificates.to_a.empty?
  portal_production_certificates.each do |cert|
    portal_certificate = cert.download
    return cert if certificate_matches(local_certificate, portal_certificate)
  end

  log_debug('no production Certificates matches')
  nil
end

def ensure_test_devices(test_devices)
  if test_devices.to_a.empty?
    log_done('no test devices registered on bitrise')
    return
  end

  portal_devices = Spaceship::Portal.device.all(mac: false, include_disabled: true) || []
  test_devices.each do |test_device|
    registered_test_device = nil

    portal_devices.each do |portal_device|
      next unless portal_device.udid == test_device.uuid

      registered_test_device = portal_device
      log_done("test device #{registered_test_device.name} (#{registered_test_device.udid}) already registered")
      break
    end

    unless registered_test_device
      registered_test_device = Spaceship::Portal.device.create!(name: test_device.title, udid: test_device.uuid)
      log_done("registering test device #{registered_test_device.name} (#{registered_test_device.udid})")
    end

    raise 'failed to find or create device' unless registered_test_device

    registered_test_device.enable!
  end
end

def find_profile_by_bundle_id(profiles, bundle_id)
  matching = []
  profiles.each do |profile|
    matching.push(profile) if profile.app.bundle_id == bundle_id
  end

  matching
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
    log_debug("#{certificate.name} not registered in profile: #{profile.name}")
    certificates.push(certificate)
    profile.certificates = certificates
  end

  profile
end

def ensure_provisioning_profile(certificate, app, distributon_type)
  portal_profile_class = nil
  case distributon_type
  when 'development'
    portal_profile_class = Spaceship::Portal.provisioning_profile.development
  when 'app-store'
    portal_profile_class = Spaceship::Portal.provisioning_profile.app_store
  when 'ad-hoc'
    portal_profile_class = Spaceship::Portal.provisioning_profile.ad_hoc
  when 'enterprise'
    portal_profile_class = Spaceship::Portal.provisioning_profile.in_house
  else
    raise "invalid distribution type provided: #{distributon_type}, available: [development, app-store, ad-hoc, enterprise]"
  end

  profiles = find_profile_by_bundle_id(portal_profile_class.all, app.bundle_id)
  # Both app_store.all and ad_hoc.all return the same
  # This is the case since September 2016, since the API has changed
  # and there is no fast way to get the type when fetching the profiles
  # Distinguish between App Store and Ad Hoc profiles
  if distributon_type == 'app-store'
    profiles = profiles.find_all { |current| !current.is_adhoc? }
  elsif distributon_type == 'ad-hoc'
    profiles = profiles.find_all(&:is_adhoc?)
  end

  profile = nil
  if profiles.to_a.empty?
    log_done("generating #{distributon_type} provisioning profile for bundle id: #{app.bundle_id}")
    profile = portal_profile_class.create!(bundle_id: app.bundle_id, certificate: certificate, name: "Bitrise #{distributon_type} - (#{app.bundle_id})")
  else
    if profiles.count > 1
      log_warning("multiple #{distributon_type} provisionig profiles for bundle id: #{app.bundle_id}, using first:")
      profiles.each_with_index { |prof, index| log_warning("#{index}. #{prof.name}") }
    end

    profile = profiles.first
    log_done("#{distributon_type} profile: #{profile.name} (#{profile.uuid}) for bundle id (#{app.bundle_id}) already exist")

    # ensure certificate is included
    log_debug("ensure #{certificate.name} is included in profile")
    profile = ensure_profile_certificate(profile, certificate)

    # add all available devices to the profile
    if ['development', 'ad-hoc'].include?(distributon_type)
      log_debug('update profile devices')
      profile.devices = Spaceship::Portal.device.all
    end

    profile = profile.update!
  end

  raise "failed to find or create provisioning profile for bundle id: #{app.bundle_id}" unless profile
  profile
end
