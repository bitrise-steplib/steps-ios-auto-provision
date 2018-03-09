# CertificateInfo
class CertificateInfo
  attr_reader :path
  attr_reader :passphrase
  attr_reader :certificate
  attr_accessor :portal_certificate

  def initialize(path, passphrase, certificate)
    @path = path
    @passphrase = passphrase
    @certificate = certificate
  end
end
