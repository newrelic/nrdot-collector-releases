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

	ConfigFile = "config.yaml"
)

type Distribution struct {
	BaseName      string
	FullName      string // dist or dist-fips
	Nightly       bool
	Fips          bool
	Goos          []string
	IgnoredBuilds []config.IgnoredBuild
	IncludeConfig bool
	SkipBinaries  bool
	SkipArchives  bool
}

var (
	Architectures = []string{"amd64", "arm64"}
	FipsLdflags   = []string{"-w", "-linkmode external", "-extldflags '-static'"}
	FipsGoTags    = []string{"netgo"}
)

func Generate(distFlag string, nightly bool, fips bool) config.Project {
	projectName := "nrdot-collector-releases"
	disableRelease := "false"

	if nightly {
		projectName = "nrdot-collector-releases-nightly"
		disableRelease = "true"
	}

	dist := NewDistribution(distFlag, nightly, fips)

	return config.Project{
		ProjectName: projectName,
		Checksum: config.Checksum{
			NameTemplate: "{{ .ArtifactName }}.sum",
			Split:        true,
			Algorithm:    "sha256",
		},
		Builds:          Builds(dist),
		Archives:        Archives(dist),
		NFPMs:           Packages(dist),
		Dockers:         DockerImages(dist),
		DockerManifests: DockerManifests(dist),
		Signs:           Sign(),
		Version:         2,
		Changelog:       config.Changelog{Disable: "true"},
		Snapshot: config.Snapshot{
			VersionTemplate: "{{ incpatch .Version }}-SNAPSHOT-{{.ShortCommit}}",
		},
		Blobs: Blobs(dist),
		Release: config.Release{
			Disable:              disableRelease,
			Draft:                true,
			UseExistingDraft:     true,
			ReplaceExistingDraft: false,
		},
	}
}

func NewDistribution(baseDist string, nightly bool, fips bool) Distribution {
	fullName := baseDist
	if fips {
		fullName += "-fips"
	}

	dist := Distribution{
		BaseName: baseDist,
		FullName: fullName,
		Nightly:  nightly,
		Fips:     fips,
		Goos:     []string{"linux", "windows"},
		IgnoredBuilds: []config.IgnoredBuild{
			{Goos: "windows", Goarch: "arm64"},
		},
		IncludeConfig: true,
		SkipBinaries:  false,
		SkipArchives:  false,
	}

	if baseDist == K8sDistro {
		dist.IncludeConfig = false
	}

	if baseDist == K8sDistro || fips {
		dist.Goos = []string{"linux"}
		dist.IgnoredBuilds = nil
		dist.SkipBinaries = true
	}

	if fips {
		dist.SkipArchives = true
	}

	return dist
}

func Blobs(dist Distribution) []config.Blob {
	if dist.SkipBinaries {
		return nil
	}

	return []config.Blob{
		Blob(dist),
	}
}

func Blob(dist Distribution) config.Blob {
	version := "{{ .Version }}"

	if dist.Nightly {
		version = "nightly"
	}

	return config.Blob{
		Provider:  "s3",
		Region:    "us-east-1",
		Bucket:    "nr-releases",
		Directory: fmt.Sprintf("nrdot-collector-releases/%s/%s", dist.FullName, version),
	}
}

func Builds(dist Distribution) []config.Build {
	return []config.Build{
		Build(dist),
	}
}

// Build configures a goreleaser build.
// https://goreleaser.com/customization/build/
func Build(dist Distribution) config.Build {
	dir := "_build"
	cgo := 0
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

	if dist.Fips {
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
		Ignore:                dist.IgnoredBuilds,
	}
}

func Archives(dist Distribution) []config.Archive {
	if dist.SkipArchives {
		return nil
	}

	return []config.Archive{
		Archive(dist),
	}
}

// Archive configures a goreleaser archive (tarball).
// https://goreleaser.com/customization/archive/
func Archive(dist Distribution) config.Archive {
	files := make([]config.File, 0)
	goos := "windows"

	if dist.IncludeConfig {
		files = append(files, config.File{
			Source: ConfigFile,
		})
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

func Packages(dist Distribution) []config.NFPM {
	if dist.SkipBinaries {
		return nil
	}

	return []config.NFPM{
		Package(dist),
	}
}

// Package configures goreleaser to build a system package.
// https://goreleaser.com/customization/nfpm/
func Package(dist Distribution) config.NFPM {
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

	if dist.IncludeConfig {
		nfpmContents = append(nfpmContents, config.NFPMContent{
			Source:      ConfigFile,
			Destination: path.Join("/etc", dist.FullName, ConfigFile),
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

func DockerImageTags(dist Distribution) []string {
	tags := []string{}
	if dist.Fips {
		tags = append(tags, "{{ .Version }}-fips")
	} else if dist.Nightly {
		tags = append(tags, "{{ .Version }}-nightly")
		tags = append(tags, "nightly")
	} else {
		tags = append(tags, "{{ .Version }}")
		tags = append(tags, "latest")
	}
	return tags
}

func DockerImages(dist Distribution) []config.Docker {
	var r []config.Docker

	for _, arch := range Architectures {
		r = append(r, DockerImage(dist, arch))
	}

	return r
}

// DockerImage configures goreleaser to build a container image.
// https://goreleaser.com/customization/docker/
func DockerImage(dist Distribution, arch string) config.Docker {

	dockerFile := "Dockerfile"

	imageTemplates := make([]string, 0)
	for _, tag := range DockerImageTags(dist) {
		imageTemplates = append(
			imageTemplates,
			fmt.Sprintf("{{ .Env.REGISTRY }}/%s:%s-%s", dist.BaseName, tag, arch),
		)
	}

	label := func(name, template string) string {
		return fmt.Sprintf("--label=org.opencontainers.image.%s={{%s}}", name, template)
	}

	files := make([]string, 0)
	if dist.IncludeConfig {
		files = append(files, ConfigFile)
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

func DockerManifests(dist Distribution) []config.DockerManifest {
	r := make([]config.DockerManifest, 0)

	for _, tag := range DockerImageTags(dist) {
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
