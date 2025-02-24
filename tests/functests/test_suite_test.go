package tests

import (
	"context"
	"flag"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	ginkgo_reporters "github.com/onsi/ginkgo/v2/reporters"
	. "github.com/onsi/gomega"
	qe_reporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"kubevirt.io/client-go/kubecli"
)

const (
	testNamespace = "common-instancetype-functest"

	defaultFedoraContainerDisk        = "quay.io/containerdisks/fedora:latest"
	defaultCentos7ContainerDisk       = "quay.io/containerdisks/centos:7-2009"
	defaultCentosStream8ContainerDisk = "quay.io/containerdisks/centos-stream:8"
	defaultCentosStream9ContainerDisk = "quay.io/containerdisks/centos-stream:9"
	defaultUbuntu1804ContainerDisk    = "quay.io/containerdisks/ubuntu:18.04"
	defaultUbuntu2004ContainerDisk    = "quay.io/containerdisks/ubuntu:20.04"
	defaultUbuntu2204ContainerDisk    = "quay.io/containerdisks/ubuntu:22.04"
	defaultValidationOsContainerDisk  = "registry:5000/validation-os-container-disk:latest"
)

var (
	afterSuiteReporters []Reporter
	virtClient          kubecli.KubevirtClient

	fedoraContainerDisk        string
	centos7ContainerDisk       string
	centosStream8ContainerDisk string
	centosStream9ContainerDisk string
	ubuntu1804ContainerDisk    string
	ubuntu2004ContainerDisk    string
	ubuntu2204ContainerDisk    string
	validationOsContainerDisk  string
)

//nolint:gochecknoinits
func init() {
	kubecli.Init()

	flag.StringVar(&fedoraContainerDisk, "fedora-container-disk",
		defaultFedoraContainerDisk, "Fedora container disk used by functional tests")
	flag.StringVar(&centos7ContainerDisk, "centos-7-container-disk",
		defaultCentos7ContainerDisk, "CentOS 7 container disk used by functional tests")
	flag.StringVar(&centosStream8ContainerDisk, "centos-stream-8-container-disk",
		defaultCentosStream8ContainerDisk, "CentOS Stream 8 container disk used by functional tests")
	flag.StringVar(&centosStream9ContainerDisk, "centos-stream-9-container-disk",
		defaultCentosStream9ContainerDisk, "CentOS Stream 9 container disk used by functional tests")
	flag.StringVar(&ubuntu1804ContainerDisk, "ubuntu-1804-container-disk",
		defaultUbuntu1804ContainerDisk, "Ubuntu 18.04 container disk used by functional tests")
	flag.StringVar(&ubuntu2004ContainerDisk, "ubuntu-2004-container-disk",
		defaultUbuntu2004ContainerDisk, "Ubuntu 20.04 container disk used by functional tests")
	flag.StringVar(&ubuntu2204ContainerDisk, "ubuntu-2204-container-disk",
		defaultUbuntu2204ContainerDisk, "Ubuntu 22.04 container disk used by functional tests")
	flag.StringVar(&validationOsContainerDisk, "validation-os-container-disk",
		defaultValidationOsContainerDisk, "Validation OS container disk used by functional tests")
}

func checkDeployedResources() {
	virtualMachineClusterInstancetypes, err := virtClient.VirtualMachineClusterInstancetype().List(context.Background(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	Expect(virtualMachineClusterInstancetypes.Items).ToNot(BeEmpty())

	virtualMachineClusterPreferences, err := virtClient.VirtualMachineClusterPreference().List(context.Background(), metav1.ListOptions{})
	Expect(err).ToNot(HaveOccurred())
	Expect(virtualMachineClusterPreferences.Items).ToNot(BeEmpty())
}

var _ = BeforeSuite(func() {
	var err error
	var config *rest.Config
	kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = kubecli.GetKubevirtClientConfig()
	}
	Expect(err).ToNot(HaveOccurred())

	virtClient, err = kubecli.GetKubevirtClientFromRESTConfig(config)
	Expect(err).ToNot(HaveOccurred())

	namespaceObj := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	_, err = virtClient.CoreV1().Namespaces().Create(context.TODO(), namespaceObj, metav1.CreateOptions{})
	Expect(err).ToNot(HaveOccurred())

	checkDeployedResources()
})

var _ = AfterSuite(func() {
	// Clean up namespaced resources
	err := virtClient.CoreV1().Namespaces().Delete(context.TODO(), testNamespace, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = ReportAfterSuite("TestFunctional", func(report Report) {
	for _, reporter := range afterSuiteReporters {
		ginkgo_reporters.ReportViaDeprecatedReporter(reporter, report) //nolint:staticcheck
	}
})

func TestFunctional(t *testing.T) {
	if qe_reporters.JunitOutput != "" {
		afterSuiteReporters = append(afterSuiteReporters, ginkgo_reporters.NewJUnitReporter(qe_reporters.JunitOutput))
	}
	if qe_reporters.Polarion.Run {
		afterSuiteReporters = append(afterSuiteReporters, &qe_reporters.Polarion)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Functional test suite")
}
