package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitserver/OpenInspur/cloud-provider-inspur/cloud-controller-manager/pkg"
	"gitserver/OpenInspur/cloud-provider-inspur/test/e2eutil"
	"log"
	"net/http"
	"time"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
const (
	TestCluster = "test-cluster"
)
var ipchange = "139.198.121.98"
var _ = Describe("InCloud LoadBalancer e2e-test", func() {
	It("Should work as expected in ReUse Mode", func() {
		inCloud := &pkg.InCloud{}
		service := &v1.Service{}
		servicePath := workspace + "/test/loadbalancers/external-http-nginx.yaml"
		serviceName := "external-http-nginx-service"
		Expect(e2eutil.KubectlApply(servicePath)).ShouldNot(HaveOccurred())
		defer func() {
			Expect(e2eutil.KubectlDelete(servicePath)).ShouldNot(HaveOccurred())
			time.Sleep(time.Second * 70)
			Eventually(func() error { return e2eutil.WaitForLoadBalancerDeleted(inCloud,service) }, time.Minute*3, time.Second*20).Should(Succeed())
		}()
		log.Println("Just wait 3 minutes before tests because following procedure is so so so slow ")
		time.Sleep(3 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, serviceName, "default", testEIPAddress)
		}, 2*time.Minute, 20*time.Second).Should(Succeed())

		log.Println("Successfully assign a ip")

		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddress, 8089) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddress, 8090) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response")
	})
	It("Should work as expected when using sample yamls", func() {
		//apply service
		inCloud := &pkg.InCloud{}
		service := &v1.Service{}
		service1Path := workspace + "/test/loadbalancers/external-http-nginx.yaml"
		serviceName := "external-http-nginx"
		Expect(e2eutil.KubectlApply(service1Path)).ShouldNot(HaveOccurred())
		defer func() {
			log.Println("Deleting test svc")
			Expect(e2eutil.KubectlDelete(service1Path)).ShouldNot(HaveOccurred())
			time.Sleep(time.Second * 45)
			Eventually(func() error { return e2eutil.WaitForLoadBalancerDeleted(inCloud,service) }, time.Minute*2, time.Second*15).Should(Succeed())
		}()
		log.Println("Just wait 2 minutes before tests because following procedure is so so so slow ")
		time.Sleep(2 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, serviceName, "default", testEIPAddress)
		}, 3*time.Minute, 20*time.Second).Should(Succeed())
		log.Println("Successfully assign a ip")
		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddress, 8088) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response")

		//update size
		svc, err := k8sclient.CoreV1().Services("default").Get(serviceName, metav1.GetOptions{})
		Expect(err).ShouldNot(HaveOccurred())
		svc, err = k8sclient.CoreV1().Services("default").Update(svc)
		Expect(err).ShouldNot(HaveOccurred(), "Failed to update svc")
		log.Println("Just wait 3 minutes before tests because following procedure is so so so slow ")
		time.Sleep(3 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			output, err := pkg.GetLoadBalancer(inCloud,service)
			if err != nil {
				return err
			}
			if output.EipAddress == ipchange && output.SlbId ==""{
					return nil
			}
			return fmt.Errorf("Lb type is not changed or ip not change")
		}, 3*time.Minute, 20*time.Second).Should(Succeed())
		Eventually(func() int { return e2eutil.GerServiceResponse(ipchange, 8088) }, time.Second*20, time.Second*2).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response after resizing")
	})
})