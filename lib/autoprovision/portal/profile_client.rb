require 'spaceship'

require_relative 'common'
require_relative 'app_client'

module Portal
  # ProfileClient ...
  class ProfileClient
    def self.ensure_xcode_managed_profile(bundle_id, entitlements, distribution_type)
      profile_class = portal_profile_class(distribution_type)
      profiles = profile_class.all(mac: false, xcode: true)
      xcode_managed_profiles = profiles.select(&:managed_by_xcode?)
      matching_profiles = xcode_managed_profiles.select { |profile| profile.app.bundle_id == bundle_id }

      if matching_profiles.empty?
        matching_profiles = xcode_managed_profiles.select do |profile|
          next unless File.fnmatch(profile.app.bundle_id, bundle_id)
          next unless AppClient.all_services_enabled?(profile.app, entitlements)
          true
        end
      end

      raise "failed to find Xcode managed provisioning profile for bundle id: #{bundle_id}" if matching_profiles.empty?
      matching_profiles[0]
    end

    def self.ensure_manual_profile(certificate, app, distribution_type, allow_retry = true)
      profile_class = portal_profile_class(distribution_type)

      profiles = nil
      profile_name = "Bitrise #{distribution_type} - (#{app.bundle_id})"
      run_and_handle_portal_function { profiles = profile_class.all.select { |profile| profile.app.bundle_id == app.bundle_id && profile.name == profile_name } }
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
        Log.debug("generating #{distribution_type} profile: #{profile_name}")
      else
        # it's easier to just create a new one, than to:
        # - add test devices
        # - add the certificate
        # - update profile
        # update seems to revoking the certificate, even if it is not neccessary
        # it has the same effects anyway, including a new UUID of the provisioning profile
        if profiles.count > 1
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
        run_and_handle_portal_function { profile = profile_class.create!(bundle_id: app.bundle_id, certificate: certificate, name: profile_name) }
      rescue => ex
        # Failed to remove already existing managed profile, try it again!
        raise ex unless allow_retry
        raise ex unless ex.to_s =~ /Multiple profiles found with the name '(.*)'.\s*Please remove the duplicate profiles and try again./i

        Log.debug(ex.to_s)
        Log.debug('failed to regenerate the profile, retrying in 5 sec ...')
        sleep(5)
        ensure_provisioning_profile(certificate, app, distribution_type, false)
      end

      raise "failed to find or create provisioning profile for bundle id: #{app.bundle_id}" unless profile
      profile
    end

    private_class_method

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
