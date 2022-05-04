package main

import (
	"flag"
	"fmt"
	"os"

	"go.aporeto.io/simulator-test-harness/common"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var log = common.Log

// A Config is the configuration for the plan generation parameters.
type Config struct {
	Name      string    `yaml:"name"`
	PUs       int       `yaml:"pus"`
	PUType    string    `yaml:"pu-type"`
	PUMeta    []string  `yaml:"pu-meta"`
	Flows     int       `yaml:"flows"`
	Lifecycle Lifecycle `yaml:"lifecycle"`
	Jitter    Jitter    `yaml:"jitter"`
}

func main() {

	var configFile string
	var planFile string
	flag.StringVar(&configFile, "config", "config.yaml",
		"Set the path to the test configuration file")
	flag.StringVar(&planFile, "output", "plan.yaml",
		"Set the path to the test configuration file")

	// Log level parameters
	logLevel := flag.String("log-level", log.Level.String(),
		fmt.Sprintf("Set the logger log level, accepts one of: %v", logrus.AllLevels),
	)

	flag.Parse()

	lvl, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("parse log level %s: %v", *logLevel, err)
	}
	log.SetLevel(lvl)

	var config Config
	if err = common.ParseYamlFile(configFile, &config); err != nil {
		log.Fatalf("unable to parse the %q to config: %v", configFile, err)
	}

	log.Debugf("The configuration read from %q: %v", configFile, config)

	if config.Name == "" {
		config.Name = "auto-generate-plan"
	}

	plan := generate(&config)
	planData, err := yaml.Marshal(plan)
	if err != nil {
		log.Fatalf("marshal plan: %v", err)
	}

	if err := os.WriteFile(planFile, planData, 0644); err != nil {
		log.Fatalf("write plan to %q: %v", planFile, err)
	}
}
