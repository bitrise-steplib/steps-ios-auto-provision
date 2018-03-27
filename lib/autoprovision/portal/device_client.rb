require 'spaceship'

require_relative 'common'

module Portal
  # DeviceClient ...
  class DeviceClient
    def self.ensure_test_devices(test_devices, device_client = Spaceship::Portal.device)
      if test_devices.to_a.empty?
        Log.success('no test devices registered on bitrise')
        return
      end

      portal_devices = nil
      run_and_handle_portal_function { portal_devices = device_client.all(mac: false, include_disabled: true) || [] }
      test_devices.each do |test_device|
        registered_test_device = nil

        portal_devices.each do |portal_device|
          next unless portal_device.udid == test_device.udid

          registered_test_device = portal_device
          Log.success("test device #{registered_test_device.name} (#{registered_test_device.udid}) already registered")
          break
        end

        unless registered_test_device
          begin
            registered_test_device = device_client.create!(name: test_device.name, udid: test_device.udid)
          rescue Spaceship::Client::UnexpectedResponse => ex
            message = result_string(ex)
            raise ex unless message
            raise message
          rescue
            raise "Failed to register device with name: #{test_device.name} udid: #{test_device.udid}"
          end

          Log.success("registering test device #{registered_test_device.name} (#{registered_test_device.udid})")
        end

        raise 'failed to find or create device' unless registered_test_device
      end

      Log.success("every test devices (#{test_devices.length}) registered on bitrise are registered on developer portal")
    end
  end
end
