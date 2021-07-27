# Xcode
class Xcode
  attr_accessor :version, :build_version, :major_version

  def initialize
    cmd_params = %w[xcodebuild -version]
    Log.debug("$ #{cmd_params}")
    cmd = cmd_params.join(' ')
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?

    @version, @build_version, @major_version = Xcode.parse_xcodebuild_version_out(out)
  end

  def self.parse_xcodebuild_version_out(out)
    # Xcode 13.0
    # Build version 13A5192j
    lines = out.split("\n")
    raise "unknown xcode version: #{out}" if lines.length < 2

    build_version_line = lines[1]
    build_version_line_components = build_version_line.split(' ')
    raise "unknown xcode build version line: #{build_version_line}" unless build_version_line_components.length == 3

    build_version = build_version_line_components[2]

    version_line = lines[0]
    version_line_components = version_line.split(' ')
    raise "unknown xcode version line: #{version_line}" unless version_line_components.length == 2

    version = version_line_components[1]
    version_components = version.split('.')
    raise "unknown xcode version: #{version}" if version_components.length < 2

    major_version_str = version_components[0]
    begin
      major_version = Integer(major_version_str || '')
    rescue ArgumentError
      raise "unknown xcode major version: #{major_version_str}"
    end

    [version, build_version, major_version]
  end
end
