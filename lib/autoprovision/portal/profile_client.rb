require 'spaceship'

require_relative 'app_client'

module Portal
  # ProfileClient ...
  class ProfileClient
    @profiles = {}

    def self.ensure_xcode_managed_profile(bundle_id, entitlements, distribution_type, portal_certificate, platform, min_profile_days_valid)
      profiles = ProfileClient.fetch_profiles(true, platform)
      profile = ProfileClient.matching_profile(profiles, distribution_type, bundle_id, entitlements, portal_certificate, min_profile_days_valid)

      unless profile
        error_message = [
          "Failed to find #{distribution_type} Xcode managed provisioning profile for bundle id: #{bundle_id}.",
          'Please open your project in your local Xcode and generate and ipa file',
          'with the desired distribution type and by using Xcode managed codesigning.',
          'This will create / refresh the desired managed profiles.'
        ].join("\n")
        raise error_message
      end

      profile
    end

    def self.ensure_manual_profile(certificate, app, entitlements, distribution_type, platform, min_profile_days_valid, allow_retry = true, portal_devices)
      all_profiles = ProfileClient.fetch_profiles(false, platform)

      # search for the Bitrise managed profile
      profile_name = "Bitrise #{distribution_type} - (#{app.bundle_id})"
      profile = all_profiles.select { |prof| prof.name == profile_name }.first
      valid = !profile.nil?

      # check the profile's bundle id
      if valid
        if profile.app.bundle_id != app.bundle_id
          Log.debug("Profile (#{profile.name}) bundle id: #{profile.app.bundle_id}, should be: #{app.bundle_id}")
          valid = false
        end
      end

      # check the profile's distribution type
      if valid
        distribution_methods = {
          'development' => 'limited',
          'app-store' => 'store',
          'ad-hoc' => 'adhoc',
          'enterprise' => 'inhouse'
        }
        desired_distribution_method = distribution_methods[distribution_type]

        unless profile.distribution_method == desired_distribution_method
          Log.debug("Profile (#{profile.name}) distribution type: #{profile.distribution_method}, should be: #{desired_distribution_method}")
          valid = false
        end
      end

      # check the profile expiry
      if valid
        # Increment the current time with days in seconds (1 day = 86400 secs) the profile has to be valid for
        expire = Time.now + (min_profile_days_valid * 86_400)

        if Time.parse(profile.expires.to_s) < expire
          if min_profile_days_valid > 0
            Log.debug("Profile (#{profile.name}) is not valid for #{min_profile_days_valid} days")
          else
            Log.debug("Profile (#{profile.name}) expired at: #{profile.expires}")
          end

          valid = false
        end
      end

      # check if project capabilities are enabled for the profile
      if valid
        unless AppClient.all_services_enabled?(profile.app, entitlements)
          Log.debug("Profile (#{profile.name}) does not contain every required services")
          valid = false
        end
      end

      # check if the profile contains the given certificate
      if valid
        unless include_certificate?(profile, certificate)
          Log.debug("Profile (#{profile.name}) does not contain certificate: #{certificate.name}")
          valid = false
        end
      end

      # check if the development and ad-hoc profile's device list is up to date
      if valid && (profile.distribution_method == distribution_methods['development'] || profile.distribution_method == distribution_methods['ad-hoc'])
        Log.info('Check the device list in the profile')  
        
        profile_device_udids = profile.devices.map { |device| device.udid }
        if platform == :tvos
          filtered_portal_device_udids = portal_devices.reject { |device| device.device_type != "tvOS" || device.status == "r" }.map { |device| device.udid}
        else 
          filtered_portal_device_udids = portal_devices.reject { |device| device.device_type == "tvOS" || device.status == "r" }.map { |device| device.udid}
        end

        if !(filtered_portal_device_udids - profile_device_udids).empty?
          Log.warn("Profile (#{profile.name}) does not contain all the test devices") 
          Log.print("Missing devices:\n#{(filtered_portal_device_udids - profile_device_udids).join("\n")}")
          valid = false
        else 
          Log.print("Profile (#{profile.name}) contains all the test devices") 
        end
      end

      return profile if valid

      # profile name needs to be unique
      if !valid && !profile.nil?
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

    def self.matching_profile(profiles, distribution_type, bundle_id, entitlements, portal_certificate, min_profile_days_valid = 0)
      # filter for distribution type
      distribution_methods = {
        'development' => 'limited',
        'app-store' => 'store',
        'ad-hoc' => 'adhoc',
        'enterprise' => 'inhouse'
      }
      profiles = profiles.select do |profile|
        profile.distribution_method == distribution_methods[distribution_type]
      end
      Log.debug("#{distribution_type} profiles (#{profiles.length}):")
      profiles.each do |profile|
        Log.debug(profile.name)
      end

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

      # Increment the current time with days in seconds (1 day = 86400 secs) the profile has to be valid for
      expire = Time.now + (min_profile_days_valid * 86_400)

      # remove profiles which does not contains all of the provided services (entitlements)
      # and the profiles which does not contains the provided certificate (portal_certificate)
      filtered_full_matching_profiles = []
      full_matching_profiles.each do |profile|
        if Time.parse(profile.expires.to_s) < expire
          if min_profile_days_valid > 0
            Log.debug("Profile (#{profile.name}) matches to target: #{bundle_id}, but it's not valid for #{min_profile_days_valid} days")
          else
            Log.debug("Profile (#{profile.name}) matches to target: #{bundle_id}, but expired at: #{profile.expires}")
          end
          next
        end

        unless AppClient.all_services_enabled?(profile.app, entitlements)
          Log.debug("Profile (#{profile.name}) matches to target: #{bundle_id}, but has missing services")
          next
        end

        unless include_certificate?(profile, portal_certificate)
          Log.debug("Profile (#{profile.name}) matches to target: #{bundle_id}, but does not contain the provided certificate")
          next
        end

        filtered_full_matching_profiles.push(profile)
      end

      # prefer the full bundle id match over the glob match
      return filtered_full_matching_profiles[0] unless filtered_full_matching_profiles.empty?

      filtered_matching_profiles = []
      matching_profiles.each do |profile|
        if Time.parse(profile.expires.to_s) < expire
          if min_profile_days_valid > 0
            Log.debug("Profile (#{profile.name}) matches to target: #{bundle_id}, but it's not valid for #{min_profile_days_valid} days")
          else
            Log.debug("Profile (#{profile.name}) matches to target: #{bundle_id}, but expired at: #{profile.expires}")
          end
          next
        end

        unless AppClient.all_services_enabled?(profile.app, entitlements)
          Log.debug("Wildcard Profile (#{profile.name}) matches to target: #{bundle_id}, but has missing services")
          next
        end

        unless include_certificate?(profile, portal_certificate)
          Log.debug("Wildcard Profile (#{profile.name}) matches to target: #{bundle_id}, but does not contain the provided certificate")
          next
        end

        filtered_matching_profiles.push(profile)
      end

      filtered_matching_profiles[0] unless filtered_matching_profiles.empty?
      nil
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

    def self.include_certificate?(profile, certificate)
      profile.certificates.each do |portal_certificate|
        return true if portal_certificate.id == certificate.id
      end
      false
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
