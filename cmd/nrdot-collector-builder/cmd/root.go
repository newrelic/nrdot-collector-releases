// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput         bool
	verbose            bool
	nrdotVersion       string
	coreStableVersion  string
	coreBetaVersion    string
	contribBetaVersion string
)

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
	rootCmd.PersistentFlags().StringVar(&nrdotVersion, "nrdot-version", "", "Pin nrdot-collector-components to this version")
	rootCmd.PersistentFlags().StringVar(&coreStableVersion, "core-stable", "", "Pin OTel core stable (v1.x) modules to this version")
	rootCmd.PersistentFlags().StringVar(&coreBetaVersion, "core-beta", "", "Pin OTel core beta (v0.x) modules to this version")
	rootCmd.PersistentFlags().StringVar(&contribBetaVersion, "contrib-beta", "", "Pin OTel contrib modules to this version")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
