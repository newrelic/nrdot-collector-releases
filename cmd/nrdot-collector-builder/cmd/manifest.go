// Copyright New Relic, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"newrelic-collector-builder/cmd/manifest"

	"github.com/spf13/cobra"
)

var manifestConfigPath string

// manifestCmd represents the manifest command
var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Manage the OCB manifest file",
	Long: `
	The manifest command allows you to manage the OCB manifest file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("manifest called")
	},
}

func init() {
	rootCmd.AddCommand(manifestCmd)
	// Register the update subcommand
	manifestCmd.AddCommand(manifest.UpdateCmd)

	// Define a persistent flag for `manifestCmd`
	manifestCmd.PersistentFlags().StringVarP(
		&manifestConfigPath,
		"config",
		"c",
		"",
		"Path to the manifest configuration file",
	)
}
