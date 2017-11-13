require 'net/http'
require 'uri'
require 'tmpdir'
require 'fileutils'
require_relative '../log/log'

def create_tmp_file(filename)
  File.join(Dir.tmpdir, filename)
end

def printable_response(response)
  str = "response\n"
  str += "status: #{response.code}\n"
  str += "body: #{response.body}\n"
  str
end

def download_to_path(url, path)
  uri = URI.parse(url)
  request = Net::HTTP::Get.new(uri.request_uri)
  http_object = Net::HTTP.new(uri.host, uri.port)
  http_object.use_ssl = (uri.scheme == 'https')
  response = http_object.start do |http|
    http.request(request)
  end

  raise printable_response(response) unless response.code == '200'

  open(path, 'wb') do |file|
    file.write(response.body)
  end

  content = File.read(path)
  raise 'empty file' if content.to_s.empty?

  path
end

def download_to_tmp_file(url, filename)
  pth = nil
  if url.start_with?('file://')
    pth = url.sub('file://', '')
    raise "Certificate not exist at: #{pth}" unless File.exist?(pth)
  else
    pth = create_tmp_file(filename)
    download_to_path(url, pth)
  end
  pth
end

def download_profile(profile)
  home_dir = ENV['HOME']
  raise 'failed to determine Xcode Provisioning Profiles dir: HOME env not set' if home_dir.to_s.empty?

  profiles_dir = File.join(home_dir, 'Library/MobileDevice/Provisioning Profiles')
  FileUtils.mkdir_p(profiles_dir) unless File.directory?(profiles_dir)

  profile_path = File.join(profiles_dir, profile.uuid + '.mobileprovision')
  Log.warn("Provisioning Profile already exists at: #{profile_path}, overwriting...") if File.file?(profile_path)

  File.write(profile_path, profile.download)
  profile_path
end
