require_relative 'certificate_info'
require_relative 'utils'
require_relative 'portal/certificate_client'

# CertificateHelper ...
class CertificateHelper
  attr_reader :development_certificate_info
  attr_reader :production_certificate_info

  def initialize
    @all_development_certificate_infos = []
    @all_production_certificate_infos = []

    @portal_certificate_by_id = {}
  end

  def download_and_identify(urls, passes)
    raise "certificates count (#{urls.length}) and passphrases count (#{passes.length}) should match" unless urls.length == passes.length

    certificate_infos = []
    urls.each_with_index do |url, idx|
      Log.debug("downloading certificate ##{idx + 1}")

      path = download_or_create_local_path(url, "Certrificate_#{idx}.p12")
      Log.debug("certificate source: #{path}")

      certificates = read_certificates(path, passes[idx])
      Log.debug("#{certificates.length} codesign identities included:")

      certificates.each do |certificate|
        Log.debug("- #{certificate_name_and_serial(certificate)}")

        if certificate.not_after < Time.now.utc
          Log.error("[X] Certificate is not valid anymore - validity ended at: #{certificate.not_after}\n")
        else
          certificate_info = CertificateInfo.new(path, passes[idx], certificate)
          certificate_infos = append_if_latest_certificate(certificate_info, certificate_infos)
        end
      end
    end
    Log.success("#{urls.length} certificate files downloaded, #{certificate_infos.length} distinct codesign identities included")

    @all_development_certificate_infos, @all_production_certificate_infos = identify_certificate_infos(certificate_infos)
  end

  def ensure_certificate(name, team_id, distribution_type)
    team_development_certificate_infos = map_certificates_infos_by_team_id(@all_development_certificate_infos)[team_id] || []
    team_production_certificate_infos = map_certificates_infos_by_team_id(@all_production_certificate_infos)[team_id] || []

    if team_development_certificate_infos.empty? && team_production_certificate_infos.empty?
      raise "no certificate uploaded for the desired team: #{team_id}"
    end

    if name
      filtered_team_development_certificate_infos = team_development_certificate_infos.select do |certificate_info|
        common_name = certificate_common_name(certificate_info.certificate)
        common_name.downcase.include?(name.downcase)
      end
      team_development_certificate_infos = filtered_team_development_certificate_infos unless filtered_team_development_certificate_infos.empty?

      filtered_team_production_certificate_infos = team_production_certificate_infos.select do |certificate_info|
        common_name = certificate_common_name(certificate_info.certificate)
        common_name.downcase.include?(name.downcase)
      end
      team_production_certificate_infos = filtered_team_production_certificate_infos unless filtered_team_production_certificate_infos.empty?
    end

    if team_development_certificate_infos.length > 1
      msg = "Multiple Development certificates mathes to development team: #{team_id}"
      msg += " and name: #{name}" if name
      Log.warn(msg)
      team_development_certificate_infos.each { |info| Log.warn(" - #{certificate_name_and_serial(info.certificate)}") }
    end

    unless team_development_certificate_infos.empty?
      certificate_info = team_development_certificate_infos[0]
      Log.success("using: #{certificate_name_and_serial(certificate_info.certificate)}")
      @development_certificate_info = certificate_info
    end

    if team_production_certificate_infos.length > 1
      msg = "Multiple Distribution certificates mathes to development team: #{team_id}"
      msg += " and name: #{name}" if name
      Log.warn(msg)
      team_production_certificate_infos.each { |info| Log.warn(" - #{certificate_name_and_serial(info.certificate)}") }
    end

    unless team_production_certificate_infos.empty?
      certificate_info = team_production_certificate_infos[0]
      Log.success("using: #{certificate_name_and_serial(certificate_info.certificate)}")
      @production_certificate_info = certificate_info
    end

    if distribution_type == 'development' && @development_certificate_info.nil?
      raise [
        'Selected distribution type: development, but forgot to provide a Development type certificate.',
        "Don't worry, it's really simple to fix! :)",
        "Simply upload a Development type certificate (.p12) on the workflow editor's CodeSign tab and we'll be building in no time!"
      ].join("\n")
    end

    if distribution_type != 'development' && @production_certificate_info.nil?
      raise [
        "Selected distribution type: #{distribution_type}, but forgot to provide a Distribution type certificate.",
        "Don't worry, it's really simple to fix! :)",
        "Simply upload a Distribution type certificate (.p12) on the workflow editor's CodeSign tab and we'll be building in no time!"
      ].join("\n")
    end
  end

  def certificate_info(distribution_type)
    if distribution_type == 'development'
      @development_certificate_info
    else
      @production_certificate_info
    end
  end

  def identify_certificate_infos(certificate_infos)
    Log.info('Identify Certificates on Developer Portal')

    portal_development_certificates = Portal::CertificateClient.download_development_certificates
    Log.debug('Development certificates on Apple Developer Portal:')
    portal_development_certificates.each do |cert|
      downloaded_portal_cert = download(cert)
      Log.debug("- #{cert.name}: #{certificate_name_and_serial(downloaded_portal_cert)} expire: #{downloaded_portal_cert.not_after}")
    end

    portal_production_certificates = Portal::CertificateClient.download_production_certificates
    Log.debug('Production certificates on Apple Developer Portal:')
    portal_production_certificates.each do |cert|
      downloaded_portal_cert = download(cert)
      Log.debug("- #{cert.name}: #{certificate_name_and_serial(downloaded_portal_cert)} expire: #{downloaded_portal_cert.not_after}")
    end

    development_certificate_infos = []
    production_certificate_infos = []
    certificate_infos.each do |certificate_info|
      Log.debug("searching for Certificate: #{certificate_name_and_serial(certificate_info.certificate)}")
      found = false

      portal_development_certificates.each do |portal_cert|
        downloaded_portal_cert = download(portal_cert)
        next unless certificate_matches(certificate_info.certificate, downloaded_portal_cert)

        Log.success("#{portal_cert.name} certificate found: #{certificate_name_and_serial(certificate_info.certificate)}")
        certificate_info.portal_certificate = portal_cert
        development_certificate_infos.push(certificate_info)
        found = true
        break
      end

      next if found

      portal_production_certificates.each do |portal_cert|
        downloaded_portal_cert = download(portal_cert)
        next unless certificate_matches(certificate_info.certificate, downloaded_portal_cert)

        Log.success("#{portal_cert.name} certificate found: #{certificate_name_and_serial(certificate_info.certificate)}")
        certificate_info.portal_certificate = portal_cert
        production_certificate_infos.push(certificate_info)
      end
    end

    if development_certificate_infos.empty? && production_certificate_infos.empty?
      raise 'no development nor production certificate identified on development portal'
    end

    [development_certificate_infos, production_certificate_infos]
  end

  def download(portal_certificate)
    downloaded_cert = @portal_certificate_by_id[portal_certificate.id]
    unless downloaded_cert
      downloaded_cert = portal_certificate.download
      @portal_certificate_by_id[portal_certificate.id] = downloaded_cert
    end
    downloaded_cert
  end

  def certificate_matches(certificate1, certificate2)
    return true if certificate1.serial == certificate2.serial

    if certificate_common_name(certificate1) == certificate_common_name(certificate2) && certificate1.not_after < certificate2.not_after
      Log.warn([
        "Provided an older version of #{certificate_common_name(certificate1)} certificate (serial: #{certificate1.serial} expire: #{certificate1.not_after}),",
        "please download the most recent version from the Apple Developer Portal (serial: #{certificate2.serial} expire: #{certificate2.not_after}) and use it on Bitrise!"
      ].join("\n"))
    end

    false
  end

  def certificate_team_id(certificate)
    certificate.subject.to_a.find { |name, _, _| name == 'OU' }[1]
  end

  def find_certificate_info_by_identity(identity, certificate_infos)
    certificate_infos.each do |certificate_info|
      common_name = certificate_common_name(certificate_info.certificate)
      return certificate_info if common_name.downcase.include?(identity.downcase)
    end
    nil
  end

  def find_certificate_infos_by_team_id(team_id, certificate_infos)
    matching_certificate_infos = []
    certificate_infos.each do |certificate_info|
      org_unit = certificate_team_id(certificate_info.certificate)
      matching_certificate_infos.push(certificate_info) if org_unit.downcase.include?(team_id.downcase)
    end
    matching_certificate_infos
  end

  def find_matching_codesign_identity_info(identity_name, team_id, certificate_infos)
    if identity_name
      certificate_info = find_certificate_info_by_identity(identity_name, certificate_infos)
      return certificate_info if certificate_info
    end

    team_certificate_infos = find_certificate_infos_by_team_id(team_id, certificate_infos)
    return team_certificate_infos[0] if team_certificate_infos.to_a.length == 1
    Log.print('no development certificate found') if team_certificate_infos.to_a.empty?
    Log.warn("#{team_certificate_infos.length} development certificate found") if team_certificate_infos.to_a.length > 1
  end

  def read_certificates(path, passphrase)
    content = File.read(path)
    p12 = OpenSSL::PKCS12.new(content, passphrase)

    certificates = [p12.certificate]
    certificates.concat(p12.ca_certs) if p12.ca_certs
    certificates
  end

  def append_if_latest_certificate(new_certificate_info, certificate_infos)
    new_certificate_common_name = certificate_common_name(new_certificate_info.certificate)
    index = certificate_infos.index { |info| certificate_common_name(info.certificate) == new_certificate_common_name }
    return certificate_infos.push(new_certificate_info) unless index

    Log.warn("multiple codesign identity uploaded with common name: #{new_certificate_common_name}")

    cert_info = certificate_infos[index]
    certificate_infos[index] = new_certificate_info if new_certificate_info.certificate.not_after > cert_info.certificate.not_after

    certificate_infos
  end

  def map_certificates_infos_by_team_id(certificate_infos)
    map = {}
    certificate_infos.each do |certificate_info|
      team_id = certificate_team_id(certificate_info.certificate)
      infos = map[team_id] || []
      infos.push(certificate_info)
      map[team_id] = infos
    end
    map
  end

  def download_or_create_local_path(url, filename)
    pth = nil
    if url.start_with?('file://')
      pth = url.sub('file://', '')
      raise "Certificate not exist at: #{pth}" unless File.exist?(pth)
    else
      pth = File.join(Dir.tmpdir, filename)
      download_to_path(url, pth)
    end
    pth
  end

  def download_to_path(url, path)
    uri = URI.parse(url)
    request = Net::HTTP::Get.new(uri.request_uri)
    http_object = Net::HTTP.new(uri.host, uri.port)
    http_object.use_ssl = true
    response = http_object.start do |http|
      http.request(request)
    end

    raise printable_response(response) unless response.code == '200'

    File.open(path, 'wb') do |file|
      file.write(response.body)
    end

    content = File.read(path)
    raise 'empty file' if content.to_s.empty?

    path
  end
end
