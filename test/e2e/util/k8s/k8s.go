package k8s

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	envutil "test/e2e/util/env"
	"testing"
	"time"
)

func NewKubectlOptions(namespacePrefix string) *k8s.KubectlOptions {
	namespace := newTestNamespace(namespacePrefix, envutil.GetDistro())
	contextName := envutil.GetK8sContextName()
	return k8s.NewKubectlOptions(contextName, "", namespace)
}

func newTestNamespace(namespacePrefix string, distro string) string {
	return fmt.Sprintf("%s-%s", namespacePrefix, distro)
}

func WaitForCollectorReady(tb testing.TB, kubectlOptions *k8s.KubectlOptions, waitForSelector metav1.ListOptions, collectorContainerName string) corev1.Pod {
	logCollectorLogsOnFail(tb, kubectlOptions, waitForSelector, collectorContainerName)
	// ensure to fail before running into overall test timeout to allow cleanups to run
	k8s.WaitUntilNumPodsCreated(tb, kubectlOptions, waitForSelector, 1, 6, 10*time.Second)
	pods := k8s.ListPods(tb, kubectlOptions, waitForSelector)
	for _, pod := range pods {
		// ensure to fail before running into overall test timeout to allow cleanups to run
		k8s.WaitUntilPodAvailable(tb, kubectlOptions, pod.Name, 6, 10*time.Second)
	}
	return pods[0]
}

func logCollectorLogsOnFail(tb testing.TB, kubectlOptions *k8s.KubectlOptions, waitForSelector metav1.ListOptions, collectorContainerName string) {
	tb.Cleanup(func() {
		if tb.Failed() {
			pods := k8s.ListPods(tb, kubectlOptions, waitForSelector)
			var logs []string
			for _, pod := range pods {
				logs = append(logs, fmt.Sprintf("==========START Collector logs for pod %s:==========", pod.Name))
				logs = append(logs, k8s.GetPodLogs(tb, kubectlOptions, &pod, collectorContainerName))
				logs = append(logs, fmt.Sprintf("==========END Collector logs for pod %s:==========", pod.Name))
			}
			toLog := strings.Join(logs, "\n")
			tb.Log(toLog)
		}
	})
}
