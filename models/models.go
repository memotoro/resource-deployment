package models

import (
	"fmt"
	"strings"

	v1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1alpha2"
)

// Status represents the data back after deletion
type Status struct {
	APIVersion string `json:"apiVersion"`
	Code       int    `json:"code"`
	Kind       string `json:"kind"`
	Message    string `json:"message"`
	Metadata   struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

// GetVersion returns the version in API format
func GetVersion(sd v1alpha2.SeldonDeployment) string {
	return sd.APIVersion[strings.LastIndex(sd.APIVersion, "/")+1:]
}

// GetAPIGroup returns the APIGroup in API format
func GetAPIGroup(sd v1alpha2.SeldonDeployment) string {
	return sd.APIVersion[:strings.LastIndex(sd.APIVersion, "/")]
}

// GetKindValue returns the name of the resource in API format. It could be extented for other resources
func GetKindValue(sd v1alpha2.SeldonDeployment) string {
	switch sd.Kind {
	case "SeldonDeployment":
		return fmt.Sprintf("%ss", strings.ToLower(sd.Kind))
	default:
		return ""
	}
}
