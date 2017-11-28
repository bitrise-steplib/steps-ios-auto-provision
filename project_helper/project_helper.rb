require 'xcodeproj'
require 'json'
require 'plist'
require 'English'

# ProjectHelper ...
class ProjectHelper
  def initialize(project_or_workspace_path, scheme)
    extname = File.extname(project_or_workspace_path)
    case extname
    when '.xcodeproj'
      @project_path = project_or_workspace_path
    when '.xcworkspace'
      @workspace_path = project_or_workspace_path
    else
      raise "unkown project extension: #{extname}, should be: .xcodeproj or .xcworkspace"
    end

    @scheme_name = scheme
  end

  def application_targets
    result = read_application_targets
    result[0].collect(&:name)
  end

  def archive_action_configuration
    result = read_scheme
    scheme = result[0]
    archive_action = scheme.archive_action
    return nil unless archive_action

    archive_action.build_configuration
  end

  def project_codesign_identity(configuration)
    codesign_identity = nil

    result = read_application_targets
    targets = result[0]
    targets_project_path = result[1]

    targets.each do |target_name|
      target_identity = target_codesign_identity(targets_project_path, target_name, configuration)
      Log.debug("#{target_name} codesign identity: #{target_identity} ")

      if target_identity.to_s.empty?
        Log.warn("no CODE_SIGN_IDENTITY build settings found for target: #{target_name}")
        next
      end

      if codesign_identity.nil?
        codesign_identity = target_identity
        next
      end

      unless codesign_identites_match?(codesign_identity, target_identity)
        Log.warn("target codesign identity: #{target_identity} does not match to the already registered codesign identity: #{codesign_identity}")
        codesign_identity = nil
        break
      end

      codesign_identity = exact_codesign_identity(codesign_identity, target_identity)
    end

    codesign_identity
  end

  def project_team_id(configuration)
    team_id = nil

    result = read_application_targets
    targets = result[0]
    targets_project_path = result[1]

    targets.each do |target_name|
      id = target_team_id(targets_project_path, target_name, configuration)
      Log.debug("#{target_name} team id: #{id} ")

      if id.to_s.empty?
        Log.warn("no DEVELOPMENT_TEAM build settings found for target: #{target_name}")
        next
      end

      if team_id.nil?
        team_id = id
        next
      end

      next if team_id == id

      Log.warn("target team id: #{id} does not match to the already registered team id: #{team_id}")
      team_id = nil
      break
    end

    team_id
  end

  def target_bundle_id(target_name, configuration)
    result = read_application_targets
    targets_project_path = result[1]

    settings = xcodebuild_target_build_settings(targets_project_path, target_name, configuration)
    find_bundle_id(settings, targets_project_path)
  end

  def target_entitlements(target_name, configuration)
    result = read_application_targets
    targets_project_path = result[1]

    settings = xcodebuild_target_build_settings(targets_project_path, target_name, configuration)
    entitlements_path = settings['CODE_SIGN_ENTITLEMENTS']
    return nil if entitlements_path.to_s.empty?

    project_dir = File.dirname(targets_project_path)
    entitlements_path = File.join(project_dir, entitlements_path)
    Plist.parse_xml(entitlements_path)
  end

  def force_code_sign_properties(target_name, configuration, development_team, code_sign_identity, provisioning_profile_uuid)
    result = read_application_targets
    targets_project_path = result[1]

    target_found = false
    configuration_found = false

    project = Xcodeproj::Project.open(targets_project_path)
    project.targets.each do |target_obj|
      next unless target_obj.name == target_name
      target_found = true

      # force manual code singing
      target_id = target_obj.uuid
      attributes = project.root_object.attributes['TargetAttributes']
      target_attributes = attributes[target_id]
      target_attributes['ProvisioningStyle'] = 'Manual'

      # apply code sign properties
      target_obj.build_configuration_list.build_configurations.each do |build_configuration|
        next unless build_configuration.name == configuration
        configuration_found = true

        build_settings = build_configuration.build_settings

        build_settings['CODE_SIGN_STYLE'] = 'Manual'
        Log.print('CODE_SIGN_STYLE: Manual')

        build_settings['DEVELOPMENT_TEAM'] = development_team
        Log.print("DEVELOPMENT_TEAM: #{development_team}")

        build_settings['PROVISIONING_PROFILE'] = provisioning_profile_uuid
        Log.print("PROVISIONING_PROFILE: #{provisioning_profile_uuid}")

        build_settings['PROVISIONING_PROFILE_SPECIFIER'] = ''
        Log.print('PROVISIONING_PROFILE_SPECIFIER: \'\'')

        # code sign identity may presents as: CODE_SIGN_IDENTITY and CODE_SIGN_IDENTITY[sdk=iphoneos*]
        build_settings.each_key do |key|
          next unless key.include?('CODE_SIGN_IDENTITY')

          build_settings[key] = code_sign_identity
          Log.print("#{key}: #{code_sign_identity}")
        end
      end
    end

    raise "target (#{target_name}) not found in project: #{targets_project_path}" unless target_found
    raise "configuration (#{configuration}) does not exist in project: #{targets_project_path}" unless configuration_found

    project.save
  end

  private

  def read_application_targets
    result = read_scheme
    scheme = result[0]
    scheme_container_project = result[1]

    result = read_scheme_archiveable_target(scheme, scheme_container_project)
    target = result[0]
    target_container_project = result[1]

    targets = []
    targets = collect_dependent_targets(target, targets)
    raise 'failed to collect scheme targets' if targets.empty?

    [targets, target_container_project]
  end

  def target_codesign_identity(project_pth, target_name, configuration)
    settings = xcodebuild_target_build_settings(project_pth, target_name, configuration)
    settings['CODE_SIGN_IDENTITY']
  end

  def target_team_id(project_pth, target_name, configuration)
    settings = xcodebuild_target_build_settings(project_pth, target_name, configuration)
    settings['DEVELOPMENT_TEAM']
  end

  def shared_scheme_path(project_or_workspace_pth, scheme_name)
    File.join(project_or_workspace_pth, 'xcshareddata', 'xcschemes', scheme_name + '.xcscheme')
  end

  def user_scheme_path(project_or_workspace_pth, scheme_name)
    user_name = ENV['user']
    File.join(project_or_workspace_pth, 'xcuserdata', user_name + '.xcuserdatad', 'xcschemes', scheme_name + '.xcscheme')
  end

  def read_scheme
    project_paths = [@project_path]
    if File.extname(@project_path) == '.xcworkspace'
      project_paths += workspace_contained_projects(@project_path)
    end

    project_paths.each do |project_path|
      scheme_pth = shared_scheme_path(project_path, @scheme_name)
      scheme_pth = user_scheme_path(project_path, @scheme_name) unless File.exist?(scheme_pth)
      next unless File.exist?(scheme_pth)

      return [Xcodeproj::XCScheme.new(scheme_pth), project_path]
    end

    raise "project (#{@project}) does not contain scheme: #{@scheme_name}"
  end

  def read_scheme_archiveable_target(scheme, project_path)
    build_action = scheme.build_action
    return nil unless build_action

    entries = build_action.entries || []
    return nil if entries.empty?

    entries = entries.select(&:build_for_archiving?) || []
    return nil if entries.empty?

    entries.each do |entry|
      buildable_references = entry.buildable_references || []
      next if buildable_references.empty?

      buildable_references = buildable_references.reject do |r|
        r.target_name.to_s.empty? || r.target_referenced_container.to_s.empty?
      end
      next if buildable_references.empty?

      buildable_reference = entry.buildable_references.first

      target_name = buildable_reference.target_name.to_s
      container = buildable_reference.target_referenced_container.to_s.sub(/^container:/, '')
      next if target_name.empty? || container.empty?

      project_dir = File.dirname(project_path)
      target_project_pth = File.expand_path(container, project_dir)
      next unless File.exist?(target_project_pth)

      project = Xcodeproj::Project.open(target_project_pth)
      next unless project

      target = project.targets.find { |t| t.name == target_name }
      next unless target
      next unless runnable_target?(target)

      return [target, target_project_pth]
    end

    raise 'failed to find scheme archivable target'
  end

  def collect_dependent_targets(target, dependent_targets)
    dependent_targets << target

    dependencies = target.dependencies || []
    return dependent_targets if dependencies.empty?

    dependencies.each do |dependency|
      dependent_target = dependency.target
      next unless dependent_target
      next unless runnable_target?(dependent_target)

      collect_dependent_targets(dependent_target, dependent_targets)
    end

    dependent_targets
  end

  def contained_projects
    return [@project_path] unless @workspace_path

    workspace = Xcodeproj::Workspace.new_from_xcworkspace(@workspace_path)
    workspace_dir = File.dirname(@workspace_path)
    project_paths = []
    workspace.file_references.each do |ref|
      pth = ref.path
      next unless File.extname(pth) == '.xcodeproj'
      next if pth.end_with?('Pods/Pods.xcodeproj')

      project_path = File.expand_path(pth, workspace_dir)
      project_paths << project_path
    end

    project_paths
  end

  def runnable_target?(target)
    return false unless target.is_a?(Xcodeproj::Project::Object::PBXNativeTarget)

    product_reference = target.product_reference
    return false unless product_reference

    product_reference.path.end_with?('.app', '.appex')
  end

  def project_targets_map
    project_targets = {}

    project_paths = contained_projects
    project_paths.each do |project_path|
      targets = []

      project = Xcodeproj::Project.open(project_path)
      project.targets.each do |target|
        next unless runnable_target?(target)

        targets.push(target.name)
      end

      project_targets[project_path] = targets
    end

    project_targets
  end

  def xcodebuild_target_build_settings(project, target, configuration)
    cmd = "xcodebuild -showBuildSettings -project \"#{project}\" -target \"#{target}\" -configuration \"#{configuration}\""
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?

    settings = {}
    lines = out.split(/\n/)
    lines.each do |line|
      line = line.strip
      next unless line.include?(' = ')

      split = line.split(' = ')
      next unless split.length == 2

      value = split[1].strip
      next if value.empty?

      key = split[0].strip
      next if key.empty?

      settings[key] = value
    end

    settings
  end

  def resolve_bundle_id(bundle_id, build_settings)
    pattern = /(.*)\$\((.*)\)(.*)/
    matches = bundle_id.match(pattern)
    raise "failed to resolve bundle id (#{bundle_id}): does not conforms to: /(.*)$\(.*\)(.*)/" unless matches

    captures = matches.captures
    raise "failed to resolve bundle id (#{bundle_id}): does not conforms to: /(.*)$\(.*\)(.*)/" if captures.to_a.length != 3

    prefix = captures[0]
    suffix = captures[2]
    env_key = captures[1]
    split = env_key.split(':')
    env_key = split[0] if split.length > 1

    env_value = build_settings[env_key]
    raise "failed to resolve bundle id (#{bundle_id}): build settings not found with key: (#{env_key})" if env_value.to_s.empty?

    prefix + env_value + suffix
  end

  def find_bundle_id(build_settings, project_path)
    bundle_id = build_settings['PRODUCT_BUNDLE_IDENTIFIER']
    return bundle_id if bundle_id

    info_plist_path = build_settings['INFOPLIST_FILE']
    raise 'failed to to determine bundle id: xcodebuild -showBuildSettings does not contains PRODUCT_BUNDLE_IDENTIFIER nor INFOPLIST_FILE' unless info_plist_path

    info_plist_path = File.expand_path(info_plist_path, File.dirname(project_path))
    info_plist = Plist.parse_xml(info_plist_path)
    bundle_id = info_plist['CFBundleIdentifier']
    raise 'failed to to determine bundle id: xcodebuild -showBuildSettings does not contains PRODUCT_BUNDLE_IDENTIFIER nor Info.plist' if bundle_id.to_s.empty?

    return bundle_id unless bundle_id.to_s.include?('$')

    Log.warn("CFBundleIdentifier defined with build settings variable: #{bundle_id}, trying to resolve it...")
    resolve_bundle_id(bundle_id, build_settings)
  end

  # 'iPhone Developer' should match to 'iPhone Developer: Bitrise Bot (ABCD)'
  def codesign_identites_match?(identity1, identity2)
    return true if identity1.downcase.include?(identity2.downcase)
    return true if identity2.downcase.include?(identity1.downcase)
    false
  end

  # 'iPhone Developer: Bitrise Bot (ABCD)' is exact compared to 'iPhone Developer'
  def exact_codesign_identity(identity1, identity2)
    return nil unless codesign_identites_match?(identity1, identity2)
    identity1.length > identity2.length ? identity1 : identity2
  end
end
