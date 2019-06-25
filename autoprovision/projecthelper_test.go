package autoprovision

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name              string
		projOrWSPath      string
		schemeName        string
		configurationName string
		wantConfiguration string
		wantErr           bool
	}{
		{
			name:              "Xcode 10 workspace - iOS",
			projOrWSPath:      "/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_default.xcworkspace",
			schemeName:        "Xcode-10_default",
			configurationName: "Debug",
			wantConfiguration: "Debug",
			wantErr:           false,
		},
		{
			name:              "Xcode 10 workspace - iOS - Default configuration",
			projOrWSPath:      "/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_default.xcworkspace",
			schemeName:        "Xcode-10_default",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "Xcode 10 workspace - iOS - Default configuration - Gdańsk scheme",
			projOrWSPath:      "/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_default.xcworkspace",
			schemeName:        "Gdańsk",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "Xcode-10_mac project - MacOS - Debug configuration",
			projOrWSPath:      "/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_mac.xcodeproj",
			schemeName:        "Xcode-10_mac",
			configurationName: "Debug",
			wantConfiguration: "Debug",
			wantErr:           false,
		},
		{
			name:              "Xcode-10_mac project - MacOS - Default configuration",
			projOrWSPath:      "/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_mac.xcodeproj",
			schemeName:        "Xcode-10_mac",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
		{
			name:              "TV_OS.xcodeproj project - MacOS - Default configuration",
			projOrWSPath:      "/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/TV_OS.xcodeproj",
			schemeName:        "TV_OS",
			configurationName: "",
			wantConfiguration: "Release",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projHelp, conf, err := New(tt.projOrWSPath, tt.schemeName, tt.configurationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if projHelp == nil {
				t.Errorf("New() error = No projectHelper was generated")
			}
			if conf != tt.wantConfiguration {
				t.Errorf("New() got1 = %v, want %v", conf, tt.wantConfiguration)
			}
		})
	}
}
