package readers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/memotoro/seldonio-resource-deployment/clients"
)

// ReadContentFile reads content of the file or http endpoint
func ReadContentFile(client clients.Client, resourceFile string) ([]byte, error) {
	if strings.Contains(resourceFile, "http") {
		req, err := http.NewRequest(http.MethodGet, resourceFile, nil)
		if err != nil {
			return nil, err
		}

		resp, data, err := client.ExecuteCall(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("%v - %v", resp.StatusCode, string(data))
		}

		return data, nil
	}

	data, err := ioutil.ReadFile(resourceFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}
