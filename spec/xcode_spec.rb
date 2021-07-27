require_relative '../xcode'

RSpec.describe 'Xcode' do
  describe '#parse_xcodebuild_version_out' do
    it 'successfully parses known xcodebuild -version output' do
      version, build_version, major_version = Xcode.send(:parse_xcodebuild_version_out, 'Xcode 13.0
Build version 13A5192j')

      expect(version).to eq('13.0')
      expect(build_version).to eq('13A5192j')
      expect(major_version).to eq(13)
    end

    it 'fails for unknown xcodebuild -version output' do
      expect { Xcode.parse_xcodebuild_version_out('13') }.to raise_error(RuntimeError)
    end
  end
end
