# ProfileInfo
class ProfileInfo
  attr_reader :path
  attr_reader :portal_profile

  def initialize(path, portal_profile)
    @path = path
    @portal_profile = portal_profile
  end
end
