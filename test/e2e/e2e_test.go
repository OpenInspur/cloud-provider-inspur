package e2e

import (
"fmt"
	"context"
	"github.com/yunify/qingcloud-cloud-controller-manager/pkg/loadbalance"
	"gitserver/OpenInspur/cloud-provider-inspur/cloud-controller-manager/pkg"
	"gitserver/OpenInspur/cloud-provider-inspur/test/e2eutil"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
"net/http"
"strconv"
"time"
. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"
corev1 "k8s.io/api/core/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
"k8s.io/client-go/util/retry"
)

const (
	listenPort1         int32 = 80
	targetPort1               = intstr.FromInt(8080)
	nodePort1           int32 = 8080
	clusterName               = "test-cluster"
	serviceUIDNoneExist       = "UID-1234567890-0987654321-1234556"
)

var ipchange = "139.198.121.98"
var _ = Describe("InCloud LoadBalancer e2e-test", func() {
	It("Should work as expected in ReUse Mode", func() {
		servicePath := workspace + "/test/loadbalancers/external-http-nginx.yaml"
		serviceName := "external-http-nginx-service"
		Expect(e2eutil.KubectlApply(servicePath)).ShouldNot(HaveOccurred())
		log.Println("Just wait 3 minutes before tests because following procedure is so so so slow ")
		time.Sleep(3 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, serviceName, "default", testEIPAddr)
		}, 2*time.Minute, 20*time.Second).Should(Succeed())

		log.Println("Successfully assign a ip")

		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddr, 8089) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddr, 8090) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response")
	})
	It("Should work as expected when using sample yamls", func() {
		//apply service
		service1Path := workspace + "/test/loadbalancers/external-http-nginx.yaml"
		serviceName := "external-http-nginx"
		Expect(e2eutil.KubectlApply(service1Path)).ShouldNot(HaveOccurred())
		defer func() {
			log.Println("Deleting test svc")
			service, err := c.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
			inCloud := &pkg.InCloud{}
			slb,err:=pkg.GetLoadBalancer(inCloud,service)
			Expect(err).ShouldNot(HaveOccurred())
			slbName := slb.SlbName
			Expect(e2eutil.KubectlDelete(service1Path)).ShouldNot(HaveOccurred())
			//make sure lb is deleted
			lbService, _ := qcService.LoadBalancer("ap2a")
			time.Sleep(time.Second * 45)
			Eventually(func() error { return e2eutil.WaitForLoadBalancerDeleted(lbService, slbName) }, time.Minute*2, time.Second*15).Should(Succeed())
		}()
		log.Println("Just wait 2 minutes before tests because following procedure is so so so slow ")
		time.Sleep(2 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, serviceName, "default", testEIPAddr)
		}, 3*time.Minute, 20*time.Second).Should(Succeed())
		log.Println("Successfully assign a ip")
		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddr, 8088) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response")

		//update size
		svc, err := k8sclient.CoreV1().Services("default").Get(serviceName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		expectedType := 2
		svc.Annotations[loadbalance.ServiceAnnotationLoadBalancerType] = strconv.Itoa(expectedType)
		svc.Annotations[loadbalance.ServiceAnnotationLoadBalancerEipIds] = "eip-e5fxdepa"
		svc, err = k8sclient.CoreV1().Services("default").Update(svc)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to update svc")
		log.Println("Just wait 3 minutes before tests because following procedure is so so so slow ")
		time.Sleep(3 * time.Minute)
		log.Println("Wake up, we can test now")
		lbService, _ := qcService.LoadBalancer("ap2a")
		name := loadbalance.GetLoadBalancerName(TestCluster, svc, nil)
		Eventually(func() error {
			input := &service.DescribeLoadBalancersInput{
				Status:     []*string{service.String("active")},
				SearchWord: &name,
			}
			output, err := lbService.DescribeLoadBalancers(input)
			if err != nil {
				return err
			}
			if len(output.LoadBalancerSet) == 1 && *output.LoadBalancerSet[0].LoadBalancerType == expectedType {
				if len(output.LoadBalancerSet[0].Cluster) == 1 && *output.LoadBalancerSet[0].Cluster[0].EIPAddr == ipchange {
					return nil
				}
			}
			return fmt.Errorf("Lb type is not changed or ip not change")
		}, 3*time.Minute, 20*time.Second).Should(Succeed())
		Eventually(func() int { return e2eutil.GerServiceResponse(ipchange, 8088) }, time.Second*20, time.Second*2).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response after resizing")
	})
})