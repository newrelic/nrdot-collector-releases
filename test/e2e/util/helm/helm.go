package helm

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"test/e2e/util/chart"
	"testing"
)

func NewHelmOptions(kubectlOptions *k8s.KubectlOptions, chartValues map[string]string) *helm.Options {
	installArg := []string{
		"--namespace", kubectlOptions.Namespace,
		"--create-namespace",
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

func ApplyChart(t *testing.T, kubectlOptions *k8s.KubectlOptions, chart chart.Chart, releaseNameSuffix string, testId string) {
	releaseName := fmt.Sprintf("%s-%s", releaseNameSuffix, testId)
	helmOptions := NewHelmOptions(kubectlOptions, chart.RequiredChartValues())
	helm.Install(t, helmOptions, chart.Meta().ChartPath(), releaseName)
	t.Cleanup(func() {
		t.Log("Cleanup 'ApplyChart': delete helm chart")
		helm.Delete(t, helmOptions, releaseName, true)
	})
}
