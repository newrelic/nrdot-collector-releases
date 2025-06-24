package helm

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"test/e2e/util/chart"
	"testing"
)

func NewHelmOptions(kubectlOptions *k8s.KubectlOptions, chartVersion string, chartValues map[string]string) *helm.Options {
	installArg := []string{
		"--version", chartVersion,
		"--namespace", kubectlOptions.Namespace,
		"--create-namespace",
		"--dependency-update",
	}
	for key, val := range chartValues {
		installArg = append(installArg, "--set", fmt.Sprintf("%s=%s", key, val))
	}
	return &helm.Options{
		KubectlOptions: kubectlOptions,
		ExtraArgs: map[string][]string{
			"install": installArg,
		},
		// Prevent logging of helm commands to avoid secrets leaking into CI logs
		Logger: logger.Discard,
	}
}

func ApplyChart(t *testing.T, kubectlOptions *k8s.KubectlOptions, chartToInstall chart.Chart, releaseNameSuffix string, testId string) {
	releaseName := fmt.Sprintf("%s-%s", releaseNameSuffix, testId)
	version := chartToInstall.Version()
	values := chartToInstall.RequiredChartValues(testId)
	fqn := chartToInstall.Meta().FullyQualifiedChartName()
	helmOptions := NewHelmOptions(kubectlOptions, version, values)
	helm.AddRepo(t, helmOptions, chart.NrRepoName, chart.NrRepoUrl)
	t.Logf("Installing chart %s:%s with release name %s", fqn, version, releaseName)
	helm.Install(t, helmOptions, fqn, releaseName)
	t.Cleanup(func() {
		t.Log("Cleanup 'ApplyChart': delete helm chart")
		helm.Delete(t, helmOptions, releaseName, true)
	})
}
