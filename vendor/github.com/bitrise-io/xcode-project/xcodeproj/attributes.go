package xcodeproj

import "github.com/bitrise-io/xcode-project/serialized"

// ProjectAtributes ...
type ProjectAtributes struct {
	TargetAttributes serialized.Object
}

func parseProjectAttributes(rawPBXProj serialized.Object) (ProjectAtributes, error) {
	var attributes ProjectAtributes
	attributesObject, err := rawPBXProj.Object("attributes")
	if err != nil {
		return ProjectAtributes{}, err
	}

	attributes.TargetAttributes, err = parseTargetAttributes(attributesObject)
	if err != nil {
		return ProjectAtributes{}, err
	}

	return attributes, nil
}

func parseTargetAttributes(attributesObject serialized.Object) (serialized.Object, error) {
	return attributesObject.Object("TargetAttributes")
}
