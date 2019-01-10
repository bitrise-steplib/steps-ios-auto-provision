require 'spaceship'

require_relative 'app_client'

module Portal
  # ProfileClient ...
  class ProfileClient
    @profiles = {}

    def self.ensure_xcode_managed_profile(bundle_id, entitlements, distribution_type, portal_certificate, platform, min_profile_days_valid)
      profiles = ProfileClient.fetch_profiles(distribution_type, true, platform)
      profile = ProfileClient.matching_profile(profiles, bundle_id, entitlements, portal_certificate, min_profile_days_valid)

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

    def self.ensure_manual_profile(certificate, app, entitlements, distribution_type, platform, min_profile_days_valid, allow_retry = true)
      profile_name = "Bitrise #{distribution_type} - (#{app.bundle_id})"

      profiles = ProfileClient.fetch_profiles(distribution_type, false, platform)
      profiles = profiles.select { |profile| profile.app.bundle_id == app.bundle_id && profile.name == profile_name }
      existing_profile = ProfileClient.matching_profile(profiles, app.bundle_id, entitlements, certificate, min_profile_days_valid)

      return existing_profile if existing_profile

      if profiles.empty?
        Log.debug("generating #{distribution_type} profile: #{profile_name}")
      else
        # it's easier to just create a new one, than to:
        # - add test devices
        # - add the certificate
        # - update profile
        # update seems to revoking the certificate, even if it is not neccessary
        # it has the same effects anyway, including a new UUID of the provisioning profile
        if profiles.size > 1
          Log.debug("multiple #{distribution_type} profiles found with name: #{profile_name}")
          profiles.each_with_index { |prof, index| Log.debug("#{index}. #{prof.name}") }
        end

        profiles.each do |profile|
          Log.debug("removing existing #{distribution_type} profile: #{profile.name}")
          profile.delete!
        end
      end

      profile = nil
      begin
        Log.debug("generating #{distribution_type} profile: #{profile_name}")
        profile_class = portal_profile_class(distribution_type)
        run_and_handle_portal_function { profile = profile_class.create!(bundle_id: app.bundle_id, certificate: certificate, name: profile_name, sub_platform: platform == :tvos ? 'tvOS' : nil) }
      rescue => ex
        raise ex unless allow_retry
        raise ex unless ex.to_s =~ /Multiple profiles found with the name/i

        # Failed to remove already existing managed profile, or
        # the profile already exist, may someone generated it during this step run
        Log.debug_exception(ex)
        Log.debug('failed to regenerate the profile, retrying in 2 sec ...')
        sleep(2)
        ProfileClient.clear_cache(distribution_type, false, platform)
        ProfileClient.ensure_manual_profile(certificate, app, entitlements, distribution_type, platform, min_profile_days_valid, false)
      end

      raise "failed to find or create provisioning profile for bundle id: #{app.bundle_id}" unless profile
      profile
    end

    def self.matching_profile(profiles, bundle_id, entitlements, portal_certificate, min_profile_days_valid = 0)
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

    def self.clear_cache(distribution_type, xcode_managed, platform)
      @profiles[platform].to_h[xcode_managed].to_h[distribution_type] = nil
    end

    def self.fetch_profiles(distribution_type, xcode_managed, platform)
      cached = @profiles[platform].to_h[xcode_managed].to_h[distribution_type]
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

      # filter for type
      distribution_types = {
        'development' => 'limited',
        'app-store' => 'store',
        'ad-hoc' => 'adhoc',
        'enterprise' => 'inhouse'
      }
      profiles = profiles.select do |profile|
        profile.distribution_method == distribution_types[distribution_type]
      end
      # Log.debug("#{distribution_type} profiles (#{profiles.length}):")
      # profiles.each do |profile|
      #   Log.debug("#{profile.name}")
      # end

      platform_profiles = @profiles[platform].to_h
      managed_profiles = platform_profiles[xcode_managed].to_h
      managed_profiles[distribution_type] = profiles

      platform_profiles[xcode_managed] = managed_profiles
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
