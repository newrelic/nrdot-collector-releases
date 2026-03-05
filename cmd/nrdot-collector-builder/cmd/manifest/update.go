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

		// Get version overrides from persistent flags.
		nrdotVersion := persistentFlag(cmd, "nrdot-version")
		coreStable := persistentFlag(cmd, "core-stable")
		coreBeta := persistentFlag(cmd, "core-beta")
		contribBeta := persistentFlag(cmd, "contrib-beta")

		// Build module prefix -> VersionUpdate map used by CopyAndUpdateConfigModules.
		nrdotUpdates := make(map[string]manifest.VersionUpdate)
		if nrdotVersion != "" {
			nrdotUpdates[manifest.NrModule] = manifest.VersionUpdate{BetaVersion: nrdotVersion}
		}
		if coreStable != "" || coreBeta != "" {
			nrdotUpdates[manifest.CoreModule] = manifest.VersionUpdate{StableVersion: coreStable, BetaVersion: coreBeta}
		}
		if contribBeta != "" {
			nrdotUpdates[manifest.ContribModule] = manifest.VersionUpdate{BetaVersion: contribBeta}
		}

		matches, _ := filepath.Glob(configPath)

		if len(matches) == 0 {
			fmt.Println("No files matched the pattern.")
			return nil
		}

		var currentVersions manifest.Versions
		var nextVersions manifest.Versions

		for _, match := range matches {
			cfg, err := loadConfig(match, verbose)
			if err != nil {
				return err
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

// persistentFlag returns the string value of a persistent flag, or "" if not found.
func persistentFlag(cmd *cobra.Command, name string) string {
	if f := cmd.Root().PersistentFlags().Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

// loadConfig reads a manifest YAML file and runs all required validation and
// initialisation steps, returning a ready-to-use Config.
func loadConfig(cfgFile string, verbose bool) (*manifest.Config, error) {
	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	cfg := &manifest.Config{Logger: log, Verbose: verbose}

	if verbose {
		log.Info("Using config file", zap.String("path", cfgFile))
	}

	k := koanf.New(".")
	if err = k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load configuration file: %w", err)
	}
	if err = k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "mapstructure"}); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	cfg.Path = cfgFile
	cfg.Dir = filepath.Dir(cfgFile)

	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	if err = cfg.SetGoPath(); err != nil {
		return nil, fmt.Errorf("go not found: %w", err)
	}
	if err = cfg.SetVersions(); err != nil {
		return nil, fmt.Errorf("versions not found: %w", err)
	}
	if err = cfg.ParseModules(); err != nil {
		return nil, fmt.Errorf("invalid module configuration: %w", err)
	}

	return cfg, nil
}
