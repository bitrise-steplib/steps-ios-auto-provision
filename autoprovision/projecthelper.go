package autoprovision

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-io/xcode-project/xcodeproj"
	"github.com/bitrise-io/xcode-project/xcscheme"
	"github.com/bitrise-io/xcode-project/xcworkspace"
)

// ProjectHelper ...
type ProjectHelper struct {
	MainTarget xcodeproj.Target
	Targets    []xcodeproj.Target
	Platform   Platform
}

// Platform of the target
// iOS, tvOS, macOS
type Platform string

// Const
const (
	IOS   Platform = "iOS"
	TVOS  Platform = "tvOS"
	MacOS Platform = "macOS"
)

// New checks the provided project or workspace and generate a ProjectHelper with the provided scheme and configuration
// Previously in the ruby version the initialize method did the same
// It returns a new ProjectHelper pointer and a configuration to use.
func New(projOrWSPath, schemeName, configurationName string) (*ProjectHelper, string, error) {
	// Maybe we should do this checks during the input parsing
	if exits, err := pathutil.IsPathExists(projOrWSPath); err != nil {
		return nil, "", err
	} else if !exits {
		return nil, "", fmt.Errorf("provided path does not exists: %s", projOrWSPath)
	}

	// Get the project of the provided .xcodeproj or .xcworkspace
	xcproj, _, err := findBuiltProject(projOrWSPath, schemeName, configurationName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find build project, error: %s", err)
	}

	mainTarget, err := mainTargetOfScheme(xcproj, schemeName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find the main target of the scheme (%s), error: %s", schemeName, err)
	}

	scheme, ok := xcproj.Scheme(schemeName)
	if !ok {
		return nil, "", fmt.Errorf("no scheme found with name: %s in project: %s", schemeName, projOrWSPath)
	}

	// Check if the archive is availabe for the scheme or not
	if _, archivable := scheme.AppBuildActionEntry(); archivable != true {
		return nil, "", fmt.Errorf("archive action not defined for scheme: %s", scheme.Name)
	}

	// Get the platform (PLATFORM_DISPLAY_NAME) -iphoneos, macosx, appletvos
	platf, err := platform(xcproj, mainTarget, configurationName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find platform of the project, error: %s", err)
	}

	// Configuration
	conf, err := configuration(configurationName, scheme, xcproj)
	if err != nil {
		return nil, "", err
	}
	return &ProjectHelper{
			MainTarget: mainTarget,
			Targets:    xcproj.Proj.Targets,
			Platform:   platf,
		}, conf,
		nil
}

// UsesXcodeAutoCodeSigning checks the project uses automatically managed code signing
// It checks the main target "ProvisioningStyle" attribute first then the "CODE_SIGN_STYLE" for the provided configuration
// It returns true if the project uses automatically code signing
func UsesXcodeAutoCodeSigning(xcProj xcodeproj.XcodeProj, mainTarget xcodeproj.Target, config string) (bool, error) {
	settings, err := xcProj.TargetBuildSettings(mainTarget.Name, config)
	if err != nil {
		return false, fmt.Errorf("failed to fetch project settings (%s), error: %s", xcProj.Path, err)
	}

	if provStle, err := settings.String("ProvisioningStyle"); err != nil && !serialized.IsKeyNotFoundError(err) {
		return false, err
	} else if provStle == "Automatic" {
		return true, nil
	}

	for _, buildConf := range mainTarget.BuildConfigurationList.BuildConfigurations {
		if buildConf.Name != config {
			continue
		}

		if signStyle, err := buildConf.BuildSettings.String("CODE_SIGN_STYLE"); err != nil && !serialized.IsKeyNotFoundError(err) {
			return false, err
		} else if signStyle == "Automatic" {
			return true, nil
		}
	}
	return false, nil
}

// Get the platform (PLATFORM_DISPLAY_NAME) - iOS, tvOS, macOS
func platform(xcproj xcodeproj.XcodeProj, mainTarget xcodeproj.Target, configurationName string) (Platform, error) {
	settings, err := xcproj.TargetBuildSettings(mainTarget.Name, configurationName)
	if err != nil {
		return "", fmt.Errorf("failed to fetch project settings (%s), error: %s", xcproj.Path, err)
	}

	platformDisplayName, err := settings.String("PLATFORM_DISPLAY_NAME")
	if err != nil {
		return "", fmt.Errorf("no PLATFORM_DISPLAY_NAME config found for (%s) target", mainTarget.Name)
	}

	if platformDisplayName != string(IOS) && platformDisplayName != string(MacOS) && platformDisplayName != string(TVOS) {
		return "", fmt.Errorf("not supported platform. Platform (PLATFORM_DISPLAY_NAME) = %s", platformDisplayName)
	}
	return Platform(platformDisplayName), nil
}

func configuration(configurationName string, scheme xcscheme.Scheme, xcproj xcodeproj.XcodeProj) (string, error) {
	defaultConfiguration := scheme.ArchiveAction.BuildConfiguration
	var configuration string
	if configurationName == "" || configurationName == defaultConfiguration {
		configuration = defaultConfiguration
	} else if configurationName != defaultConfiguration {
		for _, target := range xcproj.Proj.Targets {
			var configNames []string
			for _, conf := range target.BuildConfigurationList.BuildConfigurations {
				configNames = append(configNames, conf.Name)
			}
			if !sliceutil.IsStringInSlice(configurationName, configNames) {
				return "", fmt.Errorf("build configuration (%s) not defined for target: (%s)", configurationName, target.Name)
			}
		}
		log.Warnf("Using defined build configuration: %s instead of the scheme's default one: %s", configurationName, defaultConfiguration)
		configuration = configurationName
	}

	return configuration, nil
}

// mainTargetOfScheme return the main target
func mainTargetOfScheme(proj xcodeproj.XcodeProj, scheme string) (xcodeproj.Target, error) {
	projTargets := proj.Proj.Targets
	sch, ok := proj.Scheme(scheme)
	if !ok {
		return xcodeproj.Target{}, fmt.Errorf("Failed to found scheme (%s) in project", scheme)
	}

	var blueIdent string
	for _, entry := range sch.BuildAction.BuildActionEntries {
		if entry.BuildableReference.IsAppReference() {
			blueIdent = entry.BuildableReference.BlueprintIdentifier
			break
		}
	}

	// Search for the main target
	for _, t := range projTargets {
		if t.ID == blueIdent {
			return t, nil

		}
	}
	return xcodeproj.Target{}, fmt.Errorf("failed to find the project's main target for scheme (%s)", scheme)
}

// findBuiltProject returns the Xcode project which will be built for the provided scheme
func findBuiltProject(pth, schemeName, configurationName string) (xcodeproj.XcodeProj, string, error) {
	var scheme xcscheme.Scheme
	var schemeContainerDir string

	if xcodeproj.IsXcodeProj(pth) {
		project, err := xcodeproj.Open(pth)
		if err != nil {
			return xcodeproj.XcodeProj{}, "", err
		}

		var ok bool
		scheme, ok = project.Scheme(schemeName)
		if !ok {
			return xcodeproj.XcodeProj{}, "", fmt.Errorf("no scheme found with name: %s in project: %s", schemeName, pth)
		}
		schemeContainerDir = filepath.Dir(pth)
	} else if xcworkspace.IsWorkspace(pth) {
		workspace, err := xcworkspace.Open(pth)
		if err != nil {
			return xcodeproj.XcodeProj{}, "", err
		}

		var containerProject string
		scheme, containerProject, err = workspace.Scheme(schemeName)
		if err != nil {
			return xcodeproj.XcodeProj{}, "", fmt.Errorf("no scheme found with name: %s in workspace: %s, error: %s", schemeName, pth, err)
		}
		schemeContainerDir = filepath.Dir(containerProject)
	} else {
		return xcodeproj.XcodeProj{}, "", fmt.Errorf("unknown project extension: %s", filepath.Ext(pth))
	}

	if configurationName == "" && scheme.ArchiveAction.BuildConfiguration == "" {
		return xcodeproj.XcodeProj{}, "", fmt.Errorf("no configuration provided nor default defined for the scheme's (%s) archive action", schemeName)
	} else if configurationName == "" {
		configurationName = scheme.ArchiveAction.BuildConfiguration
	}

	var archiveEntry xcscheme.BuildActionEntry
	for _, entry := range scheme.BuildAction.BuildActionEntries {
		if entry.BuildForArchiving != "YES" || !entry.BuildableReference.IsAppReference() {
			continue
		}
		archiveEntry = entry
		break
	}

	if archiveEntry.BuildableReference.BlueprintIdentifier == "" {
		return xcodeproj.XcodeProj{}, "", fmt.Errorf("archivable entry not found")
	}

	projectPth, err := archiveEntry.BuildableReference.ReferencedContainerAbsPath(schemeContainerDir)
	if err != nil {
		return xcodeproj.XcodeProj{}, "", err
	}

	project, err := xcodeproj.Open(projectPth)
	if err != nil {
		return xcodeproj.XcodeProj{}, "", err
	}

	return project, scheme.Name, nil
}
