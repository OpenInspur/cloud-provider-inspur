package pkg

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gitserver/kubernetes/inspur-cloud-controller-manager/cloud-controller-manager/pkg/common"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider"
	"k8s.io/klog"
	"net/http"
	"os/exec"
	"strconv"
)

// LoadBalancer returns an implementation of LoadBalancer for InCloud.
func (ic *InCloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	klog.Info("LoadBalancer() called")
	return ic, true
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
func (ic *InCloud) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	//TODO 此处约定为从service yaml的annotation取slbid
	lb, err := GetLoadBalancer(ic, service)
	if err != nil {
		if err == ErrorSlbIdNotDefined {
			klog.Infof("Service:%s/%s isn't inspur loadbalancer type", service.Namespace, service.Name)
			return &v1.LoadBalancerStatus{}, false, nil
		}
		klog.Errorf("Failed to call 'GetLoadBalancer' of service:%s/%s,error:%v", service.Namespace, service.Name, err)
		return nil, false, err
	}

	stat := &v1.LoadBalancerStatus{}
	stat.Ingress = []v1.LoadBalancerIngress{{IP: lb.BusinessIp}}
	if lb.EipAddress != "" {
		stat.Ingress = append(stat.Ingress, v1.LoadBalancerIngress{IP: lb.EipAddress})
	}
	return stat, true, err
}

// GetLoadBalancerName returns the name of the load balancer. Implementations must treat the
// *v1.Service parameter as read-only and not modify it.
func (ic *InCloud) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	lb, err := GetLoadBalancer(ic, service)
	if err != nil {
		if err == ErrorSlbIdNotDefined {
			klog.Infof("Service:%s/%s isn't inspur loadbalancer type", service.Namespace, service.Name)
			return ""
		}
		klog.Errorf("Failed to call 'GetLoadBalancer' of service:%s/%s,error:%v", service.Namespace, service.Name, err)
		return ""
	}
	return lb.SlbName
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
// by inspur
// 这里不创建LoadBalancer，查询LoadBalancer，有则创建Listener以及backend，无则报错
// 改进点：根据service查询后端pod所在节点，只注册pod所在节点到loadbalancer上，当pod漂移时，需要刷新loadbalancer的member；当pod个数变更时，需要刷新loadbalancer的member
func (ic *InCloud) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	lb, err := GetLoadBalancer(ic, service)
	if err != nil {
		if err == ErrorSlbIdNotDefined {
			klog.Infof("Service:%s/%s isn't inspur loadbalancer type", service.Namespace, service.Name)
			return &v1.LoadBalancerStatus{}, nil
		}
		klog.Errorf("Failed to call 'GetLoadBalancer' of service:%s/%s,error:%v", service.Namespace, service.Name, err)
		return nil, err
	}

	ls, err := GetListeners(ic, service)

	svcNodes, erro := getServiceNodes(service, nodes)
	if erro != nil {
		return nil, erro
	}
	if len(svcNodes) == 0 {
		return nil, fmt.Errorf("there are no available nodes for LoadBalancer service %s/%s", service.Namespace, service.Name)
	}
	klog.Infof("EnsureLoadBalancer(%v,%v,%v,%v,%v)", clusterName, service.Namespace, service.Name, len(nodes), len(svcNodes))
	//verify scheme 负载均衡的网络模式，默认参数：internet-facing：公网（默认）internal：内网

	forwardRule := getServiceAnnotation(service, common.ServiceAnnotationLBForwardRule, "RR")
	healthCheck := getServiceAnnotation(service, common.ServiceAnnotationLBHealthCheck, "0")
	hc, _ := strconv.ParseBool(healthCheck)
	hcs := "0"
	if hc {
		hcs = "1"
	}
	ty := getServiceAnnotation(service, common.ServiceAnnotationLBtypeHealthCheck, "tcp")
	hpo, _ := strconv.Atoi(getServiceAnnotation(service, common.ServiceAnnotationLBportHealthCheck, "0"))
	pr, _ := strconv.Atoi(getServiceAnnotation(service, common.ServiceAnnotationLBperiodHealthCheck, "30"))
	ti, _ := strconv.Atoi(getServiceAnnotation(service, common.ServiceAnnotationLBtimeoutHealthCheck, "1"))
	ma, _ := strconv.Atoi(getServiceAnnotation(service, common.ServiceAnnotationLBmaxHealthCheck, "1"))
	do := getServiceAnnotation(service, common.ServiceAnnotationLBdomainHealthCheck, "")
	pa := getServiceAnnotation(service, common.ServiceAnnotationLBpathHealthCheck, "/")
	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports provided for inspur load balancer")
	}
	//create/update Listener
	for portIndex, port := range ports {
		klog.Infof("GetListenerForPort,ls%v,port%v", ls, port)
		listener := GetListenerForPort(ls, port)
		po := port.NodePort

		//port not assigned
		if listener == nil {
			klog.Infof("Creating listener for port %d", po)
			listener, err = CreateListener(ic, CreateListenerOpts{
				SLBId:              lb.SlbId,
				ListenerName:       fmt.Sprintf("listener_%d_%d", int(po), portIndex),
				Protocol:           Protocol(port.Protocol),
				Port:               po,
				ForwardRule:        forwardRule,
				IsHealthCheck:      hcs,
				TypeHealthCheck:    ty,
				PortHealthCheck:    hpo,
				PeriodHealthCheck:  pr,
				TimeoutHealthCheck: ti,
				MaxHealthCheck:     ma,
				DomainHealthCheck:  do,
				PathHealthCheck:    pa,
			})
			if err != nil {
				// Unknown error, retry later
				return nil, fmt.Errorf("error creating LB listener: %v", err)
			}
		} else {
			klog.Infof("Updating listener for port %d", po)
			_, erro := UpdateListener(ic, listener.ListenerId, CreateListenerOpts{
				SLBId:              lb.SlbId,
				ListenerName:       fmt.Sprintf("listener_%d_%d", int(po), portIndex),
				Protocol:           Protocol(port.Protocol),
				Port:               po,
				ForwardRule:        forwardRule,
				IsHealthCheck:      hcs,
				TypeHealthCheck:    ty,
				PortHealthCheck:    hpo,
				PeriodHealthCheck:  pr,
				TimeoutHealthCheck: ti,
				MaxHealthCheck:     ma,
				DomainHealthCheck:  do,
				PathHealthCheck:    pa,
			})
			if erro != nil {
				return nil, fmt.Errorf("Error updating LB listener: %v", err)
			}

		}
		cls, err := GetListener(ic, service, listener.ListenerId)
		if err != nil {
			return nil, fmt.Errorf("failed to get LB listener: %v", listener.ListenerId)
		}
		err = UpdateBackends(ic, cls, svcNodes)
		if err != nil {
			return nil, err
		}
	}

	status := &v1.LoadBalancerStatus{}
	status.Ingress = []v1.LoadBalancerIngress{{IP: lb.BusinessIp}}
	if lb.EipAddress != "" {
		status.Ingress = append(status.Ingress, v1.LoadBalancerIngress{IP: lb.EipAddress})
	}
	return status, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (ic *InCloud) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	lb, err := GetLoadBalancer(ic, service)
	if err != nil {
		if err == ErrorSlbIdNotDefined {
			klog.Infof("Service:%s/%s isn't inspur loadbalancer type", service.Namespace, service.Name)
			return nil
		}
		klog.Errorf("Failed to call 'GetLoadBalancer' of service:%s/%s,error:%v", service.Namespace, service.Name, err)
		return err
	}

	svcNodes, erro := getServiceNodes(service, nodes)
	if erro != nil {
		return erro
	}
	if len(svcNodes) == 0 {
		return fmt.Errorf("there are no available nodes for LoadBalancer service %s/%s", service.Namespace, service.Name)
	}
	klog.Infof("UpdateLoadBalancer(%v,%v,%v,%v,%v)", clusterName, service.Namespace, service.Name, len(nodes), len(svcNodes))

	//修改负载均衡信息，目前只支持修改名称。

	ls, err := GetListeners(ic, service)
	//verify scheme 负载均衡的网络模式，默认参数：internet-facing：公网（默认）internal：内网

	forwardRule := getServiceAnnotation(service, common.ServiceAnnotationLBForwardRule, "RR")
	healthCheck := getServiceAnnotation(service, common.ServiceAnnotationLBHealthCheck, "0")
	hc, _ := strconv.ParseBool(healthCheck)
	hcs := "0"
	if hc {
		hcs = "1"
	}

	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return fmt.Errorf("no ports provided for inspur load balancer")
	}
	//create/update Listener
	for portIndex, port := range ports {
		listener := GetListenerForPort(ls, port)
		po := port.NodePort
		//port not assigned
		if listener == nil {
			klog.Infof("Creating listener for port %d", po)
			listener, err = CreateListener(ic, CreateListenerOpts{
				SLBId:         lb.SlbId,
				ListenerName:  fmt.Sprintf("listener_%d_%d", int(po), portIndex),
				Protocol:      Protocol(port.Protocol),
				Port:          po,
				ForwardRule:   forwardRule,
				IsHealthCheck: hcs,
			})
			if err != nil {
				// Unknown error, retry later
				return fmt.Errorf("error creating LB listener: %v", err)
			}

		} else {
			klog.Infof("Updating listener for port %d", po)
			_, erro := UpdateListener(ic, listener.ListenerId, CreateListenerOpts{
				SLBId:         lb.SlbId,
				ListenerName:  fmt.Sprintf("listener_%d_%d", int(po), portIndex),
				Protocol:      Protocol(port.Protocol),
				Port:          port.NodePort,
				ForwardRule:   forwardRule,
				IsHealthCheck: hcs,
			})
			if erro != nil {
				return fmt.Errorf("Error updating LB listener: %v", err)
			}

		}
		cls, err := GetListener(ic, service, listener.ListenerId)
		if err != nil {
			return fmt.Errorf("failed to get LB listener: %v", listener.ListenerId)
		}
		UpdateBackends(ic, cls, svcNodes)
	}

	if err != nil {
		return err
	}

	return nil
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
// Implementations must treat the *v1.Service parameter as read-only and not modify it.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (ic *InCloud) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	klog.Infof("EnsureLoadBalancerDeleted(%v, %v, %v, %v, %v)", clusterName, service.Namespace, service.Name,
		service.Spec.LoadBalancerIP, service.Spec.Ports)

	lb, error := GetLoadBalancer(ic, service)
	if error != nil {
		if error == ErrorSlbIdNotDefined {
			klog.Infof("Service:%s/%s isn't inspur loadbalancer type", service.Namespace, service.Name)
			return nil
		}
		klog.Errorf("Failed to call 'GetLoadBalancer' of service:%s/%s,error:%v", service.Namespace, service.Name, error)
		return error
	}
	if nil == lb {
		klog.Infof("there is no such loadbalancer")
		return nil
	}
	ls, err := GetListeners(ic, service)
	if err != nil {
		klog.Infof("get ls fail ,error : %v", err)
		return err
	}

	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return fmt.Errorf("no ports provided for inspur load balancer")
	}
	//the delete order : backend,ls,lb
	for _, port := range ports {
		listener := GetListenerForPort(ls, port)
		//port not assigned
		if listener != nil {
			backends, err := GetBackends(ic, lb.SlbId, listener.ListenerId)
			if nil != err {
				klog.Errorf("getBackens fail ,error : %v", err)
				return err
			}
			if nil != backends {
				var backStringList []string
				for _, backend := range backends {
					backStringList = append(backStringList, backend.BackendId)
				}
				DeleteBackends(ic, lb.SlbId, listener.ListenerId, backStringList)
			}
			error = listener.DeleteListener(ic, service)
			if nil != error {
				klog.Infof("DeleteListener fail ,error : %v", err)
				return err
			}
		}
	}
	return nil
}

//getServiceAnnotation searches a given v1.Service for a specific annotationKey and either returns the annotation's value or a specified defaultSetting
func getServiceAnnotation(service *v1.Service, annotationKey string, defaultSetting string) string {
	klog.Infof("getServiceAnnotation(%v, %v, %v)", service, annotationKey, defaultSetting)
	if annotationValue, ok := service.Annotations[annotationKey]; ok {
		//if there is an annotation for this setting, set the "setting" var to it
		// annotationValue can be empty, it is working as designed
		// it makes possible for instance provisioning loadbalancer without floatingip
		klog.Infof("Found a Service Annotation: %v = %v", annotationKey, annotationValue)
		return annotationValue
	}
	//if there is no annotation, set "settings" var to the value from cloud config
	klog.Infof("Could not find a Service Annotation:%s; falling back on default setting:%v", annotationKey, defaultSetting)
	return defaultSetting
}

//getServiceAnnotation searches a given v1.Service for a specific annotationKey and either returns the annotation's value or a specified defaultSetting
func getNodeAnnotation(node *v1.Node, annotationKey string, defaultSetting string) string {
	klog.Infof("getNodeAnnotation(%v,%v,%v,%v)", node.Name, node.Annotations, annotationKey, defaultSetting)
	if annotationValue, ok := node.Annotations[annotationKey]; ok {
		//if there is an annotation for this setting, set the "setting" var to it
		// annotationValue can be empty, it is working as designed
		// it makes possible for instance provisioning loadbalancer without floatingip
		klog.Infof("Found a Node Annotation: %v = %v", annotationKey, annotationValue)
		return annotationValue
	}
	//if there is no annotation, set "settings" var to the value from cloud config
	klog.Infof("Could not find a Node Annotation:%s; falling back on default setting:%v", annotationKey, defaultSetting)
	return defaultSetting
}

// 返回service聚合的pods所在的nodes
func getServiceNodes(service *v1.Service, nodes []*v1.Node) ([]*v1.Node, error) {
	spec := service.Spec
	if spec.Selector["app"] != "" {
		sel := "app=" + spec.Selector["app"]
		cmd1 := exec.Command("cat", "/var/run/secrets/kubernetes.io/serviceaccount/token")
		res1, erro := cmd1.CombinedOutput()
		if erro != nil {
			klog.Errorf("cat /var/run/secrets/kubernetes.io/serviceaccount/token error:%v", erro)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		podsUrl := fmt.Sprintf("https://kubernetes.default.svc.cluster.local/api/v1/namespaces/%s/pods/?labelSelector=%s", service.Namespace, sel)
		req, err := http.NewRequest("GET", podsUrl, nil)
		if err != nil {
			klog.Errorf("Request error %v", err)
			return nil, err
		}
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Authorization", "Bearer "+string(res1))
		res, err := client.Do(req)
		if err != nil {
			klog.Errorf("Response error %v", err)
			return nil, err
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			klog.Errorf("Get response body fail %v", err)
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			klog.Errorf("response not ok:%v,%v", res.StatusCode, string(body))
			return nil, fmt.Errorf("response not ok:%d", res.StatusCode)
		}

		var result v1.PodList
		err = json.Unmarshal(body, &result)
		if err != nil {
			klog.Errorf("json.Unmarshal error:%s,output:%v", err, string(body))
			return nil, err
		}
		//正常情况下，nodes数量大于等于result.Items
		//异常情况下，如node notready，接口传进来的nodes只有正常的nodes如slave2，少于result.Items的nodes数量
		var retNodes = []*v1.Node{}
		for _, item := range result.Items {
			for _, node := range nodes {
				if node.Name == item.Spec.NodeName {
					retNodes = append(retNodes, node)
					break
				}
			}
		}
		return retNodes, nil
	}
	return nil, fmt.Errorf("service:%s/%s dosen't have selector(app)", service.Namespace, service.Name)
}
