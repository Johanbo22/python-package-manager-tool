package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const PythonServiceURL = "http://127.0.0.1:5000"

type LocalPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PythonBridgeClient struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewPythonBridgeClient() *PythonBridgeClient {
	return &PythonBridgeClient{
		APIKey:     os.Getenv("MANAGER_API_KEY"),
		HTTPClient: &http.Client{},
	}
}

func (client *PythonBridgeClient) GetInstalledPackages() ([]LocalPackage, error) {
	request, _ := http.NewRequest("GET", fmt.Sprintf("%s/libraries", PythonServiceURL), nil)
	request.Header.Set("X-API-KEY", client.APIKey)

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unauthorized error from python service")
	}

	var packages []LocalPackage
	if err := json.NewDecoder(response.Body).Decode(&packages); err != nil {
		return nil, err
	}
	return packages, nil
}

func (client *PythonBridgeClient) InstallPackages(packageName string) (string, error) {
	payload := map[string]string{"name": packageName}
	jsonPayload, _ := json.Marshal(payload)

	request, _ := http.NewRequest("POST", fmt.Sprintf("%s/libraries", PythonServiceURL), bytes.NewBuffer(jsonPayload))
	request.Header.Set("X-API-KEY", client.APIKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return "", fmt.Errorf("install failed: %s", string(bodyBytes))
	}

	return string(bodyBytes), nil
}

func (client *PythonBridgeClient) DeletePackage(packageName string) (string, error) {
	request, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/libraries/%s", PythonServiceURL, packageName), nil)
	request.Header.Set("X-API-KEY", client.APIKey)

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return "", fmt.Errorf("uninstall failed: %s", string(bodyBytes))
	}

	return string(bodyBytes), nil
}
