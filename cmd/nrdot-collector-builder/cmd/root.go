// Copyright 2025 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var jsonOutput bool
var verbose bool
var nopexporterVersion string
var collectorCoreStableVersion string
var collectorContribBetaVersion string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nrdot-collector-builder",
	Short: "NRDOT client for building the OpenTelemetry Collector",
	Long: `
A CLI tool for building the OpenTelemetry Collector with NRDOT extensions.
This tool allows you to create a custom OpenTelemetry Collector binary with NRDOT extensions and configurations.
It simplifies the process of building and deploying the collector with NRDOT-specific features.
	`,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(
		&nopexporterVersion,
		"nopexporter-version",
		"",
		"github.com/newrelic/nrdot-collector-components/exporter/nopexporter",
	)
	rootCmd.PersistentFlags().StringVar(
		&collectorCoreStableVersion,
		"collector-core-stable",
		"",
		"go.opentelemetry.io/collector",
	)
	rootCmd.PersistentFlags().StringVar(
		&collectorContribBetaVersion,
		"collector-contrib-beta",
		"",
		"github.com/open-telemetry/opentelemetry-collector-contrib",
	)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}
