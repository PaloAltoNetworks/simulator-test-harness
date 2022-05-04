// Package testsetup provides utilities for populating an Aporeto control plane with various
// objects: application credentials, namespaces, policies etc.
package testsetup

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/manipulate/maniphttp"
	"go.aporeto.io/simulator-test-harness/common"
	"go.aporeto.io/simulator-test-harness/libs/backend"
	"go.aporeto.io/simulator-test-harness/libs/internal/utils"
)

// TODO OPT: introduce more tags here (e.g. "automated", "benchmark" or "wontmatch")?
const (
	// RandomTag is the tag used for all random objects.
	RandomTag = "random=true"

	// QuadZeroRoute standard route for rules.
	QuadZeroRoute = "0.0.0.0/0"

	// AllTCPPorts represents TCP port 1 to 65,535.
	AllTCPPorts = "tcp/1:65535"

	// AllUDPPorts represents TCP port 1 to 65,535.
	AllUDPPorts = "udp/1:65535"

	// AllUDPPorts represents all ICMP message codes.
	AllICMP = "icmp"
)

// A Client is a client for setting up a test.
type Client struct {
	ac utils.APIClient
}

// Manipulator returns the client's underlying manipulator.
func (c *Client) Manipulator() manipulate.Manipulator {

	return c.ac.Manipulator
}

// Timeout returns the client's underlying timeout.
func (c *Client) Timeout() time.Duration {

	return c.ac.Timeout
}

// SetTimeout sets the client's underlying timeout.
func (c *Client) SetTimeout(timeout time.Duration) {

	c.ac.Timeout = timeout
}

// NewClient creates a new Client to use with the backend detailed in bd. opts are appended to the
// options used to create the manipulator.
func NewClient(bd *backend.Details, opts ...maniphttp.Option) (*Client, error) {

	ac, err := utils.NewAPIClient(bd, opts...)
	if err != nil {
		return nil, fmt.Errorf("create APIClient: %v", err)
	}
	return &Client{
		ac: *ac,
	}, nil
}

// randomPort returns a random port number (i.e. int in [1, 65535]).
func randomPort() int {
	return common.Roulette((1<<16)-1) + 1
}

// BackendClient returns a testsetup client from appcreds.
func BackendClient(appcreds string) (*Client, error) {

	details, err := backend.FromAppcred(appcreds)
	if err != nil {
		return nil, errors.Wrap(err, "read credentials")
	}

	// Create a client to connect to the backend.
	ac, err := NewClient(details)
	if err != nil {
		return nil, errors.Wrap(err, "create manipulator")
	}

	return ac, nil
}
