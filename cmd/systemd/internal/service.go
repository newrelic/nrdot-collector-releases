package internal

import (
	"fmt"
)

func GetServiceTemplate() string {
	return `[Unit]
Description={{.Unit.Description}}
After={{.Unit.After}}

[Service]
EnvironmentFile={{.Service.EnvironmentFile}}
ExecStart={{.Service.ExecStart}}
KillMode={{.Service.KillMode}}
Restart={{.Service.Restart}}
Type={{.Service.Type}}
User={{.Service.User}}
Group={{.Service.Group}}

[Install]
WantedBy={{.Install.WantedBy}}
`
}

func GenerateServiceData(dist string, fullDist string, description string) ServiceData {
	data := ServiceData{
		Unit: Unit{
			Description: description,
			After:       "network.target",
		},
		Service: Service{
			EnvironmentFile: fmt.Sprintf("/etc/%s/%s.conf", dist, fullDist),
			ExecStart:       fmt.Sprintf("/usr/bin/%s $OTELCOL_OPTIONS", fullDist),
			KillMode:        "mixed",
			Restart:         "on-failure",
			Type:            "simple",
			User:            dist,
			Group:           dist,
		},
		Install: Install{
			WantedBy: "multi-user.target",
		},
	}

	return data
}

type ServiceData struct {
	Unit    Unit
	Service Service
	Install Install
}

type Unit struct {
	Description string
	After       string
}

type Service struct {
	EnvironmentFile string
	ExecStart       string
	KillMode        string
	Restart         string
	Type            string
	User            string
	Group           string
}

type Install struct {
	WantedBy string
}
