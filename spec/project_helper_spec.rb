require 'xcodeproj'
require 'xcodeproj/plist'
require 'tmpdir'
require 'fileutils'

require_relative '../lib/autoprovision/project_helper.rb'

def recreate_shared_schemes(project)
  schemes_dir = Xcodeproj::XCScheme.user_data_dir(project.path)
  FileUtils.rm_rf(schemes_dir)
  FileUtils.mkdir_p(schemes_dir)

  xcschememanagement = {}
  xcschememanagement['SchemeUserState'] = {}
  xcschememanagement['SuppressBuildableAutocreation'] = {}

  project.targets.each do |target|
    scheme = Xcodeproj::XCScheme.new
    scheme.add_build_target(target)
    scheme.add_test_target(target) if target.respond_to?(:test_target_type?) && target.test_target_type?
    yield scheme, target if block_given?
    scheme.save_as(project.path, target.name, true)
    xcschememanagement['SchemeUserState']["#{target.name}.xcscheme"] = {}
    xcschememanagement['SchemeUserState']["#{target.name}.xcscheme"]['isShown'] = true
  end

  xcschememanagement_path = schemes_dir + 'xcschememanagement.plist'
  Xcodeproj::Plist.write_to_path(xcschememanagement, xcschememanagement_path)
end

def test_project_dir
  src = './spec/fixtures/project'
  dst = Dir.mktmpdir('foo')
  FileUtils.copy_entry(src, dst)
  dst
end

RSpec.describe 'ProjectHelper' do
  let(:project_with_target_attributes) do
    path = File.join(test_project_dir, 'foo.xcodeproj')
    project = Xcodeproj::Project.open(path)
    recreate_shared_schemes(project)
    project
  end

  let(:project_without_target_attributes) do
    project = project_with_target_attributes
    project.root_object.attributes['TargetAttributes'] = nil
    project.save
    project
  end

  describe '#uses_xcode_auto_codesigning?' do
    subject { ProjectHelper.new(project.path, 'foo', '').uses_xcode_auto_codesigning? }

    context 'when new Xcode project uses auto signing with TargetAttributes' do
      let(:project) { project_with_target_attributes }
      it { is_expected.to eq true }
    end

    context 'when new Xcode project uses auto signing without TargetAttributes' do
      let(:project) { project_without_target_attributes }
      it { is_expected.to eq true }
    end
  end
end
