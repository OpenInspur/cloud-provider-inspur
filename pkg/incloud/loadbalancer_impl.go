package incloud

import (
	"context"
	"time"

	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/loadbalance"
	"k8s.io/api/core/v1"
	"k8s.io/cloud-provider"
	"k8s.io/klog"
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
	lb,err := ic.genLoadBalancer(ctx , clusterName, service, nodes)
	if err != nil {
		klog.Errorf("Failed to call 'genLoadBalancer' of service %s , err is %s", service.Name,err.Error())
		return nil, err
	}
	return loadbalance.NewListener(lb,int(service.Spec.HealthCheckNodePort))
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
	//TODO 此处需要约定loadbalancer name生成规则
	err = lb.GetLoadBalancer(clusterName)
	if err != nil {
		klog.Errorf("Failed to call 'GetLoadBalancer' of service %s", service.Name)
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
		klog.Errorf("Failed to get lb %s in incloud of service %s", lb.Name, service.Name)
		return nil, err
	}
	ll, err := lb.LoadListeners()
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
