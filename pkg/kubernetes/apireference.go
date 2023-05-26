package kubernetes

import (
	"strings"

	openapi_v2 "github.com/google/gnostic/openapiv2"
)

func (k *K8sApiReference) GetApiDocV2(field string) string {
	startPoint := ""
	// the path must be formated like "path1.path2.path3"
	paths := strings.Split(field, ".")
	group := strings.Split(k.ApiVersion.Group, ".")

	openapiSchema, err := k.Discovery.OpenAPISchema()
	if err != nil {
		return ""
	}

	definitions := openapiSchema.GetDefinitions().GetAdditionalProperties()

	// extract the startpoint by searching the highest leaf corresponding to the requested group qnd kind
	for _, prop := range definitions {
		if strings.HasSuffix(prop.GetName(), group[0]+"."+k.ApiVersion.Version+"."+k.Kind) {
			startPoint = prop.GetName()

			break
		}
	}

	// recursively parse the definitions to find the description of the latest part of the given path
	description := k.recursePath(definitions, startPoint, paths)

	return description
}

func (k *K8sApiReference) recursePath(definitions []*openapi_v2.NamedSchema, leaf string, paths []string) string {
	description := ""

	for _, prop := range definitions {
		// search the requested leaf
		if prop.GetName() == leaf {
			for _, addProp := range prop.GetValue().GetProperties().GetAdditionalProperties() {
				// search the additional property of the leaf corresponding the current path
				if addProp.GetName() == paths[0] {
					// the last path or the path is string, we get the description and we go out
					if len(paths) == 1 || addProp.GetValue().GetType().String() == "value:\"string\"" {
						// extract the path description as we are at the end of the paths
						description = addProp.GetValue().Description
					} else {
						// the path is an object, we extract the xref
						if addProp.GetValue().GetXRef() != "" {
							splitRef := strings.Split(addProp.GetValue().GetXRef(), "/")
							reducedPaths := paths[1:]
							description = k.recursePath(definitions, splitRef[len(splitRef)-1], reducedPaths)
						}

						// the path is an array, we take the first xref from the items
						if len(addProp.GetValue().GetItems().GetSchema()) == 1 {
							splitRef := strings.Split(addProp.GetValue().GetItems().GetSchema()[0].GetXRef(), "/")
							reducedPaths := paths[1:]
							description = k.recursePath(definitions, splitRef[len(splitRef)-1], reducedPaths)
						}
					}

					break
				}
			}

			break
		}
	}

	return description
}
