// Copyright 2025 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

// errMissingGoMod indicates an empty gomod field
var errMissingGoMod = errors.New("missing gomod specification for module")

const coreModule = "go.opentelemetry.io/collector"
const contribModule = "github.com/open-telemetry/opentelemetry-collector-contrib"
const nrModule = "github.com/newrelic/nrdot-collector-components"

type Versions struct {
	BetaCoreVersion    string `json:"betaCoreVersion"`
	BetaContribVersion string `json:"betaContribVersion"`
	StableCoreVersion  string `json:"stableCoreVersion"`
	NrdotVersion 		   string `json:"nrdotVersion"`
}

// Config holds the builder's configuration
type Config struct {
	Logger *zap.Logger
	Path   string `mapstructure:"-"` // path to the config file
	Dir    string `mapstructure:"-"` // path to the config directory

	Versions Versions  `mapstructure:"-"` // only used be the go.mod template
	Verbose  bool      `mapstructure:"-"`
	YamlNode yaml.Node `mapstructure:"-"`

	Distribution      Distribution `mapstructure:"dist"`
	Exporters         []Module     `mapstructure:"exporters"`
	Extensions        []Module     `mapstructure:"extensions"`
	Receivers         []Module     `mapstructure:"receivers"`
	Processors        []Module     `mapstructure:"processors"`
	Connectors        []Module     `mapstructure:"connectors"`
	ConfmapProviders  []Module     `mapstructure:"providers"`
	ConfmapConverters []Module     `mapstructure:"converters"`
	Replaces          []string     `mapstructure:"replaces"`
	Excludes          []string     `mapstructure:"excludes"`
}

type ConfResolver struct {
	// When set, will be used to set the CollectorSettings.ConfResolver.DefaultScheme value,
	// which determines how the Collector interprets URIs that have no scheme, such as ${ENV}.
	// See https://pkg.go.dev/go.opentelemetry.io/collector/confmap#ResolverSettings for more details.
	DefaultURIScheme string `mapstructure:"default_uri_scheme"`
}

// Distribution holds the parameters for the final binary
type Distribution struct {
	Module           string `mapstructure:"module"`
	Name             string `mapstructure:"name"`
	Go               string `mapstructure:"go"`
	Description      string `mapstructure:"description"`
	OutputPath       string `mapstructure:"output_path"`
	Version          string `mapstructure:"version"`
	BuildTags        string `mapstructure:"build_tags"`
	DebugCompilation bool   `mapstructure:"debug_compilation"`
}

// Module represents a receiver, exporter, processor or extension for the distribution
type Module struct {
	Name   string `mapstructure:"name"`   // if not specified, this is package part of the go mod (last part of the path)
	Import string `mapstructure:"import"` // if not specified, this is the path part of the go mods
	GoMod  string `mapstructure:"gomod"`  // a gomod-compatible spec for the module
	Path   string `mapstructure:"path"`   // an optional path to the local version of this module
}

func isOtelCoreComponent(mod string) bool {
	// Check if the component is part of the OpenTelemetry Collector core
	if strings.HasPrefix(mod, coreModule) {
		return true
	}
	return false
}

func isOtelContribComponent(mod string) bool {
	// Check if the component is part of the OpenTelemetry Collector contrib
	if strings.HasPrefix(mod, contribModule) {
		return true
	}
	return false
}

func isNrdotComponent(component Module) bool {
	// Check if the component is part of the NRDOT Collector
	if strings.HasPrefix(component.GoMod, nrModule) {
		return true
	}
	return false
}

func isOtelComponent(component Module) bool {
	// Check if the component is part of the OpenTelemetry Collector
	if isOtelCoreComponent(component.GoMod) || isOtelContribComponent(component.GoMod) {
		return true
	}
	return false
}

func isStableVersion(version string) bool {
	// Check if the version is a stable version (not a pre-release)
	if semver.Compare(version, "v1.0.0") >= 0 {
		return true
	}
	return false
}

func isCompatibleWithNrdotComponent(nrdotVersion string, betaVersion string) bool {
	if semver.Compare(nrdotVersion, betaVersion) >= 0 {
		return true
	}
	return false
}

func (c *Config) SetVersions() error {

	versions := Versions{}

	for _, component := range c.allNrdotComponents() {
		if isNrdotComponent(component) {
			componentVersion := strings.Split(component.GoMod, " ")[1]
			versions.NrdotVersion = componentVersion
		}

		if versions.NrdotVersion != "" {
			break
		}
	}

	for _, component := range c.allOtelComponents() {
		if isOtelComponent(component) {
			componentVersion := strings.Split(component.GoMod, " ")[1]
			if isOtelCoreComponent(component.GoMod) {
				if isStableVersion(componentVersion) {
					versions.StableCoreVersion = componentVersion
				} else if versions.NrdotVersion == "" || isCompatibleWithNrdotComponent(versions.NrdotVersion, componentVersion) {
					versions.BetaCoreVersion = componentVersion
				}
			}

			if isOtelContribComponent(component.GoMod) && !isStableVersion(componentVersion) && (versions.NrdotVersion == "" || isCompatibleWithNrdotComponent(versions.NrdotVersion, componentVersion)) {
				versions.BetaContribVersion = componentVersion
			}

			if versions.StableCoreVersion != "" && versions.BetaCoreVersion != "" && versions.BetaContribVersion != "" {
				break
			}
		}
	}

	if versions.BetaCoreVersion == "" {
		return fmt.Errorf("missing beta core version")
	}

	c.Versions = versions
	return nil
}

// Validate checks whether the current configuration is valid
func (c *Config) Validate() error {
	return multierr.Combine(
		validateModules("extension", c.Extensions),
		validateModules("receiver", c.Receivers),
		validateModules("exporter", c.Exporters),
		validateModules("processor", c.Processors),
		validateModules("connector", c.Connectors),
		validateModules("provider", c.ConfmapProviders),
		validateModules("converter", c.ConfmapConverters),
	)
}

// SetGoPath sets go path
func (c *Config) SetGoPath() error {
	//nolint:gosec // #nosec G204
	if _, err := exec.Command(c.Distribution.Go, "env").CombinedOutput(); err != nil {
		path, err := exec.LookPath("go")
		if err != nil {
			return ErrGoNotFound
		}
		c.Distribution.Go = path
	}
	if c.Verbose {
		c.Logger.Info("Using go", zap.String("go-executable", c.Distribution.Go))
	}
	return nil
}

// ParseModules will parse the Modules entries and populate the missing values
func (c *Config) ParseModules() error {
	var err error
	usedNames := make(map[string]int)

	c.Extensions, err = parseModules(c.Extensions, usedNames)
	if err != nil {
		return err
	}

	c.Receivers, err = parseModules(c.Receivers, usedNames)
	if err != nil {
		return err
	}

	c.Exporters, err = parseModules(c.Exporters, usedNames)
	if err != nil {
		return err
	}

	c.Processors, err = parseModules(c.Processors, usedNames)
	if err != nil {
		return err
	}

	c.Connectors, err = parseModules(c.Connectors, usedNames)
	if err != nil {
		return err
	}

	c.ConfmapProviders, err = parseModules(c.ConfmapProviders, usedNames)
	if err != nil {
		return err
	}
	c.ConfmapConverters, err = parseModules(c.ConfmapConverters, usedNames)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) allComponents() []Module {
	return slices.Concat(c.Exporters, c.Receivers, c.Processors, c.Extensions, c.Connectors, c.ConfmapProviders, c.ConfmapConverters)
}

func (cfg *Config) allOtelComponents() []Module {
	allOtelComponents := []Module{}
	for _, component := range cfg.allComponents() {
		if isOtelComponent(component) {
			allOtelComponents = append(allOtelComponents, component)
		}
	}
	return allOtelComponents
}

func (cfg *Config) allNrdotComponents() []Module {
	allNrdotComponents := []Module{}
	for _, component := range cfg.allComponents() {
		if isNrdotComponent(component) {
			allNrdotComponents = append(allNrdotComponents, component)
		}
	}
	return allNrdotComponents
}

func validateModules(name string, mods []Module) error {
	for i, mod := range mods {
		if mod.GoMod == "" {
			return fmt.Errorf("%s module at index %v: %w", name, i, errMissingGoMod)
		}
	}
	return nil
}

func parseModules(mods []Module, usedNames map[string]int) ([]Module, error) {
	var parsedModules []Module
	for _, mod := range mods {
		if mod.Import == "" {
			mod.Import = strings.Split(mod.GoMod, " ")[0]
		}

		if mod.Name == "" {
			parts := strings.Split(mod.Import, "/")
			mod.Name = parts[len(parts)-1]
		}

		originalModName := mod.Name
		if count, exists := usedNames[mod.Name]; exists {
			var newName string
			for {
				newName = fmt.Sprintf("%s%d", mod.Name, count+1)
				if _, transformedExists := usedNames[newName]; !transformedExists {
					break
				}
				count++
			}
			mod.Name = newName
			usedNames[newName] = 1
		}
		usedNames[originalModName] = 1

		// Check if path is empty, otherwise filepath.Abs replaces it with current path ".".
		if mod.Path != "" {
			var err error
			mod.Path, err = filepath.Abs(mod.Path)
			if err != nil {
				return mods, fmt.Errorf("module has a relative \"path\" element, but we couldn't resolve the current working dir: %w", err)
			}
			// Check if the path exists
			if _, err := os.Stat(mod.Path); os.IsNotExist(err) {
				return mods, fmt.Errorf("filepath does not exist: %s", mod.Path)
			}
		}

		parsedModules = append(parsedModules, mod)
	}

	return parsedModules, nil
}
