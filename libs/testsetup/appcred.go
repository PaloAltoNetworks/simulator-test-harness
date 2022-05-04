package testsetup

import (
	"context"
	"fmt"
	"os"

	"encoding/json"

	"go.aporeto.io/addedeffect/appcreds"
	"go.aporeto.io/gaia"
)

// CreateEnforcerAppCredential creates an enforcer application credential for namespace ns,
// identified by name.
func (c *Client) CreateEnforcerAppCredential(name, ns string) (*gaia.AppCredential, error) {

	roles := []string{
		"@auth:role=enforcer",
		"@auth:role=enforcer.runtime",
	}
	ctx := context.Background()
	if c.ac.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.ac.Timeout)
		defer cancel()
	}
	return appcreds.New(ctx, c.ac.Manipulator, ns, name, roles, nil)
}

// AppCredsToJSON writes appcreds to a file in JSON format.
func AppCredsToJSON(appcred *gaia.AppCredential, filename string) error {

	byteData, err := json.MarshalIndent(appcred.Credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("encode appcred: %v", err)
	}
	if err = os.WriteFile(filename, byteData, 0644); err != nil {
		return fmt.Errorf("write %s: %v", appcred, err)
	}
	return nil
}
