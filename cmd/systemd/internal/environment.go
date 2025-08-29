package internal

func GetEnvironmentTemplate() string {
	return `# Systemd environment file for the {{.FullDist}} service
# Command-line options for the {{.FullDist}} service.
# See https://opentelemetry.io/docs/collector/configuration/ to see all available options.
OTELCOL_OPTIONS="--config=/etc/{{.Dist}}/config.yaml"
`
}

func GetEnvironmentData(dist string, fullDist string) EnvironmentData {
	return EnvironmentData{
		Dist:     dist,
		FullDist: fullDist,
	}
}

type EnvironmentData struct {
	Dist     string
	FullDist string
}
