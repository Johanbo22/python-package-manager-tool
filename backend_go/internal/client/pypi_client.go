package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PyPIPackageInfo struct {
	Info struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		Summary    string `json:"summary"`
		Author     string `json:"author"`
		HomeParams string `json:"home_page"`
	} `json:"info"`
}

func FetchPackageFromPyPI(packageName string) (*PyPIPackageInfo, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)

	response, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == 404 {
		return nil, fmt.Errorf("package not found")
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("pypi api returned status: %d", response.StatusCode)
	}

	var packageInfo PyPIPackageInfo
	if err := json.NewDecoder(response.Body).Decode(&packageInfo); err != nil {
		return nil, err
	}

	return &packageInfo, nil

}
