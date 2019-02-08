require 'spaceship'

require_relative 'app_client'

module Portal
  # ProfileClient ...
  class ProfileClient
    @profiles = {}

    def self.ensure_xcode_managed_profile(bundle_id, entitlements, distribution_type, certificate, platform, min_profile_days_valid)
      profiles = ProfileClient.fetch_profiles(true, platform)

      # Separate matching profiles
      # full_matching_profiles contains profiles which bundle id equals to the provided bundle_id
      # matching_profiles contains profiles which bundle id glob matches to the provided bundle_id
      full_matching_profiles = []
      matching_profiles = []
      profiles.each do |profile|
        if profile.app.bundle_id == bundle_id
          full_matching_profiles.push(profile)
          next
        end

        matching_profiles.push(profile) if File.fnmatch(profile.app.bundle_id, bundle_id)
      end

      profiles = full_matching_profiles.select do |profile|
        distribution_type_matches?(profile, distribution_type) &&
          !expired?(profile, min_profile_days_valid) &&
          all_services_enabled?(entitlements, entitlements) &&
          include_certificate?(profile, certificate)
      end

      return profiles[0] unless profiles.empty?

      profiles = matching_profiles.select do |profile|
        distribution_type_matches?(profile, distribution_type) &&
          !expired?(profile, min_profile_days_valid) &&
          all_services_enabled?(entitlements, entitlements) &&
          include_certificate?(profile, certificate)
      end

      return profiles[0] unless profiles.empty?

      raise [
        "Failed to find #{distribution_type} Xcode managed provisioning profile for bundle id: #{bundle_id}.",
        'Please open your project in your local Xcode and generate and ipa file',
        'with the desired distribution type and by using Xcode managed codesigning.',
        'This will create / refresh the desired managed profiles.'
      ].join("\n")
    end

    def self.ensure_manual_profile(certificate, app, entitlements, distribution_type, platform, min_profile_days_valid, allow_retry = true, portal_devices)
      all_profiles = ProfileClient.fetch_profiles(false, platform)

      # search for the Bitrise managed profile
      profile_name = "Bitrise #{distribution_type} - (#{app.bundle_id})"
      profile = all_profiles.select { |prof| prof.name == profile_name }.first

      return profile if !profile.nil? &&
                        bundle_id_matches?(profile, app) &&
                        distribution_type_matches?(profile, distribution_type) &&
                        !expired?(profile, min_profile_days_valid) &&
                        all_services_enabled?(entitlements, entitlements) &&
                        include_certificate?(profile, certificate) &&
                        device_list_up_to_date?(profile, distribution_type, platform, portal_devices)

      # profile name needs to be unique
      unless profile.nil?
        profile.delete!
        ProfileClient.clear_cache(false, platform)
      end

      begin
        Log.print("generating profile: #{profile_name}")
        profile_class = portal_profile_class(distribution_type)
        run_and_handle_portal_function { profile = profile_class.create!(bundle_id: app.bundle_id, certificate: certificate, name: profile_name, sub_platform: platform == :tvos ? 'tvOS' : nil) }
      rescue => ex
        raise ex unless allow_retry
        raise ex unless ex.to_s =~ /Multiple profiles found with the name/i

        # The profile already exist, paralell step run can produce this issue
        Log.debug_exception(ex)
        Log.debug('failed to generate the profile, retrying in 2 sec ...')
        sleep(2)
        ProfileClient.clear_cache(false, platform)
        return ProfileClient.ensure_manual_profile(certificate, app, entitlements, distribution_type, platform, min_profile_days_valid, false, portal_devices)
      end

      raise "failed to find or create provisioning profile for bundle id: #{app.bundle_id}" unless profile

      profile
    end

    def self.bundle_id_matches?(profile, app)
      if profile.app.bundle_id != app.bundle_id
        Log.debug("Profile (#{profile.name}) bundle id: #{profile.app.bundle_id}, should be: #{app.bundle_id}")
        return false
      end
      true
    end

    def self.distribution_type_matches?(profile, distribution_type)
      distribution_methods = {
        'development' => 'limited',
        'app-store' => 'store',
        'ad-hoc' => 'adhoc',
        'enterprise' => 'inhouse'
      }
      desired_distribution_method = distribution_methods[distribution_type]

      unless profile.distribution_method == desired_distribution_method
        Log.debug("Profile (#{profile.name}) distribution type: #{profile.distribution_method}, should be: #{desired_distribution_method}")
        return false
      end
      false
    end

    def self.expired?(profile, min_profile_days_valid)
      # Increment the current time with days in seconds (1 day = 86400 secs) the profile has to be valid for
      expire = Time.now + (min_profile_days_valid * 86_400)

      if Time.parse(profile.expires.to_s) < expire
        if min_profile_days_valid > 0
          Log.debug("Profile (#{profile.name}) is not valid for #{min_profile_days_valid} days")
        else
          Log.debug("Profile (#{profile.name}) expired at: #{profile.expires}")
        end

        return true
      end
      false
    end

    def self.all_services_enabled?(profile, entitlements)
      unless AppClient.all_services_enabled?(profile.app, entitlements)
        Log.debug("Profile (#{profile.name}) does not contain every required services")
        return false
      end
      true
    end

    def self.include_certificate?(profile, certificate)
      profile.certificates.each do |portal_certificate|
        return true if portal_certificate.id == certificate.id
      end
      Log.debug("Profile (#{profile.name}) does not contain certificate: #{certificate.name}")
      false
    end

    def self.device_list_up_to_date?(profile, distribution_type, platform, portal_devices)
      # check if the development and ad-hoc profile's device list is up to date
      if ['development', 'ad-hoc'].include?(distribution_type) && !portal_devices.nil?
        Log.info('Check the device list in the profile')

        profile_device_udids = profile.devices.map(&:udid)
        filtered_portal_device_udids = if platform == :tvos
                                         # Remove all the NON tvOS devices and the disabled ones
                                         portal_devices.reject { |device| device.device_type != 'tvOS' || device.status == 'r' }.map(&:udid)
                                       else
                                         # Remove all the tvOS devices and the disabled ones
                                         portal_devices.reject { |device| device.device_type == 'tvOS' || device.status == 'r' }.map(&:udid)
                                       end

        if !(filtered_portal_device_udids - profile_device_udids).empty?
          Log.warn("Profile (#{profile.name}) does not contain all the test devices")
          Log.print("Missing devices:\n#{(filtered_portal_device_udids - profile_device_udids).join("\n")}")

          false
        else
          Log.print("Profile (#{profile.name}) contains all the test devices")
        end
      end

      true
    end

    def self.clear_cache(xcode_managed, platform)
      @profiles[platform].to_h[xcode_managed] = nil
    end

    def self.fetch_profiles(xcode_managed, platform)
      cached = @profiles[platform].to_h[xcode_managed]
      return cached unless cached.to_a.empty?

      profiles = []
      run_and_handle_portal_function { profiles = Spaceship::Portal.provisioning_profile.all(mac: false, xcode: xcode_managed) }
      # Log.debug("all profiles (#{profiles.length}):")
      # profiles.each do |profile|
      #   Log.debug("#{profile.name}")
      # end

      # filter for sub_platform
      profiles = profiles.reject do |profile|
        if platform == :tvos
          profile.sub_platform.to_s.casecmp('tvos') == -1
        else
          profile.sub_platform.to_s.casecmp('tvos').zero?
        end
      end
      # Log.debug("subplatform #{platform} profiles (#{profiles.length}):")
      # profiles.each do |profile|
      #   Log.debug("#{profile.name}")
      # end

      # update the cache
      platform_profiles = @profiles[platform].to_h
      platform_profiles[xcode_managed] = profiles
      @profiles[platform] = platform_profiles

      profiles
    end

    def self.portal_profile_class(distribution_type)
      case distribution_type
      when 'development'
        Spaceship::Portal.provisioning_profile.development
      when 'app-store'
        Spaceship::Portal.provisioning_profile.app_store
      when 'ad-hoc'
        Spaceship::Portal.provisioning_profile.ad_hoc
      when 'enterprise'
        Spaceship::Portal.provisioning_profile.in_house
      else
        raise "invalid distribution type provided: #{distribution_type}, available: [development, app-store, ad-hoc, enterprise]"
      end
    end
  end
end
