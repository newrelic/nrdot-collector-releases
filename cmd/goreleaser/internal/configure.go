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
	"strings"

	"github.com/goreleaser/goreleaser/v2/pkg/config"
)

const (
	ArmArch = "arm"

	HostDistro = "nrdot-collector-host"
	K8sDistro  = "nrdot-collector-k8s"

	EnvRegistry = "{{ .Env.REGISTRY }}"

	BinaryNamePrefix = "nrdot-collector"
	ImageNamePrefix  = "nrdot-collector"
)

var (
	ImagePrefixes = []string{EnvRegistry}
	Architectures = []string{"amd64", "arm64"}
	SkipBinaries  = map[string]bool{
		K8sDistro: true,
	}
	NfpmDefaultConfig = map[string]string{
		HostDistro: "config.yaml",
		// k8s missing due to not packaged via nfpm
	}
	DockerIncludedConfigs = map[string][]string{
		HostDistro: {"config.yaml"},
		K8sDistro:  {"config-daemonset.yaml", "config-deployment.yaml"},
	}
	K8sDockerSkipArchs = map[string]bool{"arm": true, "386": true}
	K8sGoos            = []string{"linux"}
	K8sArchs           = []string{"amd64", "arm64"}
)

func Generate(dist string, nightly bool) config.Project {

	projectName := "nrdot-collector-releases"
	disableRelease := "false"

	if nightly {
		projectName = "nrdot-collector-releases-nightly"
		disableRelease = "true"
	}

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
		Dockers:         DockerImages(dist, nightly),
		DockerManifests: DockerManifests(dist, nightly),
		Signs:           Sign(),
		Version:         2,
		Changelog:       config.Changelog{Disable: "true"},
		Snapshot: config.Snapshot{
			VersionTemplate: "{{ incpatch .Version }}-SNAPSHOT-{{.ShortCommit}}",
		},
		Blobs: Blobs(dist, nightly),
		Release: config.Release{
			Disable:              disableRelease,
			Draft:                true,
			UseExistingDraft:     true,
			ReplaceExistingDraft: false,
		},
	}
}

func Blobs(dist string, nightly bool) []config.Blob {
	if skip, ok := SkipBinaries[dist]; ok && skip {
		return nil
	}
	version := "{{ .Version }}"

	if nightly {
		version = "nightly"
	}

	return []config.Blob{
		{
			Provider:  "s3",
			Region:    "us-east-1",
			Bucket:    "nr-releases",
			Directory: fmt.Sprintf("nrdot-collector-releases/%s/%s", dist, version),
		},
	}
}

func Builds(dist string) []config.Build {
	return []config.Build{
		Build(dist),
	}
}

// Build configures a goreleaser build.
// https://goreleaser.com/customization/build/
func Build(dist string) config.Build {
	goos := []string{"linux", "windows"}
	archs := Architectures

	if dist == K8sDistro {
		goos = K8sGoos
		archs = K8sArchs
	}

	return config.Build{
		ID:     dist,
		Dir:    "_build",
		Binary: dist,
		BuildDetails: config.BuildDetails{
			Env:     []string{"CGO_ENABLED=0"},
			Flags:   []string{"-trimpath"},
			Ldflags: []string{"-s", "-w"},
		},
		Goos:   goos,
		Goarch: archs,
		Ignore: IgnoreBuildCombinations(dist),
	}
}

func IgnoreBuildCombinations(dist string) []config.IgnoredBuild {
	if dist == K8sDistro {
		return nil
	}
	return []config.IgnoredBuild{
		{Goos: "windows", Goarch: "arm64"},
	}
}

func ArmVersions(dist string) []string {
	if dist == K8sDistro {
		return nil
	}
	return []string{"7"}
}

func Archives(dist string) []config.Archive {
	if skip, ok := SkipBinaries[dist]; ok && skip {
		return nil
	}
	return []config.Archive{
		Archive(dist),
	}
}

// Archive configures a goreleaser archive (tarball).
// https://goreleaser.com/customization/archive/
func Archive(dist string) config.Archive {
	return config.Archive{
		ID:           dist,
		NameTemplate: "{{ .Binary }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}{{ if .Mips }}_{{ .Mips }}{{ end }}",
		Builds:       []string{dist},
		FormatOverrides: []config.FormatOverride{
			{
				Goos: "windows", Formats: []string{"zip"},
			},
		},
	}
}

func Packages(dist string) []config.NFPM {
	if skip, ok := SkipBinaries[dist]; ok && skip {
		return nil
	}
	return []config.NFPM{
		Package(dist),
	}
}

// Package configures goreleaser to build a system package.
// https://goreleaser.com/customization/nfpm/
func Package(dist string) config.NFPM {
	nfpmContents := []config.NFPMContent{
		{
			Source:      fmt.Sprintf("%s.service", dist),
			Destination: path.Join("/lib", "systemd", "system", fmt.Sprintf("%s.service", dist)),
		},
		{
			Source:      fmt.Sprintf("%s.conf", dist),
			Destination: path.Join("/etc", dist, fmt.Sprintf("%s.conf", dist)),
			Type:        "config|noreplace",
		},
	}
	if defaultConfig, ok := NfpmDefaultConfig[dist]; ok {
		nfpmContents = append(nfpmContents, config.NFPMContent{
			Source:      defaultConfig,
			Destination: path.Join("/etc", dist, "config.yaml"),
			Type:        "config",
		})
	}
	return config.NFPM{
		ID:          dist,
		Builds:      []string{dist},
		Formats:     []string{"deb", "rpm"},
		License:     "Apache 2.0",
		Description: fmt.Sprintf("NRDOT Collector - %s", dist),
		Maintainer:  "New Relic <caos-team@newrelic.com>",
		Overrides: map[string]config.NFPMOverridables{
			"rpm": {
				Dependencies: []string{"/bin/sh"},
			},
		},
		NFPMOverridables: config.NFPMOverridables{
			PackageName: dist,
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

func DockerImages(dist string, nightly bool) []config.Docker {
	var r []config.Docker
	for _, arch := range Architectures {
		if dist == K8sDistro && K8sDockerSkipArchs[arch] {
			continue
		}
		switch arch {
		case ArmArch:
			for _, vers := range ArmVersions(dist) {
				r = append(r, DockerImage(dist, nightly, arch, vers))
			}
		default:
			r = append(r, DockerImage(dist, nightly, arch, ""))
		}
	}
	return r
}

// DockerImage configures goreleaser to build a container image.
// https://goreleaser.com/customization/docker/
func DockerImage(dist string, nightly bool, arch string, armVersion string) config.Docker {
	dockerArchName := archName(arch, armVersion)
	imageTemplates := make([]string, 0)

	imagePrefixes := ImagePrefixes
	prefixFormat := "%s/%s:{{ .Version }}-%s"
	latestPrefixFormat := "%s/%s:latest-%s"

	if nightly {
		prefixFormat = "%s/%s:{{ .Version }}-nightly-%s"
		latestPrefixFormat = "%s/%s:nightly-%s"
	}

	for _, prefix := range imagePrefixes {
		dockerArchTag := strings.ReplaceAll(dockerArchName, "/", "")
		imageTemplates = append(
			imageTemplates,
			fmt.Sprintf(prefixFormat, prefix, imageName(dist), dockerArchTag),
			fmt.Sprintf(latestPrefixFormat, prefix, imageName(dist), dockerArchTag),
		)
	}

	label := func(name, template string) string {
		return fmt.Sprintf("--label=org.opencontainers.image.%s={{%s}}", name, template)
	}
	files := make([]string, 0)
	if configFiles, ok := DockerIncludedConfigs[dist]; ok {
		for _, configFile := range configFiles {
			files = append(files, configFile)
		}
	}
	return config.Docker{
		ImageTemplates: imageTemplates,
		Dockerfile:     "Dockerfile",

		Use: "buildx",
		BuildFlagTemplates: []string{
			"--pull",
			fmt.Sprintf("--platform=linux/%s", dockerArchName),
			label("created", ".Date"),
			label("name", ".ProjectName"),
			label("revision", ".FullCommit"),
			label("version", ".Version"),
			label("source", ".GitURL"),
			"--label=org.opencontainers.image.licenses=Apache-2.0",
		},
		Files:  files,
		Goos:   "linux",
		Goarch: arch,
		Goarm:  armVersion,
	}
}

func DockerManifests(dist string, nightly bool) []config.DockerManifest {
	r := make([]config.DockerManifest, 0)

	imagePrefixes := ImagePrefixes

	for _, prefix := range imagePrefixes {
		if nightly {
			r = append(r, DockerManifest(prefix, "nightly", dist, nightly))
		} else {
			r = append(r, DockerManifest(prefix, `{{ .Version }}`, dist, nightly))
			r = append(r, DockerManifest(prefix, "latest", dist, nightly))
		}
	}
	return r
}

// DockerManifest configures goreleaser to build a multi-arch container image manifest.
// https://goreleaser.com/customization/docker_manifest/
func DockerManifest(prefix, version, dist string, nightly bool) config.DockerManifest {
	var imageTemplates []string
	prefixFormat := "%s/%s:%s-%s"

	//if nightly {
	//	prefixFormat = "%s/%s:%s-nightly-%s"
	//}

	for _, arch := range Architectures {
		if dist == K8sDistro {
			if _, ok := K8sDockerSkipArchs[arch]; ok {
				continue
			}
		}
		switch arch {
		case ArmArch:
			for _, armVers := range ArmVersions(dist) {
				dockerArchTag := strings.ReplaceAll(archName(arch, armVers), "/", "")
				imageTemplates = append(
					imageTemplates,
					fmt.Sprintf(prefixFormat, prefix, imageName(dist), version, dockerArchTag),
				)
			}
		default:
			imageTemplates = append(
				imageTemplates,
				fmt.Sprintf(prefixFormat, prefix, imageName(dist), version, arch),
			)
		}
	}

	return config.DockerManifest{
		NameTemplate:   fmt.Sprintf("%s/%s:%s", prefix, imageName(dist), version),
		ImageTemplates: imageTemplates,
	}
}

// imageName translates a distribution name to a container image name.
func imageName(dist string) string {
	return strings.Replace(dist, BinaryNamePrefix, ImageNamePrefix, 1)
}

// archName translates architecture to docker platform names.
func archName(arch, armVersion string) string {
	switch arch {
	case ArmArch:
		return fmt.Sprintf("%s/v%s", arch, armVersion)
	default:
		return arch
	}
}

func Sign() []config.Sign {
	return []config.Sign{
		{
			Artifacts: "all",
			Signature: "${artifact}.sig",
			Args: []string{
				"--batch",
				"-u",
				"{{ .Env.GPG_FINGERPRINT }}",
				"--output",
				"${signature}",
				"--detach-sign",
				"${artifact}",
			},
		},
	}
}
