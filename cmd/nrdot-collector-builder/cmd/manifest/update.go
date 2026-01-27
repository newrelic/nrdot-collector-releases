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

		// Get nopexporter version information from persistent flags
		nopexporterVersionFlag := cmd.Root().PersistentFlags().Lookup("nopexporter-version")
		nopexporterVersion := nopexporterVersionFlag.Value.String()
		nopexporterUsage := nopexporterVersionFlag.Usage

		collectorCoreStableFlag := cmd.Root().PersistentFlags().Lookup("collector-core-stable")
		collectorCoreStable := collectorCoreStableFlag.Value.String()
		collectorCoreStableUsage := collectorCoreStableFlag.Usage

		collectorContribBetaFlag := cmd.Root().PersistentFlags().Lookup("collector-contrib-beta")
		collectorContribBeta := collectorContribBetaFlag.Value.String()
		collectorContribBetaUsage := collectorContribBetaFlag.Usage

		// Create a map with module path (usage) as key and version array as value for updates
		nopexporterUpdates := make(map[string][]string)
		if nopexporterVersion != "" {
			nopexporterUpdates[nopexporterUsage] = []string{nopexporterVersion}
		}
		if collectorCoreStable != "" {
			nopexporterUpdates[collectorCoreStableUsage] = []string{collectorCoreStable}
		}
		if collectorContribBeta != "" {
			nopexporterUpdates[collectorContribBetaUsage] = []string{collectorContribBeta}
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
			if len(nopexporterUpdates) > 0 {
				updatedCfg, err = manifest.CopyAndUpdateConfigModules(cfg, nopexporterUpdates)
				if err != nil {
					return fmt.Errorf("failed to update configuration with nopexporter versions: %w", err)
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
				Nopexporter     *struct {
					NopexporterVersion   string `json:"nopexporterVersion,omitempty"`
					CollectorCoreStable  string `json:"collectorCoreStable,omitempty"`
					CollectorContribBeta string `json:"collectorContribBeta,omitempty"`
				} `json:"nopexporter,omitempty"`
			}{
				NextVersions:    nextVersions,
				CurrentVersions: currentVersions,
			}

			// Include nopexporter information if provided
			if nopexporterVersion != "" || collectorCoreStable != "" || collectorContribBeta != "" {
				output.Nopexporter = &struct {
					NopexporterVersion   string `json:"nopexporterVersion,omitempty"`
					CollectorCoreStable  string `json:"collectorCoreStable,omitempty"`
					CollectorContribBeta string `json:"collectorContribBeta,omitempty"`
				}{
					NopexporterVersion:   nopexporterVersion,
					CollectorCoreStable:  collectorCoreStable,
					CollectorContribBeta: collectorContribBeta,
				}
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
