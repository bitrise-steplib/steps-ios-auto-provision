require 'spaceship'

require_relative 'common'

module Portal
  # DeviceClient ...
  class DeviceClient
    def self.ensure_test_devices(test_devices, device_client = Spaceship::Portal.device)
      if test_devices.to_a.empty?
        Log.success('No test devices registered on Bitrise.')
        return
      end

      # Log the duplicated devices (by udid)
      duplicated_devices_groups = Device.duplicated_device_groups(test_devices)
      unless duplicated_devices_groups.to_a.empty?
        Log.debug('Devices registered multiples times on Bitrise:')

        duplicated_devices_groups.each do |duplicated_devices|
          Log.debug("#{duplicated_devices.map(&:udid).join("\n")}\n\n")
        end
      end

      # Remove the duplications from the device list
      test_devices = Device.filter_duplicated_devices(test_devices)
      portal_devices = fetch_devices(device_client)

      new_device_registered = false
      valid_devices = []
      test_devices.each do |test_device|
        registered_test_device = nil

        portal_devices.each do |portal_device|
          next unless portal_device.udid == test_device.udid

          registered_test_device = portal_device
          Log.success("test device #{registered_test_device.name} (#{registered_test_device.udid}) already registered")
          break
        end

        unless registered_test_device
          new_device_registered = true
          begin
            registered_test_device = device_client.create!(name: test_device.name, udid: test_device.udid)
          rescue Spaceship::Client::UnexpectedResponse => ex
            message = result_string(ex)
            Log.warn("Failed to register device with name: #{test_device.name} udid: #{test_device.udid} error: #{message}")
            next
          rescue
            Log.warn("Failed to register device with name: #{test_device.name} udid: #{test_device.udid}")
            next
          end

          Log.success("registering test device #{registered_test_device.name} (#{registered_test_device.udid})")
        end

        valid_devices = valid_devices.append(test_device)
        raise 'failed to find or create device' unless registered_test_device
      end

      Log.success("#{valid_devices.length} Bitrise test devices are present on Apple Developer Portal.")
      valid_devices
    end

    def self.fetch_devices(device_client = Spaceship::Portal.device)
      portal_devices = nil
      run_and_handle_portal_function { portal_devices = device_client.all(mac: false, include_disabled: true) || [] }
      portal_devices
    end
  end
end
