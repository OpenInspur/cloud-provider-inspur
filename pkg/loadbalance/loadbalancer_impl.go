package loadbalance

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"strconv"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider"
)

const (
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/inspur-internal-load-balancer"
	//Listener forwardRule
	ServiceAnnotationLoadBalancerForwardRule = "loadbalancer.inspur.com/forward-rule"
	//Listener isHealthCheck
	ServiceAnnotationLoadBalancerHealthCheck = "loadbalancer.inspur.com/is-healthcheck"
)

// LoadBalancer returns an implementation of LoadBalancer for InCloud.
func (ic *InCloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	glog.V(4).Info("LoadBalancer() called")
	return ic, true
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
func (ic *InCloud) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	//TODO 此处约定为创建集群时loadbalancer slbid注入到cloud-config中
	lb, err := GetLoadBalancer(ic)
	if err != nil {
		glog.Errorf("Failed to call 'GetLoadBalancer' of service %s,slb Id:%s", service.Name, lb.SlbId)
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
	lb, err := GetLoadBalancer(ic)
	if err != nil {
		glog.Error("Failed to GetLoadBalancer by config:%v", ic)
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
func (ic *InCloud) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		glog.V(1).Infof("EnsureLoadBalancer takes total %d seconds", elapsed/time.Second)
	}()

	glog.V(4).Infof("EnsureLoadBalancer(%v, %v, %v, %v, %v, %v, %v)", clusterName, service.Namespace, service.Name,
		service.Spec.LoadBalancerIP, service.Spec.Ports, nodes, service.Annotations)

	if len(nodes) == 0 {
		return nil, fmt.Errorf("there are no available nodes for LoadBalancer service %s/%s", service.Namespace, service.Name)
	}

	lb, err := GetLoadBalancer(ic)
	if err != nil {
		glog.Errorf("Failed to get lb by slbId:%s in incloud of service %s", lb.SlbId, service.Name)
		return nil, err
	}
	ls, err := GetListeners(ic)
	//verify scheme 负载均衡的网络模式，默认参数：internet-facing：公网（默认）internal：内网

	forwardRule := getStringFromServiceAnnotation(service, ServiceAnnotationLoadBalancerForwardRule, "RR")
	healthCheck := getStringFromServiceAnnotation(service, ServiceAnnotationLoadBalancerHealthCheck, "0")
	hc, _ := strconv.ParseBool(healthCheck)

	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports provided for inspur load balancer")
	}
	//create/update Listener
	for portIndex, port := range ports {
		listener := GetListenerForPort(ls, port)
		//port not assigned
		if listener == nil {
			glog.V(4).Infof("Creating listener for port %d", int(port.Port))
			listener, err = CreateListener(ic, CreateListenerOpts{
				SLBId:         lb.SlbId,
				ListenerName:  fmt.Sprintf("listener_%s_%d", lb.SlbId, portIndex),
				Protocol:      Protocol(port.Protocol),
				Port:          port.Port,
				ForwardRule:   forwardRule,
				IsHealthCheck: hc,
			})
			if err != nil {
				// Unknown error, retry later
				return nil, fmt.Errorf("error creating LB listener: %v", err)
			}

		} else {
			//TODO:
			_, erro := UpdateListener(ic, listener.ListenerId, CreateListenerOpts{
				SLBId:         lb.SlbId,
				ListenerName:  fmt.Sprintf("listener_%s_%d", lb.SlbId, portIndex),
				Protocol:      Protocol(port.Protocol),
				Port:          port.Port,
				ForwardRule:   forwardRule,
				IsHealthCheck: hc,
			})
			if erro != nil {
				return nil, fmt.Errorf("Error updating LB listener: %v", err)
			}

		}
		ls, err := GetListener(ic, listener.ListenerId)
		if err != nil {
			return nil, fmt.Errorf("failed to get LB listener %v: %v", ls.SLBId, ls.ListenerId)
		}
		err = UpdateBackends(ic, ls, nodes)
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
	lb, err := GetLoadBalancer(ic)
	if err != nil {
		glog.Error("Failed to GetLoadBalancer by %v", ic)
		return err
	}
	glog.V(4).Infof("UpdateLoadBalancer(%v, %v, %v)", clusterName, lb.SlbName, nodes)

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		glog.V(1).Infof("UpdateLoadBalancer takes total %d seconds", elapsed/time.Second)
	}()

	glog.V(4).Infof("UpdateLoadBalancer(%v, %v, %v, %v, %v, %v, %v)", clusterName, service.Namespace, service.Name,
		service.Spec.LoadBalancerIP, service.Spec.Ports, nodes, service.Annotations)

	if len(nodes) == 0 {
		return fmt.Errorf("there are no available nodes for LoadBalancer service %s/%s", service.Namespace, service.Name)
	}

	//修改负载均衡信息，目前只支持修改名称。

	ls, err := GetListeners(ic)
	//verify scheme 负载均衡的网络模式，默认参数：internet-facing：公网（默认）internal：内网

	forwardRule := getStringFromServiceAnnotation(service, ServiceAnnotationLoadBalancerForwardRule, "RR")
	healthCheck := getStringFromServiceAnnotation(service, ServiceAnnotationLoadBalancerHealthCheck, "0")
	hc, _ := strconv.ParseBool(healthCheck)

	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return fmt.Errorf("no ports provided for inspur load balancer")
	}
	//create/update Listener
	for portIndex, port := range ports {
		listener := GetListenerForPort(ls, port)
		//port not assigned
		if listener == nil {
			glog.V(4).Infof("Creating listener for port %d", int(port.Port))
			listener, err = CreateListener(ic, CreateListenerOpts{
				SLBId:         lb.SlbId,
				ListenerName:  fmt.Sprintf("listener_%s_%d", lb.SlbId, portIndex),
				Protocol:      Protocol(port.Protocol),
				Port:          port.Port,
				ForwardRule:   forwardRule,
				IsHealthCheck: hc,
			})
			if err != nil {
				// Unknown error, retry later
				return fmt.Errorf("error creating LB listener: %v", err)
			}

		} else {
			//TODO:
			_, erro := UpdateListener(ic, listener.ListenerId, CreateListenerOpts{
				SLBId:         lb.SlbId,
				ListenerName:  fmt.Sprintf("listener_%s_%d", lb.SlbId, portIndex),
				Protocol:      Protocol(port.Protocol),
				Port:          port.Port,
				ForwardRule:   forwardRule,
				IsHealthCheck: hc,
			})
			if erro != nil {
				return fmt.Errorf("Error updating LB listener: %v", err)
			}

		}
		ls, err := GetListener(ic, listener.ListenerId)
		if err != nil {
			return fmt.Errorf("failed to get LB listener %v: %v", ls.SLBId, ls.ListenerId)
		}
		UpdateBackends(ic, ls, nodes)
	}

	if err != nil {
		return err
	}
	//status := &v1.LoadBalancerStatus{}
	//status.Ingress = []v1.LoadBalancerIngress{{IP: slbResponse.BusinessIp}}
	//if slbResponse.EipAddress != "" {
	//	status.Ingress = append(status.Ingress, v1.LoadBalancerIngress{IP: slbResponse.EipAddress})
	//}
	return nil

	//startTime := time.Now()
	//defer func() {
	//	elapsed := time.Since(startTime)
	//	klog.V(1).Infof("UpdateLoadBalancer takes total %d seconds", elapsed/time.Second)
	//}()
	//lb, err := ic.genLoadBalancer(ctx, clusterName, service, nodes)
	//if err != nil {
	//	return err
	//}
	//err = lb.GetLoadBalancer()
	//if err != nil {
	//	klog.Errorf("Failed to get lb %s in incloud of service %s", lb.Name, service.Name)
	//	return err
	//}
	//err = lb.LoadListeners()
	//if err != nil {
	//	klog.Errorf("Failed to get listeners of lb %s of service %s", lb.Name, service.Name)
	//	return err
	//}
	//listeners := lb.GetListeners()
	//for _, listener := range listeners {
	//	listener.UpdateListener()
	//}

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

	glog.V(4).Infof("EnsureLoadBalancerDeleted(%v, %v)", clusterName, service.Name)

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		glog.V(1).Infof("EnsureLoadBalancerDeleted takes total %d seconds", elapsed/time.Second)
	}()

	glog.V(4).Infof("EnsureLoadBalancerDeleted(%v, %v, %v, %v, %v, %v)", clusterName, service.Namespace, service.Name,
		service.Spec.LoadBalancerIP, service.Spec.Ports, service.Annotations)

	lb,error := GetLoadBalancer(ic)
	if error != nil {
		glog.V(4).Infof("GetLoadBalancer fail , error :", error)
		return error
	}
	if nil == lb {
		glog.V(4).Infof("there is no such loadbalancer")
		return nil
	}
	ls, err := GetListeners(ic)
	if err != nil {
		glog.V(4).Infof("get ls fail ,error : ",err)
		return err
	}
	////verify scheme 负载均衡的网络模式，默认参数：internet-facing：公网（默认）internal：内网
	//
	//forwardRule := getStringFromServiceAnnotation(service, ServiceAnnotationLoadBalancerForwardRule, "RR")
	//healthCheck := getStringFromServiceAnnotation(service, ServiceAnnotationLoadBalancerHealthCheck, "0")
	//hc, _ := strconv.ParseBool(healthCheck)

	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return  fmt.Errorf("no ports provided for inspur load balancer")
	}
	//the delete order : backend,ls,lb
	for _, port := range ports {
		listener := GetListenerForPort(ls, port)
		//port not assigned
		if listener != nil {
			backends,err := GetBackends(ic,listener.ListenerId)
			if nil != err {
				glog.V(4).Infof("getBackens fail ,error : ",err )
				return err
			}
			if nil != backends {
				var backStringList []string
				for _, backend := range backends {
					backStringList = append(backStringList,backend.ServerId)
				}
				DeleteBackends(ic, listener.ListenerId,backStringList)
			}
			error = listener.DeleteListener(ic)
			if nil != error {
				glog.V(4).Infof("DeleteListener fail ,error : ",err )
				return err
			}
		}
	}
	err = DeleteLoadBalancer(ic)
	if nil != err {
		glog.V(4).Infof("DeleteListener fail ,error : ",err )
		return err
	}
	//startTime := time.Now()
	//defer func() {
	//	elapsed := time.Since(startTime)
	//	glog.V(1).Infof("DeleteLoadBalancer takes total %d seconds", elapsed/time.Second)
	//}()
	//lb, _ := ic.genLoadBalancer(ctx, clusterName, service, nil, true)
	//return lb.DeleteQingCloudLB()
	return nil
}

//getStringFromServiceAnnotation searches a given v1.Service for a specific annotationKey and either returns the annotation's value or a specified defaultSetting
func getStringFromServiceAnnotation(service *v1.Service, annotationKey string, defaultSetting string) string {
	glog.V(4).Infof("getStringFromServiceAnnotation(%v, %v, %v)", service, annotationKey, defaultSetting)
	if annotationValue, ok := service.Annotations[annotationKey]; ok {
		//if there is an annotation for this setting, set the "setting" var to it
		// annotationValue can be empty, it is working as designed
		// it makes possible for instance provisioning loadbalancer without floatingip
		glog.V(4).Infof("Found a Service Annotation: %v = %v", annotationKey, annotationValue)
		return annotationValue
	}
	//if there is no annotation, set "settings" var to the value from cloud config
	glog.V(4).Infof("Could not find a Service Annotation; falling back on cloud-config setting: %v = %v", annotationKey, defaultSetting)
	return defaultSetting
}
