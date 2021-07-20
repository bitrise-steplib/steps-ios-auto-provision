require 'spaceship'

require_relative 'common'

module Portal
  # DeviceClient ...
  class DeviceClient
    def self.ensure_test_devices(test_devices, platform, device_client = Spaceship::Portal.device)
      valid_devices = []
      if test_devices.to_a.empty?
        Log.success('No test devices registered on Bitrise.')

        valid_devices
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
      registered_devices = fetch_registered_devices(device_client)

      test_devices.each do |test_device|
        registered_device = nil

        registered_devices.each do |device|
          next unless device.udid == test_device.udid

          registered_device = device
          Log.success("test device #{registered_device.name} (#{registered_device.udid}) already registered")
          break
        end

        unless registered_device
          begin
            registered_device = device_client.create!(name: test_device.name, udid: test_device.udid)
          rescue Spaceship::Client::UnexpectedResponse => ex
            message = result_string(ex)
            Log.warn("Failed to register device with name: #{test_device.name} udid: #{test_device.udid} error: #{message}")
            next
          rescue
            Log.warn("Failed to register device with name: #{test_device.name} udid: #{test_device.udid}")
            next
          end

          Log.success("registering test device #{registered_device.name} (#{registered_device.udid})")
        end

        if %i[ios watchos].include?(platform)
          valid_devices = valid_devices.append(test_device) if %w[watch ipad iphone ipod].include?(registered_device.device_type)
        elsif platform == :tvos
          valid_devices = valid_devices.append(test_device) if registered_device.device_type == 'tvOS'
        end
      end

      Log.success("#{valid_devices.length} Bitrise test devices are present on Apple Developer Portal.")
      valid_devices
    end

    def self.fetch_registered_devices(device_client = Spaceship::Portal.device)
      devices = nil
      run_and_handle_portal_function { devices = device_client.all(mac: false, include_disabled: true) || [] }
      devices
    end
  end
end
