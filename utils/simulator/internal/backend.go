package internal

import (
	"fmt"
	"os"
	"path"

	midgardclient "go.aporeto.io/midgard-lib/client"
	"go.aporeto.io/simulator-test-harness/common"
	"go.aporeto.io/simulator-test-harness/libs/backend"
	"go.aporeto.io/simulator-test-harness/libs/testsetup"
)

// SimNSTree creates, given a "namespace", the "namespace" base and a flat
// hierarch with m namespaces under the base of "namespace".
func SimNSTree(mconf *testsetup.Client, namespace string,
	numNamespace int) error {

	common.Log.Info("Generating namespaces on aporeto for simulator invocation.")

	children := []testsetup.NSTree{}

	for i := 0; i < numNamespace; i++ {
		children = append(children, testsetup.NSTree{
			Name: fmt.Sprintf("%s-%d", path.Base(namespace), i),
			Tags: []string{
				fmt.Sprintf("child=%v", i),
				"creator=simulator-test-harness",
			},
		})
	}

	nstree := testsetup.NSTree{
		Name: path.Base(namespace),
		Tags: []string{
			"creator=simulator-test-harness",
		},
		Children: children,
	}

	common.Log.Debugf("The namespaces generated are: %+v", nstree)

	if err := mconf.CreateNSTree(path.Dir(namespace), &nstree, []string{}); err != nil {
		return fmt.Errorf("unable to create namespace: %v", err)
	}

	common.Log.Infof("%d namespaces generated under namespace %s", numNamespace,
		namespace)
	return nil
}

// SimMappingPolicy creates a mapping policy from srcNamespace to trgNamespace
// customized for simulator invocation.
func SimMappingPolicy(c *testsetup.Client, tag, srcNamespace,
	trgNamespace string) error {

	common.Log.Debugf("Creating a mapping policy with tag: %s", tag)

	nsmp, err := c.CreateNSMappingPolicy(
		"simulator-mapping",
		srcNamespace,
		[]string{
			"type=enforcer",
			"creator=simulator-test-harness",
		},
		[][]string{
			{
				"$identity=enforcer",
				tag,
			},
		},
		trgNamespace,
	)

	if err != nil {
		return fmt.Errorf("could not create mapping policy: %v", err)
	}

	common.Log.Debugf("Created namespace mapping policy: %v", *nsmp)
	return nil
}

// SimMappingPolicies creates many mapping policies from scrNamespace towards
// 'numNamespaces' sub-namespaces named srcNamespace-$i. It searches with tag
// "nsim=prefix$j-$i" where $j ranges through the batches needed and $i through
// the batchSize.
func SimMappingPolicies(c *testsetup.Client, srcNamespace string,
	capacity, numNamespaces, batchSize int, prefix string) error {

	common.Log.Info("Applying mapping policies")
	ns := 0
	mappings := 0
	batch := 0
	for {
		for i := 1; i <= batchSize; i++ {

			tag := fmt.Sprintf("nsim=%s%d-%d", prefix, batch, i)
			trgNamespace := path.Join(srcNamespace,
				fmt.Sprintf("%s-%d", path.Base(srcNamespace), ns))

			if err := SimMappingPolicy(c, tag, srcNamespace, trgNamespace); err != nil {
				return fmt.Errorf("mapping policies: %v", err)
			}
			mappings++
			if mappings%capacity == 0 {
				ns++
				if ns >= numNamespaces {
					common.Log.Infof("%d mapping policies are applied", mappings)
					return nil
				}
			}
		}
		batch++
	}
}

// BackendFromAppcred creates a backend.Details instance with only API.URL
// and application credentials as elements, which are read from appcred.
func BackendFromAppcred(appcred string) (*backend.Details, error) {

	data, err := os.ReadFile(appcred)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", appcred, err)
	}

	creds, _, err := midgardclient.ParseCredentials(data)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %v", err)
	}

	var bc backend.Details
	bc.API.URL = creds.APIURL
	bc.API.AppCred = appcred

	return &bc, nil
}

// BToNPolicies creates 1 mapping policy per batch with tag "simbatch=$prefix-$i"
// where i ranges through numNamespaces.
func BToNPolicies(c *testsetup.Client, srcNamespace string,
	numNamespaces int, prefix string) error {

	common.Log.Infof("Applying %d mapping policies", numNamespaces)
	for batch := 1; batch <= numNamespaces; batch++ {

		tag := fmt.Sprintf("simbase=%s-%d", prefix, batch)
		trgNamespace := path.Join(srcNamespace,
			fmt.Sprintf("%s-%d", path.Base(srcNamespace), batch-1))

		if err := SimMappingPolicy(c, tag, srcNamespace, trgNamespace); err != nil {
			return fmt.Errorf("mapping policies: %v", err)
		}
	}

	common.Log.Info("Mapping policies are applied")
	return nil
}
