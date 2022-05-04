package backend

import (
	"fmt"
	"os"

	midgardclient "go.aporeto.io/midgard-lib/client"
	"gopkg.in/yaml.v3"
)

// An APIDetails is the access details for an Aporeto API.
type APIDetails struct {
	// URL is the url of the API.
	URL string `json:"url" yaml:"url"`
	// Token is the jwt to use for authentication. Either this or AppCred must be specified. If
	// specified, this will take precedence over AppCred.
	Token string `json:"token,omitempty" yaml:"token,omitempty"`
	// AppCred is the path to an application credential to use for authentication. Either this or
	// Token must be specified.
	AppCred string `json:"appcred,omitempty" yaml:"appcred,omitempty"`
}

// A MonitoringDetails is the access details for a backend monitoring stack.
type MonitoringDetails struct {
	// URL is the url of the monitoring stack (exposing Grafana).
	URL  string `json:"url" yaml:"url"`
	Cert struct {
		// Path is the path to the monitoring certificate.
		Path string `json:"path" yaml:"path"`
		// Password is the password for the certificate.
		Password string `json:"password" yaml:"password"`
	} `json:"cert" yaml:"cert"`
	// DatasourceID is the id of the Grafana data source to use for querying Prometheus. See
	// https://grafana.com/docs/http_api/data_source/ for details.
	DatasourceID int `json:"datasource_id" yaml:"datasource_id"`
}

// A Details is the access details for an Aporeto backend. It can be readily parsed and serialized
// from / to yaml and json. Example yaml input:
//   api:
//     url: "API URL"
//     # At least one of the two below must be specified. If both are, token takes precedence.
//     token: "provide your token here"
//     appcred: "path to appcred file"
//   monitoring:
//     url: "backend monitoring stack URL"
//     cert:
//       path: "the path to a file containing the certificate for accessing the monitoring stack"
//       password: "the password to use for the certificate"
//     datasource_id: 1 # the ID of the Grafana data source to use for querying Prometheus
type Details struct {
	API        APIDetails        `json:"api" yaml:"api"`
	Monitoring MonitoringDetails `json:"monitoring" yaml:"monitoring"`
}

// FromFile parses a Details from backendFile.
func FromFile(backendFile string) (*Details, error) {

	b, err := os.ReadFile(backendFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", backendFile, err)
	}
	var bd Details
	if err := yaml.Unmarshal(b, &bd); err != nil {
		return nil, fmt.Errorf("parse backend details: %v", err)
	}

	return &bd, nil
}

// ToFile parses a Details from backendFile.
func (d *Details) ToFile(backendFile string) error {

	data, err := yaml.Marshal(d)
	if err != nil {
		return fmt.Errorf("parse backend details: %v", err)
	}

	err = os.WriteFile(backendFile, data, 0644)
	if err != nil {
		return fmt.Errorf("write to %s: %v", backendFile, err)
	}

	return nil
}

// FromAppcred creates a backend.Details instance with only API.URL
// and application credentials as elements, which are read from appcred.
func FromAppcred(appcred string) (*Details, error) {

	data, err := os.ReadFile(appcred)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", appcred, err)
	}

	creds, _, err := midgardclient.ParseCredentials(data)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %v", err)
	}

	var bc Details
	bc.API.URL = creds.APIURL
	bc.API.AppCred = appcred

	return &bc, nil
}
