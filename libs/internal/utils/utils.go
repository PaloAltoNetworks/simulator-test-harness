package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"go.aporeto.io/elemental"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/manipulate/maniphttp"
	midgardclient "go.aporeto.io/midgard-lib/client"
	"go.aporeto.io/simulator-test-harness/libs/backend"
)

// An APIClient is a client for communicating with an Aporeto API.
type APIClient struct {
	// Manipulator is the manipulator used for addressing the backend.
	Manipulator manipulate.Manipulator
	// Timeout is the timeout to specify for all Manipulator operations, if greater than zero.
	Timeout time.Duration
}

// NewAPIClient creates a new APIClient to use with the backend detailed in bd. opts are appended to
// the options used to create the manipulator.
func NewAPIClient(bd *backend.Details, opts ...maniphttp.Option) (*APIClient, error) {

	// Determine the token to use
	var err error
	token := bd.API.Token
	if token == "" {
		token, err = tokenFromAppcred(bd.API.AppCred)
		if err != nil {
			return nil, fmt.Errorf("get token from appcred: %v", err)
		}
	}

	// Create the manipulator
	mopts := []maniphttp.Option{
		maniphttp.OptionToken(token),
		maniphttp.OptionTLSConfig(&tls.Config{
			InsecureSkipVerify: true, // NOTE: Skipping API verification by default
		}),
		maniphttp.OptionEncoding(elemental.EncodingTypeMSGPACK),
	}
	mopts = append(mopts, opts...)

	m, err := maniphttp.New(context.Background(), bd.API.URL, mopts...)
	if err != nil {
		return nil, fmt.Errorf("create maniphttp: %v", err)
	}

	return &APIClient{
		Manipulator: m,
	}, nil
}

// CreateInNS creates object in namespace ns, prepending mctxopts to the options used to create
// the manipulator's context.
func (ac *APIClient) CreateInNS(ns string, object elemental.Identifiable,
	mctxopts ...manipulate.ContextOption) error {

	// NOTE: The options passed in the manipulator and the context are merged, with the context's
	// taking precedence.
	ctx := context.Background()
	if ac.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ac.Timeout)
		defer cancel()
	}
	mctxopts = append(mctxopts, manipulate.ContextOptionNamespace(ns))
	mctx := manipulate.NewContext(ctx, mctxopts...)

	// Create the object on the backend
	if err := ac.Manipulator.Create(mctx, object); err != nil {
		return err
	}

	return nil
}

// DeleteInNS deletes object from namespace ns, prepending mctxopts to the options used to create
// the manipulator's context.
func (ac *APIClient) DeleteInNS(ns string, object elemental.Identifiable,
	mctxopts ...manipulate.ContextOption) error {

	// NOTE: The options passed in the manipulator and the context are merged, with the context's
	// taking precedence.
	ctx := context.Background()
	if ac.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ac.Timeout)
		defer cancel()
	}
	mctxopts = append(mctxopts, manipulate.ContextOptionNamespace(ns))
	mctx := manipulate.NewContext(ctx, mctxopts...)

	// Delete the object on the backend
	if err := ac.Manipulator.Delete(mctx, object); err != nil {
		return err
	}

	return nil
}

// tokenFromAppcred generates a jwt for API authentication, using the application credential in the
// file specified in appcred.
func tokenFromAppcred(appcred string) (string, error) {

	const (
		maxDuration    time.Duration = 1<<63 - 1
		midgardTimeout time.Duration = 1 * time.Minute
	)

	b, err := os.ReadFile(appcred)
	if err != nil {
		return "", fmt.Errorf("read %s: %v", appcred, err)
	}

	creds, _, err := midgardclient.ParseCredentials(b)
	if err != nil {
		return "", fmt.Errorf("parse credentials: %v", err)
	}

	tlsConfig, err := midgardclient.CredsToTLSConfig(creds)
	if err != nil {
		return "", fmt.Errorf("get tls config from credentials: %v", err)
	}

	client := midgardclient.NewClientWithTLS(creds.APIURL, tlsConfig)
	ctx, cancel := context.WithTimeout(context.Background(), midgardTimeout)
	defer cancel()
	token, err := client.IssueFromCertificate(ctx, maxDuration)
	if err != nil {
		return "", fmt.Errorf("issue token from certificate: %v", err)
	}

	return token, nil
}
