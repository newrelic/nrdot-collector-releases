package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: builder <distribution-name>")
	}

	distributionName := os.Args[1]
	log.Printf("Building distribution: %s", distributionName)

	cmd := exec.Command("ocb", "--config", "manifest.yaml", "--skip-compilation=false")
	cmd.Dir = filepath.Join("distributions", distributionName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Running OpenTelemetry Collector Builder...")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run builder: %v", err)
	}

	componentsPath := filepath.Join("distributions", distributionName, "_build", "components.go")
	componentsData, err := os.ReadFile(componentsPath)
	if err != nil {
		log.Fatalf("Failed to read components.go: %v", err)
	}

	componentsContent := string(componentsData)

	if strings.Contains(componentsContent, "topnprocessfilterprocessor") {
		log.Println("topnprocessfilter processor is already registered")
		return
	}

	log.Println("Adding topnprocessfilter processor to components.go")

	newContent := strings.Replace(
		componentsContent,
		"transformprocessor \"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor\"",
		"transformprocessor \"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor\"\n\ttopnprocessfilterprocessor \"github.com/newrelic/nrdot-collector-releases/internal/processor/topnprocessfilter\"",
		1,
	)

	newContent = strings.Replace(
		newContent,
		"transformprocessor.NewFactory(),",
		"transformprocessor.NewFactory(),\n\t\ttopnprocessfilterprocessor.NewFactory(),",
		1,
	)

	newContent = strings.Replace(
		newContent,
		"factories.ProcessorModules[transformprocessor.NewFactory().Type()] = \"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.125.0\"",
		"factories.ProcessorModules[transformprocessor.NewFactory().Type()] = \"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.125.0\"\n\tfactories.ProcessorModules[topnprocessfilterprocessor.NewFactory().Type()] = \"github.com/newrelic/nrdot-collector-releases/internal/processor/topnprocessfilter v0.0.0\"",
		1,
	)

	if err := os.WriteFile(componentsPath, []byte(newContent), 0644); err != nil {
		log.Fatalf("Failed to write updated components.go: %v", err)
	}

	buildCmd := exec.Command("go", "build", "-o", distributionName)
	buildCmd.Dir = filepath.Join("distributions", distributionName, "_build")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	log.Println("Building the collector with the topnprocessfilter processor...")
	if err := buildCmd.Run(); err != nil {
		log.Fatalf("Failed to build the collector: %v", err)
	}

	log.Printf("Successfully built %s with topnprocessfilter processor", distributionName)
}
