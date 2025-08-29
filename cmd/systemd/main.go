package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/newrelic/nrdot-collector-releases/cmd/systemd/internal"
)

const (
	K8sDistro  = "nrdot-collector-k8s"
	HostDistro = "nrdot-collector-host"
	CoreDistro = "nrdot-collector"

	ServiceOutput = "service"
	EnvOutput     = "env"
)

var descriptions = map[string]string{
	HostDistro: "NRDOT Collector Host",
	K8sDistro:  "NRDOT Collector K8s",
	CoreDistro: "NRDOT Collector",
}

var distFlag = flag.String("d", "", "Collector distributions to build")
var outputFlag = flag.String("o", "", "Which systemd file to output - service or env")
var fipsFlag = flag.Bool("f", false, "Whether we're building a FIPS compliant config")

func main() {
	flag.Parse()

	if len(*distFlag) == 0 {
		log.Fatal("no distribution to build")
	}

	dist := *distFlag
	fullDist := *distFlag
	desc := descriptions[*distFlag]
	if *fipsFlag {
		fullDist = fmt.Sprint(dist, "-fips")
		desc = fmt.Sprint(desc, " FIPS")
	}

	switch *outputFlag {

	case ServiceOutput:
		serviceTemplate := internal.GetServiceTemplate()
		t := template.Must(template.New(".service").Parse(serviceTemplate))

		data := internal.GenerateServiceData(dist, fullDist, desc)

		t.Execute(os.Stdout, data)

	case EnvOutput:
		envTemplate := internal.GetEnvironmentTemplate()

		t := template.Must(template.New(".conf").Parse(envTemplate))

		data := internal.GetEnvironmentData(dist, fullDist)

		t.Execute(os.Stdout, data)

	default:
		log.Fatal("expected either \"service\" or \"env\" for output flag")
	}

}
