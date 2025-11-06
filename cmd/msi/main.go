// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"
)

type Guids struct {
	ProductUpgradeCode string `json:"product_upgrade_code"`
	ComponentGuid      string `json:"component_guid"`
}

type DistGuids struct {
	Standard Guids `json:"standard"`
	Fips     Guids `json:"fips"`
}

const (
	HostDistro       = "nrdot-collector-host"
	K8sDistro        = "nrdot-collector-k8s"
	CoreDistro       = "nrdot-collector"
	templateFilename = "cmd/msi/windows-installer.wxs.tmpl"
)

var (
	distFlag = flag.String("d", "", "Collector distributions to build")
	fipsFlag = flag.Bool("f", false, "FIPS")
)

func main() {
	flag.Parse()

	if len(*distFlag) == 0 {
		log.Fatal("no distribution to template")
	}

	// Get GUIDs

	// Parse the base template
	baseTemplate, err := template.New("base").Delims("<<", ">>").ParseFiles(templateFilename)
	if err != nil {
		panic(err)
	}

	guids := getMsiGuids(*distFlag, *fipsFlag)

	// Data for the base template
	data := map[string]interface{}{
		"InstallerName":      getInstallerName(*distFlag, *fipsFlag),
		"ProductUpgradeCode": guids.ProductUpgradeCode,
		"ComponentGUID":      guids.ComponentGuid,
	}

	// Execute the base template to generate a new template
	var generatedTemplateContent bytes.Buffer
	err = baseTemplate.ExecuteTemplate(&generatedTemplateContent, "base", data)
	if err != nil {
		panic(err)
	}

	err = baseTemplate.ExecuteTemplate(os.Stdout, "base", data)
	if err != nil {
		panic(err)
	}
}

func getInstallerName(dist string, fips bool) string {
	name := ""
	switch dist {
	case CoreDistro:
		name = "Core"
	case HostDistro:
		name = "Host"
	case K8sDistro:
		name = "K8s"
	default:
		log.Fatal("Unknown Distribution:", dist)
	}
	if fips {
		name += " (FIPS)"
	}
	return name
}

func getMsiGuids(dist string, fips bool) Guids {
	filename := fmt.Sprintf("./distributions/%s/msi-guids.json", *distFlag)
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var distGuids DistGuids
	err = json.Unmarshal(data, &distGuids)
	if err != nil {
		panic(err)
	}

	if fips {
		return distGuids.Fips
	} else {
		return distGuids.Standard
	}
}
