package main

import (
	"flag"
	"math"
	"path"

	"go.aporeto.io/benchmark-suite/common"
	"go.aporeto.io/benchmark-suite/libs/testsetup"
	"go.aporeto.io/benchmark-suite/utils/simulator/internal"
)

func main() {

	simulators := flag.Int("simulators", 300, "The number of simulators to map.")
	namespace := flag.String("namespace", "",
		"The base namespace to create the flat namespace hierarchy and/or mapping.")
	target := flag.String("target", "",
		"The target namespace in case of single mapping option.")
	prefix := flag.String("prefix", "",
		"The prefix to attach before the enforcer tag (prefix-$n-$m).")
	batch := flag.Int("batch", 300, "The batch size.")
	tag := flag.String("tag", "",
		"The tag to use for the enforcers batch, in case of single mapping option")
	capacity := flag.Int("capacity", 300, "The capacity of each namespace.")
	createNamespaces := flag.Bool("create-namespaces", false,
		"If set will create simulators/capacity namespaces.")
	multiMapping := flag.Bool("multi-mapping", false,
		"If set will create mapping policies from namespace to namespace/namespace-$i")
	singleMapping := flag.Bool("single-mapping", false,
		"If set will create one mapping policy from namespace to namespace/namespace-$i")
	appcred := flag.String("appcred", "apoctl.json",
		"Path to the backend configuration file.")
	publicCount := flag.Int("public", 0, "enforcers in public ns")
	pvtCount := flag.Int("private", 0, "enforcers in private ns")
	protectedCount := flag.Int("protected", 0, "enforcers in protected ns")

	flag.Parse()

	bc, err := internal.BackendFromAppcred(*appcred)
	if err != nil {
		common.Log.Fatalf("Unmarshal yaml file: %v", err)
	}

	mconf, err := testsetup.NewClient(bc)
	if err != nil {
		common.Log.Fatalf("Unable to create manipulator: %v", err)
	}

	if (*publicCount == 0) && (*pvtCount == 0) && (*protectedCount == 0) {

		numNamespace := int(math.Ceil(float64(*simulators) / float64(*capacity)))
		if *createNamespaces {
			err := internal.SimNSTree(mconf, *namespace, numNamespace)
			if err != nil {
				common.Log.Fatalf("creating namespaces: %v", err)
			}
		}

		if *multiMapping {
			if *batch > *simulators {
				common.Log.Fatalf("batch size cannot be larger than total enforcers")
			}

			if *batch == *capacity {
				err := internal.BToNPolicies(mconf, *namespace, numNamespace, *prefix)
				if err != nil {
					common.Log.Fatalf("creating batch mappings: %v", err)
				}
			} else {
				// Create numNamespace * n * m mapping policies from namespace to
				// namespace/namespace-$i
				err := internal.SimMappingPolicies(mconf, *namespace, *capacity,
					numNamespace, *batch, *prefix)
				if err != nil {
					common.Log.Fatalf("creating multi mappings: %v", err)
				}
			}
		}

		if *singleMapping {
			srcNamespace := *namespace
			trgNamespace := path.Join(srcNamespace, *target)
			err := internal.SimMappingPolicy(mconf, *tag, srcNamespace, trgNamespace)
			if err != nil {
				common.Log.Fatalf("creating one mapping policy: %v", err)
			}
		}
	} else {

		common.Log.Infof("Rails namespace model is detected with parameters public=%d, private=%d, protected=%d",
			*publicCount, *pvtCount, *protectedCount)
	}
}
