package main

import (
	"fmt"
	"time"

	"go.aporeto.io/gaia"
	"go.aporeto.io/simulator-test-harness/common"
)

var protocols = []int{
	1,  // ICMP
	6,  // TCP
	17, // UDP
	41, // IPv6
	47, // GRE
}

// A PlanLayout is the layout of the plan.
type PlanLayout struct {
	Plan Plan `yaml:"plan"`
}

// A Plan to simulate.
type Plan struct {
	Lifecycle *Lifecycle `yaml:"lifecycle"`
	Jitter    *Jitter    `yaml:"jitter,omitempty"`
	Nodes     []*Node    `yaml:"nodes"`
}

// A Lifecycle represents how we want the simulator to behave in terms of
// iterations, interval, bursts, burst-sizes, etc.
type Lifecycle struct {
	// PUIterations can be "infinite".
	PUIterations string        `yaml:"pu-iterations"`
	PUInterval   time.Duration `yaml:"pu-interval,omitempty"`
	PUCleanup    time.Duration `yaml:"pu-cleanup,omitempty"`
	// FlowIterations can be "infinite".
	FlowIterations string        `yaml:"flow-iterations"`
	FlowInterval   time.Duration `yaml:"flow-interval,omitempty"`
	DNSReportRate  string        `yaml:"dns-report-rate,omitempty"`
}

// A Jitter stores different delay parameters for lifecycle operations.
// Jitters are simulated for node, PU & flow cycles.
type Jitter struct {
	Variance   string        `yaml:"variance,omitempty"`
	PUStart    time.Duration `yaml:"pu-start,omitempty"`
	PUReport   time.Duration `yaml:"pu-report,omitempty"`
	FlowReport time.Duration `yaml:"flow-report,omitempty"`
}

// A Node represent a node, Pu or ExtNet.
type Node struct {
	ID              string                `yaml:"ID"`
	Type            string                `yaml:"type"`
	IP              string                `yaml:"IP"`
	ExternalNetwork *gaia.ExternalNetwork `yaml:"externalNetwork"`
	ProcessingUnit  *gaia.ProcessingUnit  `yaml:"processingUnit"`
	Edges           *Edges                `yaml:"edges,omitempty"`
}

// Edges represent the edges definition.
type Edges struct {
	Frequency string  `yaml:"frequency,omitempty"`
	Flows     []*Flow `yaml:"flows"`
}

// A Flow represents a flow.
type Flow struct {
	Report *gaia.FlowReport `yaml:"report"`
	To     string           `yaml:"to"`
}

// puType returns cType if it is a valid gaia.ProcessingUnitTypeValue, else it returns a random
// gaia.ProcessingUnitTypeValue
func puType(cType string) gaia.ProcessingUnitTypeValue {

	puTypes := []gaia.ProcessingUnitTypeValue{
		gaia.ProcessingUnitTypeDocker,
		gaia.ProcessingUnitTypeHost,
		gaia.ProcessingUnitTypeHostService,
		gaia.ProcessingUnitTypeLinuxService,
		gaia.ProcessingUnitTypeSSHSession,
	}

	for _, t := range puTypes {
		if gaia.ProcessingUnitTypeValue(cType) == t {
			return t
		}
	}
	return puTypes[common.Roulette(len(puTypes))]
}

// generate does the plan generation, according to c.
func generate(c *Config) *PlanLayout {

	plan := Plan{
		Lifecycle: &c.Lifecycle,
		Jitter:    &c.Jitter,
	}

	// Generate PUs
	plan.Nodes = make([]*Node, c.PUs)
	for i := range plan.Nodes {
		name := fmt.Sprintf("%s-%d", c.Name, i+1)
		plan.Nodes[i] = &Node{
			ID:   fmt.Sprintf("%s-pu", name),
			Type: gaia.ProcessingUnitIdentity.Name,
			IP:   common.RandIP(),
		}

		pu := gaia.NewProcessingUnit()
		pu.Name = name
		pu.Type = puType(c.PUType)
		if c.PUMeta == nil {
			pu.Metadata = []string{
				fmt.Sprintf("@sys:image=%s-image", name),
				fmt.Sprintf("@usr:app=%s-app", name),
				fmt.Sprintf("@usr:key=%s-key", name),
			}
		} else {
			pu.Metadata = c.PUMeta
		}
		// NOTE: These two are not necessary, as the simulator (currently) does not use these fields
		// (i.e. all PUs are active and running, no matter what is specified here).
		plan.Nodes[i].ProcessingUnit = pu
	}

	// Generate flows
	// list of valid service types that we expect to see.
	serviceTypes := []gaia.FlowReportServiceTypeValue{
		gaia.FlowReportServiceTypeHTTP,
		gaia.FlowReportServiceTypeL3,
		gaia.FlowReportServiceTypeTCP,
	}
	lenServiceTypes := len(serviceTypes)

	actions := []gaia.FlowReportActionValue{
		gaia.FlowReportActionReject,
		gaia.FlowReportActionAccept,
	}
	lenActions := len(actions)

	observedActions := []gaia.FlowReportObservedActionValue{
		gaia.FlowReportObservedActionReject,
		gaia.FlowReportObservedActionAccept,
	}

	lenProtocols := len(protocols)

	for _, node := range plan.Nodes {
		node.Edges = &Edges{
			Flows: make([]*Flow, c.Flows),
		}
		for i := range node.Edges.Flows {
			node.Edges.Flows[i] = &Flow{
				To: plan.Nodes[i%len(plan.Nodes)].ID,
			}

			fr := gaia.NewFlowReport()
			fr.ServiceType = serviceTypes[common.Roulette(lenServiceTypes)]

			axn := common.Roulette(lenActions)
			fr.Action = actions[axn]
			fr.ObservedAction = observedActions[axn]

			fr.DestinationPort = 1025 + common.Roulette(64_000)
			fr.Protocol = protocols[common.Roulette(lenProtocols)]
			node.Edges.Flows[i].Report = fr
		}
	}

	return &PlanLayout{plan}
}
