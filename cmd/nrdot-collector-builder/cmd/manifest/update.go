// Copyright New Relic, Inc. All rights reserved.
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
		var nrdotVersion string
		if f := cmd.Root().PersistentFlags().Lookup("nrdot-version"); f != nil {
			nrdotVersion = f.Value.String()
		}

		var coreStable string
		if f := cmd.Root().PersistentFlags().Lookup("core-stable"); f != nil {
			coreStable = f.Value.String()
		}

		var coreBeta string
		if f := cmd.Root().PersistentFlags().Lookup("core-beta"); f != nil {
			coreBeta = f.Value.String()
		}

		var contribBeta string
		if f := cmd.Root().PersistentFlags().Lookup("contrib-beta"); f != nil {
			contribBeta = f.Value.String()
		}

		// Build module prefix → version map. Keys use "prefix:stability" so that
		// CopyAndUpdateConfigModules can match stable vs beta core modules separately.
		nrdotUpdates := make(map[string][]string)
		if nrdotVersion != "" {
			nrdotUpdates["github.com/newrelic/nrdot-collector-components"] = []string{nrdotVersion}
		}
		if coreStable != "" {
			nrdotUpdates["go.opentelemetry.io/collector:stable"] = []string{coreStable}
		}
		if coreBeta != "" {
			nrdotUpdates["go.opentelemetry.io/collector:beta"] = []string{coreBeta}
		}
		if contribBeta != "" {
			nrdotUpdates["github.com/open-telemetry/opentelemetry-collector-contrib"] = []string{contribBeta}
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
