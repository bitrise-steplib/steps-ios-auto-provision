# CertificateInfo
class CertificateInfo
  attr_reader :path, :passphrase, :certificate
  attr_accessor :portal_certificate

  def initialize(path, passphrase, certificate)
    @path = path
    @passphrase = passphrase
    @certificate = certificate
  end
end
