require 'English'

# KeychainHelper
class KeychainHelper
  def self.create_keychain(keychain_path, keychain_password)
    cmd = ['security', '-v', 'create-keychain', '-p', keychain_password, "\"#{keychain_path}\""].join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end

  def initialize(keychain_path, keychain_password)
    if File.file?(keychain_path)
      @keychain_path = keychain_path
      @keychain_password = keychain_password
      return
    end

    new_keychain_path = keychain_path + '-db'
    if File.file?(new_keychain_path)
      @keychain_path = new_keychain_path
      @keychain_password = keychain_password
      return
    end

    KeychainHelper.create_keychain(keychain_path, keychain_password)
    @keychain_path = keychain_path
    @keychain_password = keychain_password
  end

  def install_certificates(certificate_passphrase_map)
    certificate_passphrase_map.each do |path, passphrase|
      import_certificate(path, passphrase)
    end

    set_key_partition_list_if_needed
    set_keychain_settings_default_lock
    add_to_keychain_search_path
    set_default_keychain
    unlock_keychain
  end

  private

  def import_certificate(path, passphrase)
    passphrase = '""' if passphrase.empty?
    cmd = ['security', 'import', "\"#{path}\"", '-k', "\"#{@keychain_path}\"", '-P', passphrase, '-A'].join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end

  def set_key_partition_list_if_needed
    # This is new behavior in Sierra, [openradar](https://openradar.appspot.com/28524119)
    # You need to use "security set-key-partition-list -S apple-tool:,apple: -k keychainPass keychainName" after importing the item and before attempting to use it via codesign.
    cmd = ['sw_vers', '-productVersion'].join(' ')
    Log.debug("$ #{cmd}")
    current_version = `#{cmd}`
    raise "#{cmd} failed, out: #{current_version}" unless $CHILD_STATUS.success?

    return if Gem::Version.new(current_version) < Gem::Version.new('10.12.0')

    cmd = ['security', 'set-key-partition-list', '-S', 'apple-tool:,apple:', '-k', @keychain_password, "\"#{@keychain_path}\""].join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end

  def set_keychain_settings_default_lock
    cmd = ['security', '-v', 'set-keychain-settings', '-lut', '72000', "\"#{@keychain_path}\""].join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end

  def list_keychains
    cmd = ['security', 'list-keychains'].join(' ')
    Log.debug("$ #{cmd}")
    list = `#{cmd}`
    raise "#{cmd} failed, out: #{list}" unless $CHILD_STATUS.success?

    list.split("\n").map(&:strip).map { |e| e.gsub!(/\A"|"\Z/, '') }
  end

  def add_to_keychain_search_path
    keychains = Set.new(list_keychains).add("\"#{@keychain_path}\"").to_a
    cmd = ['security', '-v', 'list-keychains', '-s'].concat(keychains).join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end

  def set_default_keychain
    cmd = ['security', '-v', 'default-keychain', '-s', "\"#{@keychain_path}\""].join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end

  def unlock_keychain
    cmd = ['security', '-v', 'unlock-keychain', '-p', @keychain_password, "\"#{@keychain_path}\""].join(' ')
    Log.debug("$ #{cmd}")
    out = `#{cmd}`
    raise "#{cmd} failed, out: #{out}" unless $CHILD_STATUS.success?
  end
end
