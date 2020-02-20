package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitserver/OpenInspur/cloud-provider-inspur/cloud-controller-manager/pkg"
	"gitserver/OpenInspur/cloud-provider-inspur/test/e2eutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)
type ErrorType string
type Error struct {
	Type         ErrorType
	Message      string
	ResourceType string
	Action       string
	ResouceName  string
}
func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

const ControllerName = "cloud-controller-manager"

var (
	workspace     string
	testNamespace string
	k8sclient     *kubernetes.Clientset
	ic     *pkg.InCloud
	incloudLB   pkg.LoadBalancer
	testEIPID     string
	testEIPAddress   string
)

func getWorkspace() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

var _ = BeforeSuite(func() {
	incloud,err := e2eutil.GetInloud()
	Expect(err).ShouldNot(HaveOccurred(), "Failed init qc service")
	ic = incloud
	testNamespace = os.Getenv("TEST_NS")
	Expect(testNamespace).ShouldNot(BeEmpty())
	workspace = getWorkspace() + "/../../.."
	home := homeDir()
	Expect(home).ShouldNot(BeEmpty())
	//read config
	c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	Expect(err).ShouldNot(HaveOccurred(), "Error in load kubeconfig")
	k8sclient = kubernetes.NewForConfigOrDie(c)
	//kubectl apply
	err = e2eutil.WaitForController(k8sclient, testNamespace, ControllerName, time.Second*10, time.Minute*2)
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
	inCloud :=&pkg.InCloud{}
	service := &v1.Service{}
	_, err := pkg.GetLoadBalancer(inCloud,service)
	if err != nil {
		if IsResourceNotFound(err) {
			return nil
		}
		return err
	}
	log.Println("Cleanup loadbalancers")
	return pkg.DeleteLoadBalancer(inCloud,service)
}
func (e *Error) Error() string {
	return fmt.Sprintf("[%s] happened when [%s] type: [%s] name: [%s], msg: [%s]", e.Type, e.Action, e.ResourceType, e.ResouceName, e.Message)
}
func IsResourceNotFound(e error) bool {
	er, ok := e.(*Error)
	if ok && er.Type == "ResourceNotFound" {
		return true
	}
	return false
}
