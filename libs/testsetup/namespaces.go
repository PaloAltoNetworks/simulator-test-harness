package testsetup

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// MaxAporetoDepth is the maximum namespace depth in an Aporeto control plane.
const MaxAporetoDepth = 16

// An NSTree is a namespace hierarchy (a tree of namespaces). It can be readily parsed and
// serialized from / to yaml and json. Example yaml input:
//   name: root
//   tags: ["level=0", "foo=bar"]
//   children:
//   - name: c1
//     tags: []
//     children:
//     - name: g1
//       children: null
//   - name: c2
//     tags:
//     - "level=1"
//  	 children: []
type NSTree struct {
	// Name is the name of the namespace.
	Name string `json:"name" yaml:"name"`
	// Tags are the (user) tags attached to this namespace.
	Tags []string `json:"tags" yaml:"tags"`
	// Children are the children namespaces of this namespace.
	Children []NSTree `json:"children" yaml:"children"`
}

// NSDepth returns the depth of namespace ns, which should start with "/" and NOT end with "/".
func NSDepth(ns string) (int, error) {

	if !strings.HasPrefix(ns, "/") || (len(ns) > 1 && strings.HasSuffix(ns, "/")) {
		return -1, fmt.Errorf("invalid namespace %q", ns)
	}
	if ns == "/" {
		return 0, nil
	}

	return strings.Count(ns, "/"), nil
}

// CreateNSTree creates the namespace hierarchy defined in nst, in namespace ns. If nst is nil, this
// is a no-op.
func (c *Client) CreateNSTree(ns string, nst *NSTree, tagPrexifixes []string) error {

	// TODO OPT: Optimize by creating all children of a single ns with a single mctx
	// TODO OPT: Parallelize? (if we can make concurrent requests to the backend using m)

	if nst == nil {
		// Nothing to do
		return nil
	}
	// NOTE: path.Join does exactly what we want (always use '/', only root path ending in '/').
	newRoot := path.Join(ns, nst.Name)

	depth, err := NSDepth(newRoot)
	if err != nil {
		return errors.Wrap(err, "namespace depth")
	}

	if depth > MaxAporetoDepth {
		return fmt.Errorf("[CreateNSTree] max namespace level is %d, but got %d for namespace: %q", MaxAporetoDepth, depth, newRoot)
	}

	// Create the namespace gaia object
	n := gaia.NewNamespace()
	n.Name = nst.Name
	n.AssociatedTags = nst.Tags
	n.TagPrefixes = tagPrexifixes

	if err := c.ac.CreateInNS(ns, n); err != nil {
		return fmt.Errorf("create namespace %s: %v", n.Name, err)
	}

	// Create child namespaces recursively
	for _, cn := range nst.Children {
		if err := c.CreateNSTree(newRoot, &cn, tagPrexifixes); err != nil {
			return fmt.Errorf("create %s and children: %v", cn.Name, err)
		}
	}

	return nil
}

// DeleteNS deletes ns.
func (c *Client) DeleteNS(ns string) error {

	// NOTE: The options passed in the manipulator and the context are merged, with the context's
	// taking precedence.
	ctx := context.Background()
	if c.ac.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.ac.Timeout)
		defer cancel()
	}
	mctxopts := []manipulate.ContextOption{
		manipulate.ContextOptionNamespace(path.Dir(ns)),
		manipulate.ContextOptionFilter(elemental.NewFilterComposer().
			WithKey("name").Equals(ns).Done()),
	}
	mctx := manipulate.NewContext(ctx, mctxopts...)

	// Retrieve matching namespace object from the backend. NOTE: We could use
	// go.aporeto.io/underwater/core/nsutils/RetrieveByName to do the same here, but to use it we
	// would have to provide a manipulator with "OptionNamespace" set.
	n := gaia.NamespacesList{}
	if err := c.ac.Manipulator.RetrieveMany(mctx, &n); err != nil {
		return fmt.Errorf("retrieve %s: %v", ns, err)
	}

	if len(n) < 1 {
		return fmt.Errorf("namespace %s not found", ns)
	} else if len(n) > 1 {
		return fmt.Errorf("multiple namespaces named %s found", ns)
	}

	ctx = context.Background()
	if c.ac.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.ac.Timeout)
		defer cancel()
	}
	mctx = manipulate.NewContext(ctx, mctxopts...)
	if err := c.ac.Manipulator.Delete(mctx, n[0]); err != nil {
		return fmt.Errorf("delete %s: %v", ns, err)
	}

	return nil
}

// BasicNamespaceSetup creates a namespace, and basic external networks, and
// network policies:
// Allow ssh traffic towards the namespace
// Allow traffic from PU to all TCP/UDP
// Allow RDP towards the namespace
// Allow PU-to-PU communication in the namespace
// Propagatation to sub-namespaces and encryption are applied only to PU-to-PU
// policy.
func (c *Client) BasicNamespaceSetup(ns string, propagate bool) error {

	var err error

	// Create the test namespace.
	if err = c.CreateNSTree(path.Dir(ns), &NSTree{
		Name: path.Base(ns),
	},
		[]string{
			"externalnetwork:name=",
			"role=",
			"type=",
		},
	); err != nil {
		return fmt.Errorf("creating namespace %s: %v", ns, err)
	}

	_, err = c.CreateExternalNetwork("TCP/UDP all ports",
		ns,
		[]string{"externalnetwork:name=all"},
		[]string{QuadZeroRoute},
		false)
	if err != nil {
		return fmt.Errorf("create external network TCP/UDP all ports: %v", err)
	}

	_, err = c.CreateExternalNetwork("ssh",
		ns,
		[]string{"externalnetwork:name=ssh"},
		[]string{QuadZeroRoute},
		false)
	if err != nil {
		return fmt.Errorf("create external network UDP all ports: %v", err)
	}

	_, err = c.CreateExternalNetwork("rdp",
		ns,
		[]string{"externalnetwork:name=rdp"},
		[]string{QuadZeroRoute},
		false)
	if err != nil {
		return fmt.Errorf("create external network UDP all ports: %v", err)
	}

	incomings := []*gaia.NetworkRule{
		CreateNewNetworkRule("allow-ssh", gaia.NetworkRuleActionAllow,
			[][]string{{"externalnetwork:name=ssh"}},
			[]string{"tcp/22"}),
		CreateNewNetworkRule("allow-rdp", gaia.NetworkRuleActionAllow,
			[][]string{{"externalnetwork:name=rdp"}},
			[]string{"tcp/3389"}),
		CreateNewNetworkRule("allow-ptp", gaia.NetworkRuleActionAllow,
			[][]string{{"$identity=processingunit"}},
			[]string{"any"}),
	}

	outgoings := []*gaia.NetworkRule{
		CreateNewNetworkRule("allow-all", gaia.NetworkRuleActionAllow,
			[][]string{{"externalnetwork:name=all"}},
			[]string{"any"}),
		CreateNewNetworkRule("allow-ptp", gaia.NetworkRuleActionAllow,
			[][]string{{"$identity=processingunit"}},
			[]string{"any"}),
	}

	// Create the relevant network policies.
	_, err = c.CreateNetworkPolicy(
		"basic-setup",
		ns,
		[]string{"setup=basic"},
		[][]string{{"$namespace=" + ns}},
		propagate,
		incomings, outgoings)
	if err != nil {
		return fmt.Errorf("create network policy Allow ssh traffic: %v", err)
	}

	return nil
}
