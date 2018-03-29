def certificate_common_name(certificate)
  common_name = certificate.subject.to_a.find { |name, _, _| name == 'CN' }[1]
  common_name = common_name.force_encoding('UTF-8')
  common_name
end

private

def certificate_name_and_serial(certificate)
  "#{certificate_common_name(certificate)} [#{certificate.serial}]"
end
