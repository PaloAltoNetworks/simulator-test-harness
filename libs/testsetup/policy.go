package testsetup

import (
	"fmt"

	"go.aporeto.io/gaia"
	"go.aporeto.io/simulator-test-harness/common"
)

const (
	// AuthorizedIdentityNamespaceAdministrator is the Namespace Administrator role.
	AuthorizedIdentityNamespaceAdministrator = "@auth:role=namespace.administrator"
	// AuthorizedIdentityNamespaceViewer is the Namespace Viewer role.
	AuthorizedIdentityNamespaceViewer = "@auth:role=namespace.viewer"
	// AuthorizedIdentityNamespaceContributor is the Namespace Contributor role.
	AuthorizedIdentityNamespaceContributor = "@auth:role=namespace.contributor"
	// AuthorizedIdentityNamespaceAuditor is the Namespace Auditor role.
	AuthorizedIdentityNamespaceAuditor = "@auth:role=namespace.auditor"
	// AuthorizedIdentityNamespaceImporter is the Namespace importer role.
	AuthorizedIdentityNamespaceImporter = "@auth:role=namespace.importer"
	// AuthorizedIdentityNamespaceExporter is the Namespace Exporter role.
	AuthorizedIdentityNamespaceExporter = "@auth:role=namespace.exporter"
	// AuthorizedIdentityStorageEditor is the Storage Editor role.
	AuthorizedIdentityStorageEditor = "@auth:role=storage.editor"
	// AuthorizedIdentityStorageViewer is the Storage Viewer role.
	AuthorizedIdentityStorageViewer = "@auth:role=storage.viewer"
	// AuthorizedIdentitySystemEditor is the System Editor role.
	AuthorizedIdentitySystemEditor = "@auth:role=system.editor"
	// AuthorizedIdentitySystemViewer si the System Viewer role.
	AuthorizedIdentitySystemViewer = "@auth:role=system.viewer"
	// AuthorizedIdentityAutomationEditor si the Automation Editor role.
	AuthorizedIdentityAutomationEditor = "@auth:role=automation.editor"
	// AuthorizedIdentityAutomationViewer is the Automation Viewer role.
	AuthorizedIdentityAutomationViewer = "@auth:role=automation.viewer"
	// AuthorizedIdentityAppsEditor is the Apps Editor role.
	AuthorizedIdentityAppsEditor = "@auth:role=apps.editor"
	// AuthorizedIdentityAppsViewer is the Apps Viewer role.
	AuthorizedIdentityAppsViewer = "@auth:role=apps.viewer"
	// AuthorizedIdentityEnforcer is the Enforcer role.
	AuthorizedIdentityEnforcer = "@auth:role=enforcer"
	// AuthorizedIdentityEnforcerRuntime is the Enforcer Runtime role.
	AuthorizedIdentityEnforcerRuntime = "@auth:role=enforcer.runtime"
	// AuthorizedIdentityAporetoOperator is the Aporeto Operator role.
	AuthorizedIdentityAporetoOperator = "@auth:role=aporeto-operator"
	// AuthorizedIdentityAppCredentials is the Application Credentials Creator role.
	AuthorizedIdentityAppCredentials = "@auth:role=appcredentials"
	// AuthorizedIdentitySSHIdentityRequester is the SSH Credentials Requester role.
	AuthorizedIdentitySSHIdentityRequester = "@auth:role=sshidentity.requester"
	// AuthorizedIdentitySSHIdentityEditor is the SSH Credentials Editor role.
	AuthorizedIdentitySSHIdentityEditor = "@auth:role=sshidentity.editor"
	// AuthorizedIdentitySSHIdentityViewer is the SSH Credentials Viewer role.
	AuthorizedIdentitySSHIdentityViewer = "@auth:role=sshidentity.viewer"
	// AuthorizedIdentityPlatformViewer is the Platform Viewer role.
	AuthorizedIdentityPlatformViewer = "@auth:role=platform.viewer"
	// AuthorizedIdentityDashboardViewer is the Dashboard Viewer role.
	AuthorizedIdentityDashboardViewer = "@auth:role=dashboard.viewer"
	// AuthorizedIdentityComputeEditor is the Compute Editor role.
	AuthorizedIdentityComputeEditor = "@auth:role=compute.editor"
	// AuthorizedIdentityComputeViewer is the Compute Viewer role.
	AuthorizedIdentityComputeViewer = "@auth:role=compute.viewer"
	// AuthorizedIdentityNetworkEditor is the Network Editor role.
	AuthorizedIdentityNetworkEditor = "@auth:role=network.editor"
	// AuthorizedIdentityNetworkViewer is the Network Viewer role.
	AuthorizedIdentityNetworkViewer = "@auth:role=network.viewer"
)

// CreateNewNetworkRule creates a new network rule object.
func CreateNewNetworkRule(name string, action gaia.NetworkRuleActionValue,
	object [][]string, protocolPorts []string) *gaia.NetworkRule {

	nr := gaia.NewNetworkRule()
	nr.Name = name
	nr.Action = action
	nr.LogsDisabled = false
	nr.Object = object
	nr.ProtocolPorts = protocolPorts

	return nr
}

// CreateNetworkPolicy creates a network access policy in namespace ns, identified by name and
// tagged with tags (user tags). The policy applies for both incoming and outgoing traffic.
func (c *Client) CreateNetworkPolicy(name, ns string, tags []string, subject [][]string,
	propagate bool, incomings, outgoings []*gaia.NetworkRule) (*gaia.NetworkRuleSetPolicy, error) {

	np := gaia.NewNetworkRuleSetPolicy()
	np.SetName(name)
	np.AssociatedTags = tags
	np.Propagate = propagate
	np.Subject = subject
	np.OutgoingRules = outgoings
	np.IncomingRules = incomings

	if err := c.ac.CreateInNS(ns, np); err != nil {
		return nil, fmt.Errorf("create network policy %s: %v", np.Name, err)
	}

	return np, nil
}

// CreateRandNetworkPolicy creates a random (intended to match nothing) network access policy in
// namespace ns.
func (c *Client) CreateRandNetworkPolicy(ns string,
	propagate bool) (*gaia.NetworkRuleSetPolicy, error) {

	tcpOrUDP := func() string {
		protocols := []string{"tcp", "udp"}
		return protocols[common.Roulette(len(protocols))]
	}

	randAction := func() gaia.NetworkRuleActionValue {
		actions := []gaia.NetworkRuleActionValue{
			gaia.NetworkRuleActionAllow,
			gaia.NetworkRuleActionReject,
		}
		return actions[common.Roulette(len(actions))]
	}

	name := "random-" + common.RandomString(6)
	tags := []string{RandomTag}
	src := [][]string{{"matchedby=" + name}}
	dest := [][]string{{"matchedby=" + name}}
	ports := []string{fmt.Sprintf("%s/%d", tcpOrUDP(), randomPort())}
	action := randAction()

	outgoings := []*gaia.NetworkRule{
		CreateNewNetworkRule(name, action,
			dest,
			ports)}

	incomings := []*gaia.NetworkRule{
		CreateNewNetworkRule(name, action,
			dest,
			ports)}

	np, err := c.CreateNetworkPolicy(name, ns, tags, src, propagate, incomings, outgoings)
	if err != nil {
		return nil, fmt.Errorf("create random network policy: %v", err)
	}
	return np, err
}

// CreateNSMappingPolicy creates a namespace mapping policy, mapping subject from ns to mapns,
// identified by name and tagged with tags (user tags).
func (c *Client) CreateNSMappingPolicy(name, ns string, tags []string, subject [][]string,
	mapns string) (*gaia.NamespaceMappingPolicy, error) {

	mp := gaia.NewNamespaceMappingPolicy()
	mp.Name = name
	mp.AssociatedTags = tags
	mp.MappedNamespace = mapns
	mp.Subject = subject

	if err := c.ac.CreateInNS(ns, mp); err != nil {
		return nil, fmt.Errorf("create ns mapping policy %s: %v", name, err)
	}

	return mp, nil
}

// CreateRandNSMappingPolicy creates a random (intended to match nothing) namespace mapping policy
// in namespace ns.
func (c *Client) CreateRandNSMappingPolicy(ns string) (*gaia.NamespaceMappingPolicy, error) {

	name := "random-" + common.RandomString(6)
	tags := []string{RandomTag}
	subject := [][]string{{"matchedby=" + name}}
	mapns := ns

	mp, err := c.CreateNSMappingPolicy(name, ns, tags, subject, mapns)
	if err != nil {
		return nil, fmt.Errorf("create random ns mapping policy: %v", err)
	}
	return mp, nil
}

// CreateAPIAuthPolicy creates an API authorization policy in namespace ns (applying to ns and its
// children), tagged with tags (user tags) and identified by name. subjectRealm is the realm with
// which the subject is authenticated and subjectAuthTags are the rest of the subject tags (without
// the "@auth:" prefix).
func (c *Client) CreateAPIAuthPolicy(name, ns string, tags, roles []string,
	subjectRealm gaia.IssueRealmValue, subjectAuthTags []string) (*gaia.APIAuthorizationPolicy,
	error) {

	aap := gaia.NewAPIAuthorizationPolicy()
	aap.Name = name
	aap.AssociatedTags = tags
	aap.AuthorizedIdentities = roles
	aap.AuthorizedNamespace = ns // NOTE: We set the target ns to the one the policy is created in.
	// Prepend the "@auth:" prefix to all elements in subjectAuthTags
	for i := range subjectAuthTags {
		subjectAuthTags[i] = "@auth:" + subjectAuthTags[i]
	}
	aap.Subject = [][]string{
		append([]string{"@auth:realm=" + string(subjectRealm)}, subjectAuthTags...),
	}

	if err := c.ac.CreateInNS(ns, aap); err != nil {
		return nil, fmt.Errorf("create API authorization policy %s: %v", name, err)
	}

	return aap, nil
}

// CreateRandAPIAuthPolicy creates a random (intended to match nothing) API authorization policy in
// namespace ns (applying to ns and its children).
func (c *Client) CreateRandAPIAuthPolicy(ns string) (*gaia.APIAuthorizationPolicy, error) {

	randID := common.RandomString(6)
	name := "random-" + randID
	tags := []string{RandomTag}
	roles := []string{"@auth:role=" + randID}
	realm := gaia.IssueRealmValue("random")
	subjectTags := []string{"matchedBy=" + randID}

	aap, err := c.CreateAPIAuthPolicy(name, ns, tags, roles, realm, subjectTags)
	if err != nil {
		return nil, fmt.Errorf("create random API authorization policy: %v", err)
	}
	return aap, nil
}

// TODO OPT: Move this to a separate package under simulator-test-harness/libs? (as it's not really setup)

// RenderPolicies returns the policies that would apply to a PU in namespace ns, identified by name
// and tagged with tags (user tags).
func (c *Client) RenderPolicies(name, ns string, tags []string,
	puType gaia.ProcessingUnitTypeValue) (*gaia.RenderedPolicy, error) {

	pu := gaia.NewProcessingUnit()
	pu.Name = name
	pu.AssociatedTags = tags
	pu.Type = gaia.ProcessingUnitTypeHost

	rp := gaia.NewRenderedPolicy()
	rp.ProcessingUnit = pu
	rp.ProcessingUnitTags = tags

	if err := c.ac.CreateInNS(ns, rp); err != nil {
		return nil, fmt.Errorf("render policies for PU %s: %v", name, err)
	}

	return rp, nil
}

// CreateHSMappingPolicy creates a host service mapping policy in namespace ns.
func (c *Client) CreateHSMappingPolicy(name, ns string, tags,
	hostservices [][]string) (*gaia.HostServiceMappingPolicy, error) {

	hsp := gaia.NewHostServiceMappingPolicy()
	hsp.Name = name
	hsp.Object = hostservices
	hsp.Subject = tags

	if err := c.ac.CreateInNS(ns, hsp); err != nil {
		return nil, fmt.Errorf("create host service mapping policy %s: %v", name, err)
	}

	return hsp, nil
}

// DeleteHSMappingPolicy removes a host service mapping policy from namespace
// ns, identified by id.
func (c *Client) DeleteHSMappingPolicy(hsmp *gaia.HostServiceMappingPolicy, ns string) error {

	if err := c.ac.DeleteInNS(ns, hsmp); err != nil {
		return fmt.Errorf("delete host service mapping policy %s from ns %s: %v",
			hsmp.Name, ns, err)
	}

	return nil
}
