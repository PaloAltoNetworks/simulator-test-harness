package testsetup

import (
	"fmt"

	"go.aporeto.io/gaia"
	"go.aporeto.io/simulator-test-harness/common"
)

// CreateExternalNetwork creates an external network in namespace ns, identified by name and
// tagged with tags (user tags).
func (c *Client) CreateExternalNetwork(name, ns string, tags, networks []string,
	propagate bool) (*gaia.ExternalNetwork, error) {

	en := gaia.NewExternalNetwork()
	en.Name = name
	en.AssociatedTags = tags
	en.Entries = networks
	en.Propagate = propagate

	if err := c.ac.CreateInNS(ns, en); err != nil {
		return nil, fmt.Errorf("create external network %s: %v", name, err)
	}

	return en, nil
}

// CreateRandExternalNetwork creates a random (intended to match nothing) external network in
// namespace ns. Will create nTags random associated tags.
func (c *Client) CreateRandExternalNetwork(ns string, propagate bool,
	nTags int) (*gaia.ExternalNetwork, error) {

	name := "random-" + common.RandomString(6)
	var tags []string
	for i := 0; i < nTags; i++ {
		tags = append(tags, "random-"+common.RandomString(6))
	}
	networks := []string{name + "extnet.aporeto.com"}
	en, err := c.CreateExternalNetwork(name, ns, tags, networks, propagate)
	if err != nil {
		return nil, fmt.Errorf("create random extnet: %v", err)
	}
	return en, nil
}
