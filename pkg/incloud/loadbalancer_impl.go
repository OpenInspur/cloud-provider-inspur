package incloud

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"strconv"
	"time"

	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/loadbalance"
	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider"
	"k8s.io/klog"
)

const (
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/inspur-internal-load-balancer"
)

var _ cloudprovider.LoadBalancer = &InCloud{}

func (ic *InCloud) genLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*loadbalance.LoadBalancer, error) {
	opt := &loadbalance.NewLoadBalancerOption{
		NodeLister:  ic.nodeInformer.Lister(),
		K8sNodes:    nodes,
		K8sService:  service,
		Context:     ctx,
		ClusterName: clusterName,
	}
	return loadbalance.NewLoadBalancer(opt, ic)
}

func (ic *InCloud) newListener(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node, slbId string) (*loadbalance.Listener, error) {
	lb, err := ic.genLoadBalancer(ctx, clusterName, service, nodes)
	if err != nil {
		klog.Errorf("Failed to call 'genLoadBalancer' of service %s , err is %s", service.Name, err.Error())
		return nil, err
	}
	return loadbalance.NewListener(lb, int(service.Spec.Ports[0].Port), slbId)
}

// LoadBalancer returns an implementation of LoadBalancer for InCloud.
func (ic *InCloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	klog.V(4).Info("LoadBalancer() called")
	return ic, true
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
func (ic *InCloud) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (status *v1.LoadBalancerStatus, exists bool, err error) {
	lb, err := ic.genLoadBalancer(ctx, clusterName, service, nil)
	if err != nil {
		return nil, false, err
	}
	//TODO 此处约定为创建集群时loadbalancer slbid注入到cloud-config中
	err = lb.GetLoadBalancer()
	if err != nil {
		klog.Errorf("Failed to call 'GetLoadBalancer' of service %s,slb Id:%s", service.Name, lb.LbId)
		return nil, false, err
	}

	stat := &v1.LoadBalancerStatus{}
	stat.Ingress = []v1.LoadBalancerIngress{{IP: lb.LoadBalancerSpec.BusinessIp}}
	if lb.LoadBalancerSpec.EipAddress != "" {
		stat.Ingress = append(stat.Ingress, v1.LoadBalancerIngress{IP: lb.LoadBalancerSpec.EipAddress})
	}

	return stat, true, err
}

// GetLoadBalancerName returns the name of the load balancer. Implementations must treat the
// *v1.Service parameter as read-only and not modify it.
func (ic *InCloud) GetLoadBalancerName(_ context.Context, clusterName string, service *v1.Service) string {
	return loadbalance.GetLoadBalancerName(clusterName, service)
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
// by inspur
// 这里不创建LoadBalancer，查询LoadBalancer，有则创建Listener，无则报错
func (ic *InCloud) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		klog.V(1).Infof("EnsureLoadBalancer takes total %d seconds", elapsed/time.Second)
	}()
	//组装LoadBalancer结构
	lb, err := ic.genLoadBalancer(ctx, clusterName, service, nodes)
	if err != nil {
		return nil, err
	}
	err = lb.GetLoadBalancer()
	if err != nil {
		klog.Errorf("Failed to get lb by slbId:%s in incloud of service %s", lb.LbId, service.Name)
		return nil, err
	}
	ll, err := lb.LoadListeners()
	//verify scheme 负载均衡的网络模式，默认参数：internet-facing：公网（默认）internal：内网

	//verify ports
	ports := service.Spec.Ports
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports provided for inspur load balancer")
	}
	for portIndex, port := range ports {
		listener := getListenerForPort(oldListeners, port)
		climit := getStringFromServiceAnnotation(apiService, ServiceAnnotationLoadBalancerConnLimit, "-1")
		connLimit := -1
		tmp, err := strconv.Atoi(climit)
		if err != nil {
			glog.V(4).Infof("Could not parse int value from \"%s\" error \"%v\" failing back to default", climit, err)
		} else {
			connLimit = tmp
		}

		if listener == nil {
			glog.V(4).Infof("Creating listener for port %d", int(port.Port))
			listener, err = listeners.Create(lbaas.lb, listeners.CreateOpts{
				Name:           fmt.Sprintf("listener_%s_%d", name, portIndex),
				Protocol:       listeners.Protocol(port.Protocol),
				ProtocolPort:   int(port.Port),
				ConnLimit:      &connLimit,
				LoadbalancerID: loadbalancer.ID,
			}).Extract()
			if err != nil {
				// Unknown error, retry later
				return nil, fmt.Errorf("error creating LB listener: %v", err)
			}
			provisioningStatus, err := waitLoadbalancerActiveProvisioningStatus(lbaas.lb, loadbalancer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to loadbalance ACTIVE provisioning status %v: %v", provisioningStatus, err)
			}
		} else {
			if connLimit != listener.ConnLimit {
				glog.V(4).Infof("Updating listener connection limit from %d to %d", listener.ConnLimit, connLimit)

				updateOpts := listeners.UpdateOpts{
					ConnLimit: &connLimit,
				}

				_, err := listeners.Update(lbaas.lb, listener.ID, updateOpts).Extract()
				if err != nil {
					return nil, fmt.Errorf("Error updating LB listener: %v", err)
				}

				provisioningStatus, err := waitLoadbalancerActiveProvisioningStatus(lbaas.lb, loadbalancer.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to loadbalance ACTIVE provisioning status %v: %v", provisioningStatus, err)
				}
			}
		}

		glog.V(4).Infof("Listener for %s port %d: %s", string(port.Protocol), int(port.Port), listener.ID)

		// After all ports have been processed, remaining listeners are removed as obsolete.
		// Pop valid listeners.
		oldListeners = popListener(oldListeners, listener.ID)
		pool, err := getPoolByListenerID(lbaas.lb, loadbalancer.ID, listener.ID)
		if err != nil && err != ErrNotFound {
			// Unknown error, retry later
			return nil, fmt.Errorf("error getting pool for listener %s: %v", listener.ID, err)
		}
		if pool == nil {
			glog.V(4).Infof("Creating pool for listener %s", listener.ID)
			pool, err = v2pools.Create(lbaas.lb, v2pools.CreateOpts{
				Name:        fmt.Sprintf("pool_%s_%d", name, portIndex),
				Protocol:    v2pools.Protocol(port.Protocol),
				LBMethod:    lbmethod,
				ListenerID:  listener.ID,
				Persistence: persistence,
			}).Extract()
			if err != nil {
				// Unknown error, retry later
				return nil, fmt.Errorf("error creating pool for listener %s: %v", listener.ID, err)
			}
			provisioningStatus, err := waitLoadbalancerActiveProvisioningStatus(lbaas.lb, loadbalancer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to loadbalance ACTIVE provisioning status %v: %v", provisioningStatus, err)
			}

		}

		glog.V(4).Infof("Pool for listener %s: %s", listener.ID, pool.ID)
		members, err := getMembersByPoolID(lbaas.lb, pool.ID)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("error getting pool members %s: %v", pool.ID, err)
		}
		for _, node := range nodes {
			addr, err := nodeAddressForLB(node)
			if err != nil {
				if err == ErrNotFound {
					// Node failure, do not create member
					glog.Warningf("Failed to create LB pool member for node %s: %v", node.Name, err)
					continue
				} else {
					return nil, fmt.Errorf("error getting address for node %s: %v", node.Name, err)
				}
			}

			if !memberExists(members, addr, int(port.NodePort)) {
				glog.V(4).Infof("Creating member for pool %s", pool.ID)
				_, err := v2pools.CreateMember(lbaas.lb, pool.ID, v2pools.CreateMemberOpts{
					Name:         fmt.Sprintf("member_%s_%d_%s", name, portIndex, node.Name),
					ProtocolPort: int(port.NodePort),
					Address:      addr,
					SubnetID:     lbaas.opts.SubnetID,
				}).Extract()
				if err != nil {
					return nil, fmt.Errorf("error creating LB pool member for node: %s, %v", node.Name, err)
				}

				provisioningStatus, err := waitLoadbalancerActiveProvisioningStatus(lbaas.lb, loadbalancer.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to loadbalance ACTIVE provisioning status %v: %v", provisioningStatus, err)
				}
			} else {
				// After all members have been processed, remaining members are deleted as obsolete.
				members = popMember(members, addr, int(port.NodePort))
			}

			glog.V(4).Infof("Ensured pool %s has member for %s at %s", pool.ID, node.Name, addr)
		}

		// Delete obsolete members for this pool
		for _, member := range members {
			glog.V(4).Infof("Deleting obsolete member %s for pool %s address %s", member.ID, pool.ID, member.Address)
			err := v2pools.DeleteMember(lbaas.lb, pool.ID, member.ID).ExtractErr()
			if err != nil && !isNotFound(err) {
				return nil, fmt.Errorf("error deleting obsolete member %s for pool %s address %s: %v", member.ID, pool.ID, member.Address, err)
			}
			provisioningStatus, err := waitLoadbalancerActiveProvisioningStatus(lbaas.lb, loadbalancer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to loadbalance ACTIVE provisioning status %v: %v", provisioningStatus, err)
			}
		}

		monitorID := pool.MonitorID
		if monitorID == "" && lbaas.opts.CreateMonitor {
			glog.V(4).Infof("Creating monitor for pool %s", pool.ID)
			monitor, err := v2monitors.Create(lbaas.lb, v2monitors.CreateOpts{
				Name:       fmt.Sprintf("monitor_%s_%d", name, portIndex),
				PoolID:     pool.ID,
				Type:       string(port.Protocol),
				Delay:      int(lbaas.opts.MonitorDelay.Duration.Seconds()),
				Timeout:    int(lbaas.opts.MonitorTimeout.Duration.Seconds()),
				MaxRetries: int(lbaas.opts.MonitorMaxRetries),
			}).Extract()
			if err != nil {
				return nil, fmt.Errorf("error creating LB pool healthmonitor: %v", err)
			}
			provisioningStatus, err := waitLoadbalancerActiveProvisioningStatus(lbaas.lb, loadbalancer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to loadbalance ACTIVE provisioning status %v: %v", provisioningStatus, err)
			}
			monitorID = monitor.ID
		} else if lbaas.opts.CreateMonitor == false {
			glog.V(4).Infof("Do not create monitor for pool %s when create-monitor is false", pool.ID)
		}

		if monitorID != "" {
			glog.V(4).Infof("Monitor for pool %s: %s", pool.ID, monitorID)
		}
	}

	if err != nil {
		return nil, err
	}
	err = lb.EnsureQingCloudLB()
	if err != nil {
		return nil, err
	}
	return lb.Status.K8sLoadBalancerStatus, nil
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
// Implementations must treat the *v1.Service and *v1.Node
// parameters as read-only and not modify them.
// Parameter 'clusterName' is the name of the cluster as presented to kube-controller-manager
func (ic *InCloud) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		klog.V(1).Infof("UpdateLoadBalancer takes total %d seconds", elapsed/time.Second)
	}()
	lb, err := ic.genLoadBalancer(ctx, clusterName, service, nodes)
	if err != nil {
		return err
	}
	err = lb.GetLoadBalancer()
	if err != nil {
		klog.Errorf("Failed to get lb %s in incloud of service %s", lb.Name, service.Name)
		return err
	}
	err = lb.LoadListeners()
	if err != nil {
		klog.Errorf("Failed to get listeners of lb %s of service %s", lb.Name, service.Name)
		return err
	}
	listeners := lb.GetListeners()
	for _, listener := range listeners {
		listener.UpdateListener()
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
	//startTime := time.Now()
	//defer func() {
	//	elapsed := time.Since(startTime)
	//	klog.V(1).Infof("DeleteLoadBalancer takes total %d seconds", elapsed/time.Second)
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
