package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/memotoro/seldonio-resource-deployment/clients"
	"github.com/memotoro/seldonio-resource-deployment/models"

	v1alpha2 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1alpha2"
)

// CreateResource creates the resource in the cluster via API
func CreateResource(client clients.Client, resourceData []byte, namespace string) (*v1alpha2.SeldonDeployment, error) {
	var resource v1alpha2.SeldonDeployment
	loadResourceFromData(resourceData, namespace, &resource)

	url := getCreateResourceURL(client, resource)

	payload, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resp, data, err := client.ExecuteCall(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 202 {
		return nil, fmt.Errorf("%v - %v", resp.StatusCode, string(data))
	}

	var result v1alpha2.SeldonDeployment
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetResourceStatus gets the resource status in the cluster via API
func GetResourceStatus(client clients.Client, resourceData []byte, namespace string) (*v1alpha2.SeldonDeployment, error) {
	var resource v1alpha2.SeldonDeployment
	loadResourceFromData(resourceData, namespace, &resource)

	url := getResourceStatusURL(client, resource)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, data, err := client.ExecuteCall(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode == 200 {
		var result v1alpha2.SeldonDeployment
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		return &result, nil
	}

	return nil, fmt.Errorf("%v - %v", resp.StatusCode, string(data))
}

// DeleteResource changes the cluster by deleting the resource via API
func DeleteResource(client clients.Client, resourceData []byte, namespace string) (*models.Status, error) {
	var resource v1alpha2.SeldonDeployment
	loadResourceFromData(resourceData, namespace, &resource)

	url := getDeleteResourceURL(client, resource)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, data, err := client.ExecuteCall(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 202 {
		return nil, fmt.Errorf("%v - %v", resp.StatusCode, string(data))
	}

	var result models.Status
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func getCreateResourceURL(client clients.Client, resource v1alpha2.SeldonDeployment) string {
	return fmt.Sprintf("%s/apis/%s/%s/namespaces/%s/%s", client.BaseEndpoint(), models.GetAPIGroup(resource), models.GetVersion(resource), resource.ObjectMeta.Namespace, models.GetKindValue(resource))
}

func getResourceStatusURL(client clients.Client, resource v1alpha2.SeldonDeployment) string {
	return fmt.Sprintf("%s/apis/%s/%s/namespaces/%s/%s/%s/status", client.BaseEndpoint(), models.GetAPIGroup(resource), models.GetVersion(resource), resource.ObjectMeta.Namespace, models.GetKindValue(resource), resource.ObjectMeta.Name)
}

func getDeleteResourceURL(client clients.Client, resource v1alpha2.SeldonDeployment) string {
	return fmt.Sprintf("%s/apis/%s/%s/namespaces/%s/%s/%s", client.BaseEndpoint(), models.GetAPIGroup(resource), models.GetVersion(resource), resource.ObjectMeta.Namespace, models.GetKindValue(resource), resource.ObjectMeta.Name)
}

func loadResourceFromData(resourceData []byte, namespace string, resource *v1alpha2.SeldonDeployment) {
	if err := json.Unmarshal(resourceData, &resource); err != nil {
		log.Fatalf("Error unmarshalling data. Details : %v", err)
	}
	if resource.ObjectMeta.Namespace == "" {
		resource.ObjectMeta.Namespace = namespace
	}
}
