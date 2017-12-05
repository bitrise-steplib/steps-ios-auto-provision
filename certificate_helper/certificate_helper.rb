require 'openssl'

def certificate_common_name(certificate)
  common_name = certificate.subject.to_a.find { |name, _, _| name == 'CN' }[1]
  common_name = common_name.force_encoding('UTF-8')
  common_name
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

  match = certificate_infos.find { |info| certificate_common_name(info.certificate) == new_certificate_common_name }
  index = certificate_infos.index(match)

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
