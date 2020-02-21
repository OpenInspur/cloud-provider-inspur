package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitserver/OpenInspur/cloud-provider-inspur/cloud-controller-manager/pkg"
	"gitserver/OpenInspur/cloud-provider-inspur/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"time"
)
const (
	TestCluster = "test-cluster"
)
var ipchange = "139.198.121.98"
var _ = Describe("InCloud LoadBalancer e2e-test", func() {
	It("Should work as expected in ReUse Mode", func() {
		servicePath := workspace + "/test/test-case/case.yaml"
		service1Name := "case1"
		service2Name := "case2"
		Expect(e2eutil.KubectlApply(servicePath)).ShouldNot(HaveOccurred())
		defer func() {
			_, err := k8sclient.CoreV1().Services("default").Get(service1Name, metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(e2eutil.KubectlDelete(servicePath)).ShouldNot(HaveOccurred())
		}()
		log.Println("Just wait 3 minutes before tests because following procedure is so so so slow ")
		time.Sleep(3 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, service1Name, "default", testEIPAddress)
		}, 2*time.Minute, 20*time.Second).Should(Succeed())
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, service2Name, "default", testEIPAddress)
		}, 1*time.Minute, 5*time.Second).Should(Succeed())
		log.Println("Successfully assign a ip")

		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddress, 8089) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddress, 8090) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response")
	})
	It("Should work as expected when using sample yamls", func() {
		service1Path := workspace + "/test/loadbalancers/external-http-nginx.yaml"
		serviceName := "external-http-nginx"
		Expect(e2eutil.KubectlApply(service1Path)).ShouldNot(HaveOccurred())

		defer func() {
			_, err := k8sclient.CoreV1().Services("default").Get(serviceName, metav1.GetOptions{})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(e2eutil.KubectlDelete(service1Path)).ShouldNot(HaveOccurred())
		}()
		log.Println("Just wait 2 minutes before tests because following procedure is so so so slow ")
		time.Sleep(2 * time.Minute)
		log.Println("Wake up, we can test now")
		Eventually(func() error {
			return e2eutil.ServiceHasEIP(k8sclient, serviceName, "default", testEIPAddress)
		}, 3*time.Minute, 20*time.Second).Should(Succeed())
		log.Println("Successfully assign a ip")
		Eventually(func() int { return e2eutil.GerServiceResponse(testEIPAddress, 80) }, time.Second*20, time.Second*5).Should(Equal(http.StatusOK))
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
			output, err := pkg.GetLoadBalancer(ic,svc)
			if err != nil {
				return err
			}
			if output.EipAddress == ipchange {
					return nil
			}
			return fmt.Errorf("ip not change")
		}, 3*time.Minute, 20*time.Second).Should(Succeed())
		Eventually(func() int { return e2eutil.GerServiceResponse(ipchange, 8088) }, time.Second*20, time.Second*2).Should(Equal(http.StatusOK))
		log.Println("Successfully get a 200 response after resizing")
	})
})