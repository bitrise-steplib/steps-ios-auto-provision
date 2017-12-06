require 'xcodeproj'
require 'json'
require 'plist'
require 'English'

# ProjectHelper ...
class ProjectHelper
  attr_accessor :targets

  def initialize(project_or_workspace_path, scheme_name, configuration_name)
    raise "project not exist at: #{project_or_workspace_path}" unless File.exist?(project_or_workspace_path)

    extname = File.extname(project_or_workspace_path)
    case extname
    when '.xcodeproj'
      @project_path = project_or_workspace_path
    when '.xcworkspace'
      @workspace_path = project_or_workspace_path
    else
      raise "unkown project extension: #{extname}, should be: .xcodeproj or .xcworkspace"
    end

    # ensure scheme exist
    scheme, scheme_container_project_path = read_scheme_and_container_project(scheme_name)

    # read scheme application targets
    target, @targets_container_project_path = read_scheme_archivable_target_and_container_project(scheme, scheme_container_project_path)
    @targets = collect_dependent_targets(target)
    raise 'failed to collect scheme targets' if @targets.empty?

    # ensure configuration exist
    action = scheme.archive_action
    raise "archive action not defined for scheme: #{scheme_name}" unless action
    default_configuration_name = action.build_configuration
    raise "archive action's configuration not found for scheme: #{scheme_name}" unless default_configuration_name

    if configuration_name.empty?
      @configuration_name = default_configuration_name
    elsif configuration_name != default_configuration_name
      targets.each do |target_obj|
        configurations = target_obj.build_configuration_list.build_configurations.find { |c| configuration_name.to_s == c.name }
        raise "build configuration (#{configuration_name}) not defined for target: #{target.name}" if configurations.to_a.empty?
      end

      Log.warn("Using defined build configuration: #{configuration_name} instead of the scheme's default one: #{default_configuration_name}")
      @configuration_name = configuration_name
    end
  end

  def project_codesign_identity
    codesign_identity = nil

    @targets.each do |target_name|
      target_identity = target_codesign_identity(@targets_container_project_path, target_name, @configuration_name)
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

  def project_team_id
    team_id = nil

    @targets.each do |target_name|
      id = target_team_id(@targets_container_project_path, target_name, @configuration_name)
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

  def target_bundle_id(target_name)
    build_settings = xcodebuild_target_build_settings(@targets_container_project_path, target_name, @configuration_name)

    bundle_id = build_settings['PRODUCT_BUNDLE_IDENTIFIER']
    return bundle_id if bundle_id

    Log.debug("PRODUCT_BUNDLE_IDENTIFIER env not found in 'xcodebuild -showBuildSettings -project \"#{@targets_container_project_path}\" -target \"#{target_name}\" -configuration \"#{@configuration_name}\"' command's output")
    Log.debug("build settings:\n#{build_settings}")
    Log.debug("checking the Info.plist file's CFBundleIdentifier property...")

    info_plist_path = build_settings['INFOPLIST_FILE']
    raise 'failed to to determine bundle id: xcodebuild -showBuildSettings does not contains PRODUCT_BUNDLE_IDENTIFIER nor INFOPLIST_FILE' unless info_plist_path

    info_plist_path = File.expand_path(info_plist_path, File.dirname(project_path))
    info_plist = Plist.parse_xml(info_plist_path)
    bundle_id = info_plist['CFBundleIdentifier']
    raise 'failed to to determine bundle id: xcodebuild -showBuildSettings does not contains PRODUCT_BUNDLE_IDENTIFIER nor Info.plist' if bundle_id.to_s.empty?

    return bundle_id unless bundle_id.to_s.include?('$')

    Log.warn("CFBundleIdentifier defined with variable: #{bundle_id}, trying to resolve it...")
    resolve_bundle_id(bundle_id, build_settings)
  end

  def target_entitlements(target_name)
    settings = xcodebuild_target_build_settings(@targets_container_project_path, target_name, @configuration_name)
    entitlements_path = settings['CODE_SIGN_ENTITLEMENTS']
    return if entitlements_path.to_s.empty?

    project_dir = File.dirname(@targets_container_project_path)
    entitlements_path = File.join(project_dir, entitlements_path)
    Plist.parse_xml(entitlements_path)
  end

  def force_code_sign_properties(target_name, development_team, code_sign_identity, provisioning_profile_uuid)
    target_found = false
    configuration_found = false

    project = Xcodeproj::Project.open(@targets_container_project_path)
    project.targets.each do |target_obj|
      next unless target_obj.name == target_name
      target_found = true

      # force manual code signing
      target_id = target_obj.uuid
      attributes = project.root_object.attributes['TargetAttributes']
      target_attributes = attributes[target_id]
      target_attributes['ProvisioningStyle'] = 'Manual'

      # apply code sign properties
      target_obj.build_configuration_list.build_configurations.each do |build_configuration|
        next unless build_configuration.name == @configuration_name
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

    raise "target (#{target_name}) not found in project: #{@targets_container_project_path}" unless target_found
    raise "configuration (#{@configuration_name}) does not exist in project: #{@targets_container_project_path}" unless configuration_found

    project.save
  end

  private

  def read_scheme_and_container_project(scheme_name)
    project_paths = [@project_path]
    if File.extname(@project_path) == '.xcworkspace'
      project_paths += contained_projects
    end

    project_paths.each do |project_path|
      schema_path = shared_scheme_path(project_path, scheme_name)
      schema_path = user_scheme_path(project_path, scheme_name) unless File.exist?(schema_path)
      next unless File.exist?(schema_path)

      return Xcodeproj::XCScheme.new(schema_path), project_path
    end

    raise "project (#{@project_path}) does not contain scheme: #{scheme_name}"
  end

  def archivable_target_and_container_project(buildable_references, scheme_container_project_dir)
    buildable_references.each do |reference|
      next if reference.target_name.to_s.empty?
      next if reference.target_referenced_container.to_s.empty?

      container = reference.target_referenced_container.sub(/^container:/, '')
      next if container.empty?

      target_project_path = File.expand_path(container, scheme_container_project_dir)
      next unless File.exist?(target_project_path)

      project = Xcodeproj::Project.open(target_project_path)
      target = project.targets.find { |t| t.name == reference.target_name }
      next unless target
      next unless runnable_target?(target)

      return target, target_project_path
    end
  end

  def read_scheme_archivable_target_and_container_project(scheme, scheme_container_project_path)
    build_action = scheme.build_action
    return nil unless build_action

    entries = build_action.entries || []
    return nil if entries.empty?

    entries = entries.select(&:build_for_archiving?) || []
    return nil if entries.empty?

    scheme_container_project_dir = File.dirname(scheme_container_project_path)

    entries.each do |entry|
      buildable_references = entry.buildable_references || []
      next if buildable_references.empty?

      target, target_project_path = archivable_target_and_container_project(buildable_references, scheme_container_project_dir)
      next if target.nil? || target_project_path.nil?

      return target, target_project_path
    end

    raise 'failed to find scheme archivable target'
  end

  def collect_dependent_targets(target, dependent_targets = [])
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
    user_name = ENV['USER']
    File.join(project_or_workspace_pth, 'xcuserdata', user_name + '.xcuserdatad', 'xcschemes', scheme_name + '.xcscheme')
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
    # Bitrise.$(PRODUCT_NAME:rfc1034identifier)
    pattern = /(.*)\$\((.*)\)(.*)/
    matches = bundle_id.match(pattern)
    raise "failed to resolve bundle id (#{bundle_id}): does not conforms to: /(.*)$\(.*\)(.*)/" unless matches

    captures = matches.captures
    prefix = captures[0]
    suffix = captures[2]
    env_key = captures[1]
    split = env_key.split(':')
    raise "failed to resolve bundle id (#{bundle_id}): failed to determine settings key" if split.empty?

    env_key = split[0]
    env_value = build_settings[env_key]
    raise "failed to resolve bundle id (#{bundle_id}): build settings not found with key: (#{env_key})" if env_value.to_s.empty?

    prefix + env_value + suffix
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
