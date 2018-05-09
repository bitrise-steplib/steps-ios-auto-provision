# Params
class Params
  attr_accessor :build_url
  attr_accessor :build_api_token
  attr_accessor :team_id
  attr_accessor :certificate_urls_str
  attr_accessor :passphrases_str
  attr_accessor :distribution_type
  attr_accessor :project_path
  attr_accessor :scheme
  attr_accessor :configuration
  attr_accessor :keychain_path
  attr_accessor :keychain_password
  attr_accessor :verbose_log
  attr_accessor :generate_profiles

  attr_accessor :certificate_urls
  attr_accessor :passphrases

  def initialize
    @build_url = ENV['build_url']
    @build_api_token = ENV['build_api_token']
    @team_id = ENV['team_id']
    @certificate_urls_str = ENV['certificate_urls']
    @passphrases_str = ENV['passphrases']
    @distribution_type = ENV['distribution_type']
    @project_path = ENV['project_path']
    @scheme = ENV['scheme']
    @configuration = ENV['configuration']
    @keychain_path = ENV['keychain_path']
    @keychain_password = ENV['keychain_password']
    @verbose_log = ENV['verbose_log']
    @generate_profiles = ENV['generate_profiles']

    @certificate_urls = split_pipe_separated_list(@certificate_urls_str)
    @passphrases = split_pipe_separated_list(@passphrases_str)
  end

  def print
    Log.info('Params:')
    Log.print("team_id: #{@team_id}")
    Log.print("certificate_urls: #{Log.secure_value(@certificate_urls_str)}")
    Log.print("passphrases: #{Log.secure_value(@passphrases_str)}")
    Log.print("distribution_type: #{@distribution_type}")
    Log.print("project_path: #{@project_path}")
    Log.print("scheme: #{@scheme}")
    Log.print("configuration: #{@configuration}")
    Log.print("build_url: #{@build_url}")
    Log.print("build_api_token: #{Log.secure_value(@build_api_token)}")
    Log.print("keychain_path: #{@keychain_path}")
    Log.print("keychain_password: #{Log.secure_value(@keychain_password)}")
    Log.print("verbose_log: #{@verbose_log}")
    Log.print("generate_profiles: #{@generate_profiles}")
  end

  def validate
    raise 'missing: build_url' if @build_url.to_s.empty?
    raise 'missing: build_api_token' if @build_api_token.to_s.empty?
    raise 'missing: certificate_urls' if @certificate_urls_str.to_s.empty?
    raise 'missing: distribution_type' if @distribution_type.to_s.empty?
    raise 'missing: project_path' if @project_path.to_s.empty?
    raise 'missing: scheme' if @scheme.to_s.empty?
    raise 'missing: keychain_path' if @keychain_path.to_s.empty?
    raise 'missing: keychain_password' if @keychain_password.to_s.empty?
    raise 'missing: verbose_log' if @verbose_log.to_s.empty?
    raise 'missing: generate_profiles' if @generate_profiles.to_s.empty?
  end

  private

  def split_pipe_separated_list(list)
    return [''] if list.to_s.empty?
    list.split('|', -1)
  end
end
