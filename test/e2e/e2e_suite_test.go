package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitserver/OpenInspur/cloud-provider-inspur/cloud-controller-manager/pkg"
	"gitserver/OpenInspur/cloud-provider-inspur/test/e2eutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cloudprovider "k8s.io/cloud-provider"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

const ControllerName = "cloud-controller-manager"

var (
	workspace     string
	testNamespace string
	k8sclient     *kubernetes.Clientset
	inService     *pkg.InCloud
	loadBalancer   cloudprovider.LoadBalancer
	testEIPID     string
	testEIPAddr   string
)

func getWorkspace() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

var _ = BeforeSuite(func() {
	qcs, err := e2eutil.GetIncloudService()
	Expect(err).ShouldNot(HaveOccurred(), "Failed init qc service")
	inService = qcs
	initInCloudLB()
	Expect(cleanup()).ShouldNot(HaveOccurred())
	testNamespace = os.Getenv("TEST_NS")
	testEIPID = os.Getenv("TEST_EIP")
	testEIPAddr = os.Getenv("TEST_EIP_ADDR")
	Expect(testNamespace).ShouldNot(BeEmpty())
	workspace = getWorkspace() + "/../../.."
	home := homeDir()
	Expect(home).ShouldNot(BeEmpty())
	//read config
	c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	Expect(err).ShouldNot(HaveOccurred(), "Error in load kubeconfig")
	k8sclient = kubernetes.NewForConfigOrDie(c)
	//kubectl apply
	Expect(err).ShouldNot(HaveOccurred(), "Failed to start controller")
	log.Println("Ready for testing")
})

var _ = AfterSuite(func() {
	cmd := exec.Command("kubectl", "delete", "-f", workspace+"/test/manager.yaml")
	Expect(cmd.Run()).ShouldNot(HaveOccurred())
	Expect(cleanup()).ShouldNot(HaveOccurred())
})

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return ""
}

func cleanup() error {
	usedLb1, err := loadBalancer.GetLoadBalancer(ic,service)
	if err != nil {

		return err
	}
	log.Println("Cleanup loadbalancers")
	return loadBalancer.Delete(*usedLb1.LoadBalancerID)
}

func initInCloudLB() {
	lbapi, _ := qcService.LoadBalancer(qcService.Config.Zone)
	jobapi, _ := qcService.Job(qcService.Config.Zone)
	api, _ := qcService.Accesskey(qcService.Config.Zone)
	output, err := api.DescribeAccessKeys(&qc.DescribeAccessKeysInput{
		AccessKeys: []*string{&qcService.Config.AccessKeyID},
	})
	Expect(err).ShouldNot(HaveOccurred())
}