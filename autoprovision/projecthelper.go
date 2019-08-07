package autoprovision

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/xcode-project/serialized"
	"github.com/bitrise-io/xcode-project/xcodeproj"
	"github.com/bitrise-io/xcode-project/xcscheme"
	"github.com/bitrise-io/xcode-project/xcworkspace"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
	"howett.net/plist"
)

// ProjectHelper ...
type ProjectHelper struct {
	MainTarget    xcodeproj.Target
	Targets       []xcodeproj.Target
	Platform      Platform
	XcProj        xcodeproj.XcodeProj
	Configuration string
}

// Platform of the target
// iOS, tvOS, macOS
type Platform string

// BundleIDPlatform ...
func (p Platform) BundleIDPlatform() (*appstoreconnect.BundleIDPlatform, error) {
	var apiPlatform appstoreconnect.BundleIDPlatform
	switch p {
	case IOS, TVOS:
		apiPlatform = appstoreconnect.IOS
	case MacOS:
		apiPlatform = appstoreconnect.MacOS
	default:
		return nil, errors.New("unknown platform: " + string(p))
	}
	return &apiPlatform, nil
}

// Const
const (
	IOS   Platform = "iOS"
	TVOS  Platform = "tvOS"
	MacOS Platform = "macOS"
)

// NewProjectHelper checks the provided project or workspace and generate a ProjectHelper with the provided scheme and configuration
// Previously in the ruby version the initialize method did the same
// It returns a new ProjectHelper pointer and a configuration to use.
func NewProjectHelper(projOrWSPath, schemeName, configurationName string) (*ProjectHelper, string, error) {
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
		MainTarget:    mainTarget,
		Targets:       xcproj.Proj.Targets,
		Platform:      platf,
		XcProj:        xcproj,
		Configuration: conf,
	}, conf,
		nil
}

func (p ProjectHelper) ArchivableTargetBundleIDToEntitlements() (map[string]serialized.Object, error) {
	targets := append([]xcodeproj.Target{p.MainTarget}, p.MainTarget.DependentExecutableProductTargets(false)...)

	entitlementsByBundleID := map[string]serialized.Object{}

	for _, target := range targets {
		bundleID, err := p.TargetBundleID(target.Name, p.Configuration)
		if err != nil {
			return nil, fmt.Errorf("Failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlements, err := p.targetEntitlements(target.Name, p.Configuration)
		if err != nil && !serialized.IsKeyNotFoundError(err) {
			return nil, fmt.Errorf("Failed to get target (%s) bundle id: %s", target.Name, err)
		}

		entitlementsByBundleID[bundleID] = entitlements
	}

	return entitlementsByBundleID, nil
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

// ProjectTeamID returns the development team's ID
// If there is mutlitple development team in the project (different team for targets) it will return an error
// It returns the development team's ID
func (p ProjectHelper) ProjectTeamID(config string) (string, error) {
	var teamID string

	for _, target := range p.Targets {
		currentTeamID, err := targetTeamID(p.XcProj, target.Name, config)
		if err != nil {
			// Do nothing
		}
		log.Debugf("%s target build settings team id: %s", target.Name, currentTeamID)

		if currentTeamID == "" {
			log.Warnf("no DEVELOPMENT_TEAM build settings found for target: %s, checking target attributes...", target.Name)

			targetAttributes, err := p.XcProj.Proj.Attributes.TargetAttributes.Object(target.ID)
			if err != nil {
				return "", fmt.Errorf("failed to parse target %s target attributes, error: %s", target.ID, err)
			}

			targetAttributesTeamID, err := targetAttributes.String("DevelopmentTeam")
			if err != nil && !serialized.IsKeyNotFoundError(err) {
				return "", fmt.Errorf("failed to parse development team for target %s, error: %s", target.ID, err)
			}
			if targetAttributesTeamID == "" {
				log.Warnf("no DevelopmentTeam target attribute found for target: %s", target.Name)
				continue
			}

			currentTeamID = targetAttributesTeamID
		}

		if teamID == "" {
			teamID = currentTeamID
			continue
		}

		if teamID != currentTeamID {
			log.Warnf("target team id: %s does not match to the already registered team id: %s", currentTeamID, teamID)
			teamID = ""
			break
		}
	}

	return teamID, nil

}

// ProjectCodeSignIdentity returns the codesign identity of the project
// If there is mutlitple codesign identity in the project (different identity for targets) it will return an error
// It returns the codesign identity
func (p ProjectHelper) ProjectCodeSignIdentity(config string) (string, error) {
	var codesignIdentity string

	for _, t := range p.Targets {
		targetIdentity, err := targetCodesignIdentity(p.XcProj, t.Name, config)
		if err != nil {
			return "", err
		}

		log.Debugf("%s codesign identity: %s", t.Name, targetIdentity)

		if targetIdentity == "" {
			log.Warnf("no CODE_SIGN_IDENTITY build settings found for target: %s", t.Name)
			continue
		}

		if codesignIdentity == "" {
			codesignIdentity = targetIdentity
			continue
		}

		if !codesignIdentitesMatch(codesignIdentity, targetIdentity) {
			log.Warnf("target codesign identity: %s does not match to the already registered codesign identity: %s", targetIdentity, codesignIdentity)
			codesignIdentity = ""
			break
		}
	}
	return codesignIdentity, nil
}

// 'iPhone Developer' should match to 'iPhone Developer: Bitrise Bot (ABCD)'
func codesignIdentitesMatch(identity1, identity2 string) bool {
	if strings.Contains(strings.ToLower(identity1), strings.ToLower(identity2)) {
		return true
	}
	if strings.Contains(strings.ToLower(identity2), strings.ToLower(identity1)) {
		return true
	}
	return false
}

func targetCodesignIdentity(xcProj xcodeproj.XcodeProj, targatName, config string) (string, error) {
	settings, err := xcProj.TargetBuildSettings(targatName, config)
	if err != nil {
		return "", fmt.Errorf("failed to fetch target (%s) settings, error: %s", targatName, err)
	}
	return settings.String("CODE_SIGN_IDENTITY")
}

func targetTeamID(xcProj xcodeproj.XcodeProj, targatName, config string) (string, error) {
	settings, err := xcProj.TargetBuildSettings(targatName, config)
	if err != nil {
		return "", fmt.Errorf("failed to fetch target (%s) settings, error: %s", targatName, err)
	}

	devTeam, err := settings.String("DEVELOPMENT_TEAM")
	if serialized.IsKeyNotFoundError(err) {
		return devTeam, nil
	}
	return devTeam, err

}

// TargetBundleID returns the target bundle ID
// First it tries to fetch the bundle ID from the `PRODUCT_BUNDLE_IDENTIFIER` build settings
// If it's no available it will fetch the target's Info.plist and search for the `CFBundleIdentifier` key.
// The CFBundleIdentifier's value is not resolved in the Info.plist, so it will try to resolve it by the resolveBundleID()
// It returns  the target bundle ID
func (p ProjectHelper) TargetBundleID(name, conf string) (string, error) {
	settings, err := p.XcProj.TargetBuildSettings(name, conf)
	if err != nil {
		return "", fmt.Errorf("failed to fetch target (%s) settings, error: %s", name, err)
	}

	bundleID, err := settings.String("PRODUCT_BUNDLE_IDENTIFIER")
	if bundleID != "" {
		return bundleID, nil
	}

	log.Debugf("PRODUCT_BUNDLE_IDENTIFIER env not found in 'xcodebuild -showBuildSettings -project %s -target %s -configuration %s command's output", p.XcProj.Path, name, conf)
	log.Debugf("checking the Info.plist file's CFBundleIdentifier property...")

	infoPlistPath, err := settings.String("INFOPLIST_FILE")
	if err != nil {
		return "", fmt.Errorf("failed to find info.plst file, error: %s", err)
	}

	if infoPlistPath == "" {
		return "", fmt.Errorf("failed to to determine bundle id: xcodebuild -showBuildSettings does not contains PRODUCT_BUNDLE_IDENTIFIER nor INFOPLIST_FILE' unless info_plist_path")
	}

	b, err := fileutil.ReadBytesFromFile(infoPlistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Info.plist, error: %s", err)
	}

	var options map[string]interface{}
	if _, err := plist.Unmarshal(b, &options); err != nil {
		return "", fmt.Errorf("failed to unmarshal Info.plist, error: %s ", err)
	}

	bundleID, ok := options["CFBundleIdentifier"].(string)
	if !ok || bundleID == "" {
		return "", fmt.Errorf("failed to parse CFBundleIdentifier from the Info.plist")
	}

	if !strings.Contains(bundleID, "$") {
		return bundleID, nil
	}

	log.Warnf("CFBundleIdentifier defined with variable: %s, trying to resolve it...", bundleID)
	resolved, err := resolveBundleID(bundleID, settings)
	if err != nil {
		return "", fmt.Errorf("failed to resolve bundle ID, error: %s", err)
	}
	log.Warnf("resolved CFBundleIdentifier: %s", resolved)

	return resolved, nil
}

func resolveBundleID(bundleID string, buildSettings serialized.Object) (string, error) {
	r, err := regexp.Compile(".+[.][$][(].+[:].+[)]*")
	if err != nil {
		return "", err
	}

	if !r.MatchString(bundleID) {
		return "", fmt.Errorf("failed to match regex .+[.][$][(].+[:].+[)]* to %s bundleID", bundleID)
	}

	captures := r.FindString(bundleID)

	prefix := strings.Split(captures, "$")[0]
	envKey := strings.Split(strings.SplitAfter(captures, "(")[1], ":")[0]
	suffix := strings.Join(strings.SplitAfter(captures, ")")[1:], "")

	envValue, err := buildSettings.String(envKey)
	if err != nil {
		return "", fmt.Errorf("failed to find enviroment variable value for key %s, error: %s", envKey, err)
	}
	return prefix + envValue + suffix, nil

}

func (p ProjectHelper) targetEntitlements(name, config string) (serialized.Object, error) {
	o, err := p.XcProj.TargetCodeSignEntitlements(name, config)
	if err != nil && !serialized.IsKeyNotFoundError(err) {
		return nil, err
	}
	return o, nil
}

// TODO
//   def force_code_sign_properties
