// Copyright 2025 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

var (
	// ErrGoNotFound is returned when a Go binary hasn't been found
	ErrGoNotFound = errors.New("go binary not found")

	ErrVersionMismatch = errors.New("mismatch in go.mod and builder configuration versions")
)

// runGoCommand replicates behavoir of the OCB, effectively running `go list` to fetch module versions
func runGoCommand(cfg *Config, args ...string) ([]byte, error) {
	if cfg.Verbose {
		cfg.Logger.Info("Running go subcommand.", zap.Any("arguments", args))
	}

	//nolint:gosec // #nosec G204 -- cfg.Distribution.Go is trusted to be a safe path and the caller is assumed to have carried out necessary input validation
	cmd := exec.Command(cfg.Distribution.Go, args...)
	cmd.Dir = cfg.Dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go subcommand failed with args '%v': %w, error message: %s", args, err, stderr.String())
	}
	if cfg.Verbose && stderr.Len() != 0 {
		cfg.Logger.Info("go subcommand error", zap.String("message", stderr.String()))
	}

	return stdout.Bytes(), nil
}

func fetchAllModuleVersions(cfg *Config, modules []string) (map[string][]string, error) {
	// Run `go list -m -versions <module1> <module2>` to fetch the versions for this module
	output, err := runGoCommand(cfg, append([]string{"list", "-versions", "-m"}, modules...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch module versions: %w", err)
	}

	moduleVersions := make(map[string][]string)
	// Split the input into lines and iterate over each line
	for line := range strings.SplitSeq(string(output), "\n") {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Split the line into module and versions
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // Skip lines that don't have at least a module and one version
		}
		// Add to the map
		moduleVersions[parts[0]] = parts[1:]
	}

	return moduleVersions, nil
}

func fetchLatestModuleVersions(cfg *Config) (map[string][]string, error) {
	updates := make(map[string][]string)

	var modules []string
	for _, component := range cfg.allOtelComponents() {
		module, _, _ := strings.Cut(component.GoMod, " ")
		modules = append(modules, module)
	}

	versions, err := fetchAllModuleVersions(cfg, modules)
	if err != nil {
		cfg.Logger.Warn("Failed to fetch module updates", zap.String("module", "all"), zap.Error(err))
		return nil, err
	}

	// Iterate over all components in cfg.allComponents()
	for _, component := range cfg.allOtelComponents() {
		// Extract the module name from the component's GoMod field
		module, currentVersion, _ := strings.Cut(component.GoMod, " ")
		// Log the current version
		if cfg.Verbose {
			cfg.Logger.Info("Checking for updates", zap.String("module", module), zap.String("currentVersion", currentVersion))
		}
		// Iterate through the versions and filter out any that are equal to or less than the current version using semver.Compare
		for _, version := range versions[module] {
			if semver.Compare(version, currentVersion) > 0 {
				updates[module] = append(updates[module], version)
			}
		}

		if len(updates[module]) > 0 {
			// This is a sanity check, as the go list command should already return sorted versions
			sort.Slice(updates[module], func(i, j int) bool {
				return semver.Compare(updates[module][i], updates[module][j]) < 0
			})
			if cfg.Verbose {
				cfg.Logger.Debug("Valid update candidates", zap.String("module", module), zap.Strings("versions", updates[module]))
			}
		} else {
			if cfg.Verbose {
				cfg.Logger.Debug("No updates available for module",
					zap.String("module", module),
					zap.String("currentVersion", currentVersion),
				)
			}
		}
	}

	return updates, nil
}

func CopyAndUpdateConfigModules(cfg *Config, updates map[string][]string) (*Config, error) {
	// Create a deep copy of the cfg struct
	cfgCopy := *cfg
	update := func(components []Module) []Module {
		updatedComponents := make([]Module, len(components))
		for i, component := range components {
			module, _, _ := strings.Cut(component.GoMod, " ")
			latestVersions, exists := updates[module]
			if exists {
				// Update the GoMod field with the latest version
				component.GoMod = fmt.Sprintf("%s %s", module, latestVersions[len(latestVersions)-1])
			} else {
				if cfg.Verbose {
					// log the module name and version
					cfg.Logger.Info("No updates available for module", zap.String("module", module))
				}
			}

			updatedComponents[i] = component
		}

		return updatedComponents
	}

	cfgCopy.Exporters = update(cfg.Exporters)
	cfgCopy.Receivers = update(cfg.Receivers)
	cfgCopy.Processors = update(cfg.Processors)
	cfgCopy.Extensions = update(cfg.Extensions)
	cfgCopy.Connectors = update(cfg.Connectors)
	cfgCopy.ConfmapProviders = update(cfg.ConfmapProviders)
	cfgCopy.ConfmapConverters = update(cfg.ConfmapConverters)

	cfgCopy.SetVersions()

	return &cfgCopy, nil
}

// GetModules retrieves the go modules, updating go.mod and go.sum in the process
func UpdateConfigModules(cfg *Config) (*Config, error) {
	updates, err := fetchLatestModuleVersions(cfg)
	if err != nil {
		cfg.Logger.Error("Failed to fetch latest module versions", zap.Error(err))
		return nil, err
	}

	// Iterate over all components and check for updates
	// Create a copy of cfg with updated modules
	updatedCfg, err := CopyAndUpdateConfigModules(cfg, updates)
	if err != nil {
		cfg.Logger.Error("Failed to update config modules", zap.Error(err))
		return nil, err
	}

	// Log the updated modules
	for _, component := range updatedCfg.allComponents() {
		if cfg.Verbose {
			cfg.Logger.Info("Updated module",
				zap.String("module", component.GoMod),
			)
		}
	}

	return updatedCfg, nil
}

func WriteConfigFile(cfg *Config) error {
	if cfg.Verbose {
		// Log the start of the operation
		cfg.Logger.Info("Writing updated configuration to file", zap.String("path", cfg.Path))
	}
	// Open the YAML file
	file, err := os.Open(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer file.Close()

	// Decode the YAML into a yaml.Node
	var root yaml.Node
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&root); err != nil {
		return fmt.Errorf("failed to decode YAML file: %w", err)
	}

	// Map the components to their respective YAML keys
	var componentMap = map[string][]Module{
		"exporters":  cfg.Exporters,
		"extensions": cfg.Extensions,
		"receivers":  cfg.Receivers,
		"processors": cfg.Processors,
		"connectors": cfg.Connectors,
		"providers":  cfg.ConfmapProviders,
		"converters": cfg.ConfmapConverters,
	}

	// Recursively walk through the YAML nodes and update them
	updateYamlNodes(&root, componentMap)

	// Write the updated YAML back to the file
	outputFile, err := os.Create(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %w", err)
	}
	defer outputFile.Close()

	encoder := yaml.NewEncoder(outputFile)
	encoder.SetIndent(2) // Optional: Set indentation for readability
	if err := encoder.Encode(&root); err != nil {
		return fmt.Errorf("failed to encode YAML file: %w", err)
	}

	if cfg.Verbose {
		cfg.Logger.Info("Successfully wrote updated configuration to file", zap.String("path", cfg.Path))
	}
	return nil
}

// Recursive function to walk through YAML nodes and update them, we do this rather than simply write out the config in order to preserve comments and formatting
func updateYamlNodes(node *yaml.Node, componentMap map[string][]Module) {
	// If the node is a MappingNode, process its key-value pairs
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			// Check if the key matches a component type (e.g., "processors")
			if components, ok := componentMap[key.Value]; ok {
				// If the value is a SequenceNode, iterate over its items
				if value.Kind == yaml.SequenceNode {
					for _, item := range value.Content {
						if item.Kind == yaml.MappingNode {
							// Update the `gomod` key in the mapping node
							updateGomodKey(item, components)
						}
					}
				}

				sortComponentYamlNode(value)
			} else {
				// Recursively process the value node
				updateYamlNodes(value, componentMap)
			}
		}
	}

	// If the node is a SequenceNode, process its items
	if node.Kind == yaml.SequenceNode || node.Kind == yaml.DocumentNode {
		for _, item := range node.Content {
			updateYamlNodes(item, componentMap)
		}
	}
}

// Update the `gomod` key in a MappingNode
func updateGomodKey(node *yaml.Node, components []Module) {

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]

		if key.Value == "gomod" {
			// Update the `gomod` value with the corresponding value from the components
			for _, component := range components {
				if strings.HasPrefix(value.Value, component.GoMod[:strings.Index(component.GoMod, " ")]) {
					value.Value = component.GoMod
					break
				}
			}
		}
	}
}

func getNodeGoModValue(node *yaml.Node) Module {
	if node.Kind == yaml.ScalarNode {
		return Module{}
	}

	mod := Module{}

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			switch key.Value {
			case "name":
				mod.Name = value.Value
			case "import":
				mod.Import = value.Value
			case "gomod":
				mod.GoMod = value.Value
			case "path":
				mod.Path = value.Value
			}
		}

	}

	if mod.Import == "" {
		mod.Import = strings.Split(mod.GoMod, " ")[0]
	}

	if mod.Name == "" {
		parts := strings.Split(mod.Import, "/")
		mod.Name = parts[len(parts)-1]
	}

	return mod
}

func sortComponentYamlNode(node *yaml.Node) {
	if node.Kind != yaml.SequenceNode {
		return
	}

	// Sort value.Content by otel core components first, then otel contrib components, then others
	sort.Slice(node.Content, func(i, j int) bool {
		nodeI := node.Content[i]
		nodeJ := node.Content[j]

		gomodI := getNodeGoModValue(nodeI)
		gomodJ := getNodeGoModValue(nodeJ)
		// Check if both components are part of the OpenTelemetry Collector
		isOtelCoreI := isOtelCoreComponent(gomodI.GoMod)
		isOtelCoreJ := isOtelCoreComponent(gomodJ.GoMod)
		if isOtelCoreI && !isOtelCoreJ {
			return true
		} else if !isOtelCoreI && isOtelCoreJ {
			return false
		}
		isOtelContribI := isOtelContribComponent(gomodI.GoMod)
		isOtelContribJ := isOtelContribComponent(gomodJ.GoMod)

		if isOtelContribI && !isOtelContribJ {
			return true
		} else if !isOtelContribI && isOtelContribJ {
			return false
		}

		return gomodI.Name < gomodJ.Name
	})
}
