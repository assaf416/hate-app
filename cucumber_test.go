package main

import (
	"os"
	"testing"

	"github.com/cucumber/godog"

	stepdefinitions "insurance/features/step_definitions"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "insurance",
		ScenarioInitializer: stepdefinitions.InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		os.Exit(1)
	}
}
