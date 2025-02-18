package hostmetrics

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	httphelper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"strings"
	"test/e2e/util/chart"
	helmutil "test/e2e/util/helm"
	k8sutil "test/e2e/util/k8s"
	testutil "test/e2e/util/test"
	"testing"
	"time"
)

const (
	TestNamespace = "mock-hostmetrics"
)

var (
	kubectlOptions *k8s.KubectlOptions
	testChart      chart.MockedBackendChart
	testId         string
)

// TODO: Export from mocked module
type ValidationPayload struct {
	DurationInMillis uint64 `json:"duration"`
	Transactions     uint32 `json:"transactions"`
}

func TestStartupBehavior(t *testing.T) {
	testutil.TagAsFastTest(t)
	kubectlOptions = k8sutil.NewKubectlOptions(TestNamespace)
	testChart = chart.NewMockedBackendChart(kubectlOptions.Namespace)
	testId = testutil.NewTestId()
	helmutil.ApplyChart(t, kubectlOptions, testChart.AsChart(), "hostmetrics-startup", testId)

	t.Run("healthcheck succeeds", func(t *testing.T) {
		collectorPod := k8sutil.WaitForCollectorReady(t, kubectlOptions)

		tunnel := k8s.NewTunnel(kubectlOptions, k8s.ResourceTypePod, collectorPod.Name, 13133, 13133)
		t.Cleanup(tunnel.Close)
		tunnel.ForwardPort(t)

		url := fmt.Sprintf("http://%s/", tunnel.Endpoint())

		httphelper.HttpGetWithRetryWithCustomValidation(t, url, &tls.Config{},
			10, 5*time.Second, func(status int, body string) bool {
				return status == 200 && strings.Contains(body, "Server available")
			})
	})

	t.Run("validation-backend logs indicate processed metrics", func(t *testing.T) {
		k8sutil.WaitForCollectorReady(t, kubectlOptions)
		tunnel := k8s.NewTunnel(kubectlOptions, k8s.ResourceTypeService, "validation-backend", 8080, 8080)
		t.Cleanup(tunnel.Close)
		tunnel.ForwardPort(t)
		url := fmt.Sprintf("http://%s/validate", tunnel.Endpoint())

		httphelper.HttpGetWithRetryWithCustomValidation(t, url, nil, 10, 5*time.Second, func(statusCode int, body string) bool {

			if statusCode != 200 {
				return false
			}

			var payload ValidationPayload
			err := json.NewDecoder(strings.NewReader(body)).Decode(&payload)

			if err != nil {
				fmt.Println(err)
				return false
			}

			return payload.Transactions >= 1
		})
	})
}
