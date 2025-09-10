// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

// This file is a script which generates the .goreleaser.yaml file for all
// supported NRDOT Collector distributions.
//
// Run it with `make generate-goreleaser`.

import (
	"fmt"
	"path"

	"github.com/goreleaser/goreleaser/v2/pkg/config"
)

const (
	HostDistro = "nrdot-collector-host"
	K8sDistro  = "nrdot-collector-k8s"
	CoreDistro = "nrdot-collector"
)

type Distribution struct {
	BaseName string
	FullName string // dist or dist-fips
	Goos     []string
}

var (
	Architectures = []string{"amd64", "arm64"}
	SkipBinaries  = map[string]bool{
		K8sDistro: true,
	}
	IncludedConfig = map[string]string{
		HostDistro: "config.yaml",
		CoreDistro: "config.yaml",
		// k8s missing due to not packaged via nfpm
	}
	FipsLdflags = []string{"-w", "-linkmode external", "-extldflags '-static'"}
	FipsGoTags  = []string{"netgo"}
)

func GetDistribution(distFlag string, fips bool) Distribution {
	fullName := distFlag
	if fips {
		fullName += "-fips"
	}

	dist := Distribution{
		BaseName: distFlag,
		FullName: fullName,
		Goos:     []string{"linux", "windows"},
	}

	if distFlag == K8sDistro || fips {
		dist.Goos = []string{"linux"}
	}

	return dist
}

func Generate(distFlag string, nightly bool, fips bool) config.Project {
	projectName := "nrdot-collector-releases"
	disableRelease := "false"

	if nightly {
		projectName = "nrdot-collector-releases-nightly"
		disableRelease = "true"
	}

	fullName := distFlag
	if fips {
		fullName += "-fips"
	}
	dist := GetDistribution(distFlag, fips)

	return config.Project{
		ProjectName: projectName,
		Checksum: config.Checksum{
			NameTemplate: "{{ .ArtifactName }}.sum",
			Split:        true,
			Algorithm:    "sha256",
		},
		Builds:          Builds(dist, fips),
		Archives:        Archives(dist, fips),
		NFPMs:           Packages(dist, fips),
		Dockers:         DockerImages(dist, nightly, fips),
		DockerManifests: DockerManifests(dist, nightly, fips),
		Signs:           Sign(),
		Version:         2,
		Changelog:       config.Changelog{Disable: "true"},
		Snapshot: config.Snapshot{
			VersionTemplate: "{{ incpatch .Version }}-SNAPSHOT-{{.ShortCommit}}",
		},
		Blobs: Blobs(dist, nightly, fips),
		Release: config.Release{
			Disable:              disableRelease,
			Draft:                true,
			UseExistingDraft:     true,
			ReplaceExistingDraft: false,
		},
	}
}

func Blobs(dist Distribution, nightly bool, fips bool) []config.Blob {
	if skip, ok := SkipBinaries[dist.BaseName]; ok && skip {
		return nil
	}

	return []config.Blob{
		Blob(dist, nightly, fips),
	}
}

func Blob(dist Distribution, nightly bool, fips bool) config.Blob {
	version := "{{ .Version }}"

	if nightly {
		version = "nightly"
	}

	return config.Blob{
		Provider:  "s3",
		Region:    "us-east-1",
		Bucket:    "nr-releases",
		Directory: fmt.Sprintf("nrdot-collector-releases/%s/%s", dist.FullName, version),
	}
}

func Builds(dist Distribution, fips bool) []config.Build {
	return []config.Build{
		Build(dist, fips),
	}
}

// Build configures a goreleaser build.
// https://goreleaser.com/customization/build/
func Build(dist Distribution, fips bool) config.Build {
	dir := "_build"
	cgo := 0
	ignoreBuild := IgnoreBuildCombinations(dist, fips)
	ldflags := []string{"-s", "-w"}
	gotags := []string{}
	goexperiment := ""

	var buildDetailsOverrides []config.BuildDetailsOverride

	cc := map[string]string{
		"amd64": "x86_64-linux-gnu-gcc",
		"arm64": "aarch64-linux-gnu-gcc",
	}

	cxx := map[string]string{
		"amd64": "x86_64-linux-gnu-g++",
		"arm64": "aarch64-linux-gnu-g++",
	}

	if fips {
		dir = "_build-fips"
		cgo = 1
		goexperiment = "boringcrypto"
		ldflags = FipsLdflags
		gotags = FipsGoTags
		for _, arch := range Architectures {
			buildDetailsOverrides = append(buildDetailsOverrides, config.BuildDetailsOverride{
				Goos:   dist.Goos[0],
				Goarch: arch,
				BuildDetails: config.BuildDetails{
					Env: []string{
						fmt.Sprint("CC=", cc[arch]),
						fmt.Sprint("CXX=", cxx[arch]),
					},
				},
			})
		}
	}

	return config.Build{
		ID:     dist.FullName,
		Dir:    dir,
		Binary: dist.FullName,
		BuildDetails: config.BuildDetails{
			Env:     []string{fmt.Sprint("CGO_ENABLED=", cgo), fmt.Sprint("GOEXPERIMENT=", goexperiment)},
			Flags:   []string{"-trimpath"},
			Ldflags: ldflags,
			Tags:    gotags,
		},
		BuildDetailsOverrides: buildDetailsOverrides,
		Goos:                  dist.Goos,
		Goarch:                Architectures,
		Ignore:                ignoreBuild,
	}
}

func IgnoreBuildCombinations(dist Distribution, fips bool) []config.IgnoredBuild {
	if dist.BaseName == K8sDistro || fips {
		return nil
	}
	return []config.IgnoredBuild{
		{Goos: "windows", Goarch: "arm64"},
	}
}

func Archives(dist Distribution, fips bool) []config.Archive {
	return []config.Archive{
		Archive(dist, fips),
	}
}

// Archive configures a goreleaser archive (tarball).
// https://goreleaser.com/customization/archive/
func Archive(dist Distribution, fips bool) config.Archive {
	files := make([]config.File, 0)
	goos := "windows"

	if configFile, ok := IncludedConfig[dist.BaseName]; ok {
		files = append(files, config.File{
			Source: configFile,
		})
	}

	if fips {
		goos = "linux"
	}

	return config.Archive{
		ID:           dist.FullName,
		NameTemplate: "{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}{{ if .Mips }}_{{ .Mips }}{{ end }}",
		IDs:          []string{dist.FullName},
		Files:        files,
		FormatOverrides: []config.FormatOverride{
			{
				Goos: goos, Formats: []string{"zip"},
			},
		},
	}
}

func Packages(dist Distribution, fips bool) []config.NFPM {
	if skip, ok := SkipBinaries[dist.BaseName]; ok && skip {
		return nil
	}

	return []config.NFPM{
		Package(dist, fips),
	}
}

// Package configures goreleaser to build a system package.
// https://goreleaser.com/customization/nfpm/
func Package(dist Distribution, fips bool) config.NFPM {
	configFile, ok := IncludedConfig[dist.BaseName]

	nfpmContents := []config.NFPMContent{
		{
			Source:      fmt.Sprintf("%s.service", dist.FullName),
			Destination: path.Join("/lib", "systemd", "system", fmt.Sprintf("%s.service", dist.FullName)),
		},
		{
			Source:      fmt.Sprintf("%s.conf", dist.FullName),
			Destination: path.Join("/etc", dist.FullName, fmt.Sprintf("%s.conf", dist.FullName)),
			Type:        "config|noreplace",
		},
	}

	if ok {
		nfpmContents = append(nfpmContents, config.NFPMContent{
			Source:      configFile,
			Destination: path.Join("/etc", dist.FullName, configFile),
			Type:        "config",
		})
	}
	return config.NFPM{
		ID:          dist.FullName,
		IDs:         []string{dist.FullName},
		Formats:     []string{"deb", "rpm"},
		License:     "Apache 2.0",
		Description: fmt.Sprintf("NRDOT Collector - %s", dist.FullName),
		Maintainer:  "New Relic <otelcomm-team@newrelic.com>",
		Overrides: map[string]config.NFPMOverridables{
			"rpm": {
				Dependencies: []string{"/bin/sh"},
			},
		},
		NFPMOverridables: config.NFPMOverridables{
			PackageName: dist.FullName,
			FileNameTemplate: "{{ .PackageName }}_{{ .Version }}_{{ .Os }}_" +
				"{{- if not (eq (filter .ConventionalFileName \"\\\\.rpm$\") \"\") }}" +
				"{{- replace .Arch \"amd64\" \"x86_64\" }}" +
				"{{- else }}" +
				"{{- .Arch }}" +
				"{{- end }}" +
				"{{- with .Arm }}v{{ . }}{{- end }}" +
				"{{- with .Mips }}_{{ . }}{{- end }}" +
				"{{- if not (eq .Amd64 \"v1\") }}{{ .Amd64 }}{{- end }}",
			Scripts: config.NFPMScripts{
				PreInstall:  "preinstall.sh",
				PostInstall: "postinstall.sh",
				PreRemove:   "preremove.sh",
			},
			Contents: nfpmContents,
			RPM: config.NFPMRPM{
				Signature: config.NFPMRPMSignature{
					KeyFile: "{{ .Env.GPG_KEY_PATH }}",
				},
			},
			Deb: config.NFPMDeb{
				Signature: config.NFPMDebSignature{
					KeyFile: "{{ .Env.GPG_KEY_PATH }}",
				},
			},
		},
	}
}

func DockerImageTags(nightly bool, fips bool) []string {
	tags := []string{}
	if fips {
		tags = append(tags, "{{ .Version }}-fips")
	} else if nightly {
		tags = append(tags, "{{ .Version }}-nightly")
		tags = append(tags, "nightly")
	} else {
		tags = append(tags, "{{ .Version }}")
		tags = append(tags, "latest")
	}
	return tags
}

func DockerImages(dist Distribution, nightly bool, fips bool) []config.Docker {
	var r []config.Docker

	for _, arch := range Architectures {
		r = append(r, DockerImage(dist, nightly, arch, fips))
	}

	return r
}

// DockerImage configures goreleaser to build a container image.
// https://goreleaser.com/customization/docker/
func DockerImage(dist Distribution, nightly bool, arch string, fips bool) config.Docker {

	dockerFile := "Dockerfile"

	imageTemplates := make([]string, 0)
	for _, tag := range DockerImageTags(nightly, fips) {
		imageTemplates = append(
			imageTemplates,
			fmt.Sprintf("{{ .Env.REGISTRY }}/%s:%s-%s", dist.BaseName, tag, arch),
		)
	}

	label := func(name, template string) string {
		return fmt.Sprintf("--label=org.opencontainers.image.%s={{%s}}", name, template)
	}

	files := make([]string, 0)
	if configFile, ok := IncludedConfig[dist.BaseName]; ok {
		files = append(files, configFile)
	}

	return config.Docker{
		ImageTemplates: imageTemplates,
		Dockerfile:     dockerFile,

		Use: "buildx",
		BuildFlagTemplates: []string{
			"--pull",
			fmt.Sprintf("--platform=linux/%s", arch),
			label("created", ".Date"),
			label("name", ".ProjectName"),
			label("revision", ".FullCommit"),
			label("version", ".Version"),
			label("source", ".GitURL"),
			"--label=org.opencontainers.image.licenses=Apache-2.0",
			fmt.Sprint("--build-arg=DIST_NAME=", dist.FullName),
		},
		Files:  files,
		Goos:   "linux",
		Goarch: arch,
	}
}

func DockerManifests(dist Distribution, nightly bool, fips bool) []config.DockerManifest {
	r := make([]config.DockerManifest, 0)

	for _, tag := range DockerImageTags(nightly, fips) {
		r = append(r, DockerManifest(tag, dist))
	}

	return r
}

// DockerManifest configures goreleaser to build a multi-arch container image manifest.
// https://goreleaser.com/customization/docker_manifest/
func DockerManifest(version string, dist Distribution) config.DockerManifest {
	var imageTemplates []string

	for _, arch := range Architectures {
		imageTemplates = append(
			imageTemplates,
			fmt.Sprintf("{{ .Env.REGISTRY }}/%s:%s-%s", dist.BaseName, version, arch),
		)
	}

	return config.DockerManifest{
		NameTemplate:   fmt.Sprintf("{{ .Env.REGISTRY }}/%s:%s", dist.BaseName, version),
		ImageTemplates: imageTemplates,
	}
}

func Sign() []config.Sign {
	return []config.Sign{
		{
			Artifacts: "all",
			Signature: "${artifact}.asc",
			Args: []string{
				"--batch",
				"-u",
				"{{ .Env.GPG_FINGERPRINT }}",
				"--output",
				"${signature}",
				"--detach-sign",
				"--armor",
				"${artifact}",
			},
		},
	}
}
