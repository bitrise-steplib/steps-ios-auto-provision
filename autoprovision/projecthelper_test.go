package autoprovision

import (
	"testing"

	"github.com/bitrise-io/xcode-project/xcodeproj"
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

func TestUsesXcodeAutoCodeSigning(t *testing.T) {
	//
	// Init test cases
	schemeCases := []string{
		"Xcode-10_default",
		"Xcode-10_default",
		"Xcode-10_mac",
		"Xcode-10_mac",
		"TV_OS",
		"TV_OS",
	}
	configCases := []string{
		"Debug",
		"Release",
		"Debug",
		"Release",
		"Debug",
		"Release",
	}
	projectCases := []string{
		"/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_default.xcworkspace",
		"/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_default.xcworkspace",
		"/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_mac.xcodeproj",
		"/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/Xcode-10_mac.xcodeproj",
		"/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/TV_OS.xcodeproj",
		"/Users/akosbirmacher/Develop/go/src/github.com/bitrise-steplib/steps-ios-auto-provision/_tmp/TV_OS.xcodeproj",
	}
	var xcProjCases []xcodeproj.XcodeProj
	var projHelpCases []ProjectHelper

	for i, schemeCase := range schemeCases {
		xcProj, _, err := findBuiltProject(
			projectCases[i],
			schemeCase,
			configCases[i],
		)
		if err != nil {
			t.Fatalf("Failed to generate XcodeProj for test case, error: %s", err)
		}
		xcProjCases = append(xcProjCases, xcProj)

		projHelp, _, err := New(
			projectCases[i],
			schemeCase,
			configCases[i],
		)
		if err != nil {
			t.Fatalf("Failed to generate projectHelper for test case, error: %s", err)
		}
		projHelpCases = append(projHelpCases, *projHelp)
	}

	//
	// Test
	tests := []struct {
		name       string
		xcProj     xcodeproj.XcodeProj
		mainTarget xcodeproj.Target
		config     string
		want       bool
		wantErr    bool
	}{
		{
			name:       schemeCases[0] + " Debug",
			xcProj:     xcProjCases[0],
			mainTarget: projHelpCases[0].MainTarget,
			config:     configCases[0],
			want:       true,
			wantErr:    false,
		},
		{
			name:       schemeCases[1] + " Release",
			xcProj:     xcProjCases[1],
			mainTarget: projHelpCases[1].MainTarget,
			config:     configCases[1],
			want:       true,
			wantErr:    false,
		},
		{
			name:       schemeCases[2] + " Debug",
			xcProj:     xcProjCases[2],
			mainTarget: projHelpCases[2].MainTarget,
			config:     configCases[2],
			want:       false,
			wantErr:    false,
		},
		{
			name:       schemeCases[3] + " Release",
			xcProj:     xcProjCases[3],
			mainTarget: projHelpCases[3].MainTarget,
			config:     configCases[3],
			want:       false,
			wantErr:    false,
		},
		{
			name:       schemeCases[4] + " Debug",
			xcProj:     xcProjCases[4],
			mainTarget: projHelpCases[4].MainTarget,
			config:     configCases[4],
			want:       false,
			wantErr:    false,
		},
		{
			name:       schemeCases[5] + " Release",
			xcProj:     xcProjCases[5],
			mainTarget: projHelpCases[5].MainTarget,
			config:     configCases[5],
			want:       false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UsesXcodeAutoCodeSigning(tt.xcProj, tt.mainTarget, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("UsesXcodeAutoCodeSigning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UsesXcodeAutoCodeSigning() = %v, want %v", got, tt.want)
			}
		})
	}
}
