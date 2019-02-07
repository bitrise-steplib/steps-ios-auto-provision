package analyser

import (
	"fmt"
	"path/filepath"

	"github.com/bitrise-tools/xcode-project/xcodeproj"
	"github.com/bitrise-tools/xcode-project/xcscheme"
	"github.com/bitrise-tools/xcode-project/xcworkspace"
)

type Project struct {
	pth               string
	schemeName        string
	configurationName string
}

func New(pth, scheme, configuration string) Info {
	return Info{
		pth:               pth,
		schemeName:        scheme,
		configurationName: configuration,
	}
}

func (i Info) CodesignIdentity() (string, error) {
	targets, err := i.targetsToSign()
	if err != nil {
		return "", err
	}

	configuration, err := i.configuration()
	if err != nil {
		return "", err
	}

	for _, t := range targets {
		for _, c := range t.BuildConfigurationList.BuildConfigurations {
			if c.Name == configuration {
				c.BuildSettings.String("CODE_SIGN_IDENTITY")
			}
		}
	}

	return "", err
}

func (i Info) configuration() (string, error) {
	scheme, _, err := i.scheme()
	if err != nil {
		return "", err
	}

	configuration := i.configurationName
	if configuration == "" {
		configuration = scheme.ArchiveAction.BuildConfiguration
	}

	if configuration == "" {
		return "", fmt.Errorf("no configuration provided nor default defined for the scheme's (%s) archive action", i.schemeName)
	}

	return configuration, nil
}

func (i Info) scheme() (*xcscheme.Scheme, string, error) {
	var scheme xcscheme.Scheme
	var schemeContainerDir string

	if xcodeproj.IsXcodeProj(i.pth) {
		project, err := xcodeproj.Open(i.pth)
		if err != nil {
			return nil, "", err
		}

		var ok bool
		scheme, ok = project.Scheme(i.schemeName)
		if !ok {
			return nil, "", fmt.Errorf("no scheme found with name: %s in project: %s", i.schemeName, i.pth)
		}
		schemeContainerDir = filepath.Dir(i.pth)
	} else if xcworkspace.IsWorkspace(i.pth) {
		workspace, err := xcworkspace.Open(i.pth)
		if err != nil {
			return nil, "", err
		}

		var containerProject string
		scheme, containerProject, err = workspace.Scheme(i.schemeName)
		if err != nil {
			if xcworkspace.IsSchemeNotFoundError(err) {
				return nil, "", err
			}
			return nil, "", fmt.Errorf("failed to find scheme with name: %s in workspace: %s, error: %s", i.schemeName, i.pth, err)
		}
		schemeContainerDir = filepath.Dir(containerProject)
	} else {
		return nil, "", fmt.Errorf("unknown project extension: %s", filepath.Ext(i.pth))
	}

	return &scheme, schemeContainerDir, nil
}

func (i Info) targetsToSign() ([]xcodeproj.Target, error) {
	scheme, schemeContainerDir, err := i.scheme()
	if err != nil {
		return nil, err
	}

	configuration, err := i.configuration()
	if err != nil {
		return nil, err
	}

	archiveEntry, ok := scheme.AppBuildActionEntry()
	if !ok {
		return nil, fmt.Errorf("archivable entry not found")
	}

	projectPth, err := archiveEntry.BuildableReference.ReferencedContainerAbsPath(schemeContainerDir)
	if err != nil {
		return nil, err
	}

	project, err := xcodeproj.Open(projectPth)
	if err != nil {
		return nil, err
	}

	mainTarget, ok := project.Proj.Target(archiveEntry.BuildableReference.BlueprintIdentifier)
	if !ok {
		return nil, fmt.Errorf("target not found: %s", archiveEntry.BuildableReference.BlueprintIdentifier)
	}

	targets := []xcodeproj.Target{mainTarget}

	for _, t := range mainTarget.DependentTargets() {
		if t.IsExecutableProduct() {
			targets = append(targets, t)
		}

		for _, c := range t.BuildConfigurationList.BuildConfigurations {
			if c.Name == configuration {
				hasCodesignSetting := false
				for _, k := range []string{
					"CODE_SIGN_STYLE",
					"DEVELOPMENT_TEAM",
					"PROVISIONING_PROFILE_SPECIFIER",
					"PROVISIONING_PROFILE",
				} {
					v, err := c.BuildSettings.String(k)
					hasCodesignSetting = (err == nil && len(v) > 0)
					if hasCodesignSetting {
						break
					}
				}

				if hasCodesignSetting {
					targets = append(targets, t)
				}
			}
		}
	}

	return targets, nil
}
