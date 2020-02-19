package e2eutil

import (
	"context"
	"fmt"
	"gitserver/OpenInspur/cloud-provider-inspur/cloud-controller-manager/pkg"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"os/exec"
	"time"
)

func WaitForController(c *kubernetes.Clientset, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		controller, err := c.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			fmt.Println("Cannot find controller")
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if controller.Status.ReadyReplicas == 1 {
			return true, nil
		}
		return false, nil
	})
	return err
}

func KubectlApply(filename string) error {
	cmd := exec.Command("kubectl", "apply", "-f", filename)
	str, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("kubectl apply failed, error :%s\n", str)
	}
	return err
}

func KubectlDelete(filename string) error {
	ctx, cancle := context.WithTimeout(context.Background(), time.Second*20)
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "-f", filename)
	defer cancle()
	return cmd.Run()
}

func ServiceHasEIP(c *kubernetes.Clientset, name, namespace, ip string) error {
	service, err := c.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		if ip != "" && service.Status.LoadBalancer.Ingress[0].IP != ip {
			return fmt.Errorf("got a different ip")
		}
		return nil
	}
	return fmt.Errorf("Still got no ip")
}

func GerServiceResponse(ip string, port int) int {
	url := fmt.Sprintf("http://%s:%d", ip, port)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error in sending request,err: " + err.Error())
		return -1
	}
	return resp.StatusCode
}

func WaitForLoadBalancerDeleted(config *pkg.InCloud,service *v1.Service) error {
	output, err := pkg.GetLoadBalancer(config,service)
	if err != nil {
		return err
	}
	if output == nil {
		return nil
	}
	log.Printf("id:%s, name:%s, status:%s", output.SlbId, output.SlbName, output.State)
	return fmt.Errorf("LB has not been deleted")
}
