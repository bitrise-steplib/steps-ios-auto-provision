require 'spaceship'

require_relative 'common'

module Portal
  # CertificateClient ...
  class CertificateClient
    def self.download_development_certificates
      development_certificates = []
      run_and_handle_portal_function { development_certificates = Spaceship::Portal.certificate.development.all }

      certificates = []
      development_certificates.each do |cert|
        if cert.can_download
          certificates.push(cert)
        else
          Log.debug("development certificate: #{cert.name} is not downloadable, skipping...")
        end
      end

      certificates
    end

    def self.download_production_certificates
      production_certificates = []
      run_and_handle_portal_function { production_certificates = Spaceship::Portal.certificate.production.all }

      certificates = []
      production_certificates.each do |cert|
        if cert.can_download
          certificates.push(cert)
        else
          Log.debug("production certificate: #{cert.name} is not downloadable, skipping...")
        end
      end

      if production_certificates.to_a.empty?
        run_and_handle_portal_function { production_certificates = Spaceship::Portal.certificate.in_house.all }

        production_certificates.each do |cert|
          if cert.can_download
            certificates.push(cert)
          else
            Log.debug("production certificate: #{cert.name} is not downloadable, skipping...")
          end
        end
      end

      certificates
    end
  end
end
