require_relative 'profile_info'
require_relative 'portal/profile_client'
require_relative 'portal/app_client'

# ProfileHelper ...
class ProfileHelper
  def initialize(project_helper, certificate_helper)
    @project_helper = project_helper
    @certificate_helper = certificate_helper

    @profiles = {}
  end

  def ensure_profiles(distribution_type, generate_profiles = false)
    distribution_types = [distribution_type]
    if distribution_type != 'development' && @certificate_helper.certificate_info('development')
      distribution_types = ['development'].concat(distribution_types)
    end

    Log.debug("distribution_types: #{distribution_types}")

    if @project_helper.uses_xcode_auto_codesigning? && generate_profiles
      Log.warn('project uses Xcode managed signing, but generate_profiles set to true, trying to generate Provisioning Profiles')

      begin
        distribution_types.each { |distr_type| ensure_manual_profiles(distr_type, @project_helper.platform) }
      rescue => ex
        Log.error('generate_profiles set to true, but failed to generate Provisioning Profiles with error:')
        Log.error(ex.to_s)
        Log.info("\nTrying to use Xcode managed Provisioning Profiles")

        ensure_profiles(distribution_type, false)
      end

      return false
    end

    distribution_types.each do |distr_type|
      if @project_helper.uses_xcode_auto_codesigning?
        ensure_xcode_managed_profiles(distr_type, @project_helper.platform)
      else
        ensure_manual_profiles(distr_type, @project_helper.platform)
      end
    end

    @project_helper.uses_xcode_auto_codesigning?
  end

  def profiles_by_bundle_id(distribution_type)
    @profiles[distribution_type]
  end

  private

  def ensure_xcode_managed_profiles(distribution_type, platform)
    certificate = @certificate_helper.certificate_info(distribution_type).portal_certificate

    targets = @project_helper.targets
    targets.each do |target|
      target_name = target.name
      bundle_id = @project_helper.target_bundle_id(target_name)
      entitlements = @project_helper.target_entitlements(target_name) || {}

      Log.print("checking xcode managed #{distribution_type} profile for target: #{target_name} (#{bundle_id}) with #{entitlements.length} services on developer portal")
      portal_profile = Portal::ProfileClient.ensure_xcode_managed_profile(bundle_id, entitlements, distribution_type, certificate, platform)

      Log.print("downloading #{distribution_type} profile: #{portal_profile.name}")
      profile_path = write_profile(portal_profile)
      Log.debug("profile path: #{profile_path}")

      profile_info = ProfileInfo.new(profile_path, portal_profile)
      @profiles[distribution_type] ||= {}
      @profiles[distribution_type][bundle_id] = profile_info
    end
  end

  def ensure_manual_profiles(distribution_type, platform)
    certificate = @certificate_helper.certificate_info(distribution_type).portal_certificate

    targets = @project_helper.targets
    targets.each do |target|
      target_name = target.name
      bundle_id = @project_helper.target_bundle_id(target_name)
      entitlements = @project_helper.target_entitlements(target_name) || {}

      Log.print("checking app for target: #{target_name} (#{bundle_id}) on developer portal")
      app = Portal::AppClient.ensure_app(bundle_id)

      Log.debug('sync app services')
      app = Portal::AppClient.sync_app_services(app, entitlements)

      Log.print("ensure #{distribution_type} profile for target: #{target_name} on developer portal")
      portal_profile = Portal::ProfileClient.ensure_manual_profile(certificate, app, entitlements, distribution_type, platform)

      Log.print("downloading #{distribution_type} profile: #{portal_profile.name}")
      profile_path = write_profile(portal_profile)
      Log.debug("profile path: #{profile_path}")

      profile_info = ProfileInfo.new(profile_path, portal_profile)
      @profiles[distribution_type] ||= {}
      @profiles[distribution_type][bundle_id] = profile_info
    end
  end

  def write_profile(profile)
    home_dir = ENV['HOME']
    raise 'failed to determine xcode provisioning profiles dir: $HOME not set' if home_dir.to_s.empty?

    profiles_dir = File.join(home_dir, 'Library/MobileDevice/Provisioning Profiles')
    FileUtils.mkdir_p(profiles_dir) unless File.directory?(profiles_dir)

    profile_path = File.join(profiles_dir, profile.uuid + '.mobileprovision')
    Log.warn("profile already exists at: #{profile_path}, overwriting...") if File.file?(profile_path)

    File.write(profile_path, profile.download)
    profile_path
  end
end
