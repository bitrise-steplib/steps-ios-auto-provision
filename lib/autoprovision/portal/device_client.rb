require 'spaceship'

require_relative 'common'

module Portal
  # DeviceClient ...
  class DeviceClient
    def self.ensure_test_devices(register_test_devices, test_devices, platform, device_client = Spaceship::Portal.device)
      Log.info('Fetching Apple Developer Portal devices')
      dev_portal_devices = fetch_registered_devices(device_client)

      Log.print("#{dev_portal_devices.length} devices are registered on the Apple Developer Portal")
      dev_portal_devices.each do |dev_portal_device|
        Log.debug("- #{dev_portal_device.name}, #{dev_portal_device.device_type}, UDID (#{dev_portal_device.udid})")
      end

      if register_test_devices && !test_devices.empty?
        unique_test_devices = Device.filter_duplicated_devices(test_devices)

        Log.info("Checking if #{unique_test_devices.length} Bitrise test device(s) are registered on Developer Portal")
        unique_test_devices.each do |test_device|
          Log.debug("- #{test_device.name}, UDID (#{test_device.udid})")
        end

        duplicated_devices_groups = Device.duplicated_device_groups(test_devices)
        unless duplicated_devices_groups.to_a.empty?
          Log.warn('Devices with duplicated UDID are registered on Bitrise, will be ignored:')
          duplicated_devices_groups.each do |duplicated_devices|
            Log.warn("- #{duplicated_devices.map(&:udid).join(' - ')}")
          end
        end

        new_dev_portal_devices = register_missing_test_devices(device_client, unique_test_devices, dev_portal_devices)
        dev_portal_devices = dev_portal_devices.concat(new_dev_portal_devices)
      end

      filter_dev_portal_devices(dev_portal_devices, platform)
    end

    def self.filter_dev_portal_devices(dev_portal_devices, platform)
      filtered_devices = []
      dev_portal_devices.each do |dev_portal_device|
        if %i[ios watchos].include?(platform)
          filtered_devices = filtered_devices.append(dev_portal_device) if %w[watch ipad iphone ipod].include?(dev_portal_device.device_type)
        elsif platform == :tvos
          filtered_devices = filtered_devices.append(dev_portal_device) if dev_portal_device.device_type == 'tvOS'
        end
      end
      filtered_devices
    end

    def self.find_dev_portal_device(test_device, dev_portal_devices)
      device = nil
      dev_portal_devices.each do |dev_portal_device|
        if test_device.udid == dev_portal_device.udid
          device = dev_portal_device
          break
        end
      end
      device
    end

    def self.register_missing_test_devices(device_client = Spaceship::Portal.device, test_devices, dev_portal_devices)
      new_dev_portal_devices = []

      test_devices.each do |test_device|
        Log.print("checking if the device (#{test_device.udid}) is registered")

        dev_portal_device = find_dev_portal_device(test_device, dev_portal_devices)
        unless dev_portal_device.nil?
          Log.print('device already registered')
          next
        end

        Log.print('registering device')
        new_dev_portal_device = register_test_device_on_dev_portal(device_client, test_device)
        new_dev_portal_devices.append(new_dev_portal_device) unless new_dev_portal_device.nil?
      end

      new_dev_portal_devices
    end

    def self.register_test_device_on_dev_portal(device_client = Spaceship::Portal.device, test_device)
      device_client.create!(name: test_device.name, udid: test_device.udid)
    rescue Spaceship::Client::UnexpectedResponse => ex
      message = preferred_error_message(ex)
      Log.warn("Failed to register device with name: #{test_device.name} udid: #{test_device.udid} error: #{message}")
      nil
    rescue
      Log.warn("Failed to register device with name: #{test_device.name} udid: #{test_device.udid}")
      nil
    end

    def self.fetch_registered_devices(device_client = Spaceship::Portal.device)
      devices = nil
      run_or_raise_preferred_error_message { devices = device_client.all(mac: false, include_disabled: false) || [] }
      devices
    end
  end
end
