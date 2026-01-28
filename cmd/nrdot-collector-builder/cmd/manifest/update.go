// Copyright 2025 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"newrelic-collector-builder/internal/manifest"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	"golang.org/x/mod/semver"

	"github.com/knadh/koanf/providers/file"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// UpdateCmd represents the `manifest update` subcommand
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the manifest file",
	Long:  "Update the manifest file to ensure otel components are up to date.",

	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")

		jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
		verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose")

		// Get nrdot version information from persistent flags
		nrdotVersionFlag := cmd.Root().PersistentFlags().Lookup("nrdot-version")
		nrdotVersion := nrdotVersionFlag.Value.String()
		nrdotUsage := nrdotVersionFlag.Usage

		coreStableFlag := cmd.Root().PersistentFlags().Lookup("core-stable")
		coreStable := coreStableFlag.Value.String()
		coreStableUsage := coreStableFlag.Usage

		contribBetaFlag := cmd.Root().PersistentFlags().Lookup("contrib-beta")
		contribBeta := contribBetaFlag.Value.String()
		contribBetaUsage := contribBetaFlag.Usage

		// Create a map with module path (usage) as key and version array as value for updates
		nrdotUpdates := make(map[string][]string)
		if nrdotVersion != "" {
			nrdotUpdates[nrdotUsage] = []string{nrdotVersion}
		}
		if coreStable != "" {
			nrdotUpdates[coreStableUsage] = []string{coreStable}
		}
		if contribBeta != "" {
			nrdotUpdates[contribBetaUsage] = []string{contribBeta}
		}

		matches, _ := filepath.Glob(configPath)

		if len(matches) == 0 {
			fmt.Println("No files matched the pattern.")
			return nil
		}

		var currentVersions manifest.Versions
		var nextVersions manifest.Versions

		for _, match := range matches {
			cfg, _, err := initConfig(match, verbose)

			if err != nil {
				return err
			}

			if err = cfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			if err = cfg.SetGoPath(); err != nil {
				return fmt.Errorf("go not found: %w", err)
			}

			if err = cfg.SetVersions(); err != nil {
				return fmt.Errorf("versions not found: %w", err)
			}

			if err = cfg.ParseModules(); err != nil {
				return fmt.Errorf("invalid module configuration: %w", err)
			}

			if (currentVersions.BetaCoreVersion == "") || semver.Compare(cfg.Versions.BetaCoreVersion, currentVersions.BetaCoreVersion) > 0 {
				currentVersions = cfg.Versions
			}

			var updatedCfg *manifest.Config
			if len(nrdotUpdates) > 0 {
				updatedCfg, err = manifest.CopyAndUpdateConfigModules(cfg, nrdotUpdates)
				if err != nil {
					return fmt.Errorf("failed to update configuration with nrdot versions: %w", err)
				}
			} else {
				updatedCfg, err = manifest.UpdateConfigModules(cfg)
				if err != nil {
					return fmt.Errorf("failed to update configuration: %w", err)
				}
			}

			if err = manifest.WriteConfigFile(updatedCfg); err != nil {
				return fmt.Errorf("failed to write configuration file: %w", err)
			}

			if (nextVersions.BetaCoreVersion == "") || semver.Compare(updatedCfg.Versions.BetaCoreVersion, nextVersions.BetaCoreVersion) > 0 {
				nextVersions = updatedCfg.Versions
			}
		}

		if jsonOutput {
			// print JSON output of all otel versions
			output := struct {
				NextVersions    manifest.Versions `json:"nextVersions"`
				CurrentVersions manifest.Versions `json:"currentVersions"`
			}{
				NextVersions:    nextVersions,
				CurrentVersions: currentVersions,
			}

			b, err := json.Marshal(output)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON output: %w", err)
			}
			fmt.Println(string(b))
		}

		return nil

	},
}

func initConfig(cfgFile string, verbose bool) (*manifest.Config, *koanf.Koanf, error) {
	var err error
	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	cfg := &manifest.Config{
		Logger:  log,
		Verbose: verbose,
	}

	if cfg.Verbose {
		cfg.Logger.Info("Using config file", zap.String("path", cfgFile))
	}
	// load the config file
	provider := file.Provider(cfgFile)

	k := koanf.New(".")

	if err = k.Load(provider, yaml.Parser()); err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration file: %w", err)
	}

	if err = k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "mapstructure"}); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	cfg.Path = cfgFile
	cfg.Dir = filepath.Dir(cfgFile)

	return cfg, k, nil
}
