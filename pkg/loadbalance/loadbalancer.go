package loadbalance

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
)

var (
	ErrorNotFoundInCloud   = fmt.Errorf("Cannot find lb in incloud")
	ErrorSGNotFoundInCloud = fmt.Errorf("Cannot find security group in incloud")
)

type LoadBalancer struct {
	//service     *corev1.Service
	//Type        int
	//TCPPorts    []int
	//NodePorts   []int
	//Nodes       []*corev1.Node
	//Name        string
	//clusterName string
	RegionId          string `json:"regionId"`
	CreatedTime       string `json:"createdTime"`
	ExpiredTime       string `json:"expiredTime"`
	SpecificationId   string `json:"specificationId"`
	SpecificationName string `json:"specificationName"`
	SlbId             string `json:"slbId"`
	SlbName           string `json:"slbName"`
	Scheme            string `json:"scheme"`
	BusinessIp        string `json:"businessIp"`
	VpcId             string `json:"vpcId"`
	VpcName           string `json:"vpcName"`
	SubnetId          string `json:"subnetId"`
	EipId             string `json:"eipId"`
	EipAddress        string `json:"eipAddress"`
	ListenerCount     string `json:"listenerCount"`
	SlbType           string `json:"slbType"`
	State             string `json:"state"`
	UserId            string `json:"userId"`
}

type LoadBalancerStatus struct {
	K8sLoadBalancerStatus *corev1.LoadBalancerStatus
}

type NewLoadBalancerOption struct {
	NodeLister corev1lister.NodeLister

	K8sNodes    []*corev1.Node
	K8sService  *corev1.Service
	Context     context.Context
	ClusterName string
}

//GetLoadBalancer by slbid,use incloud api to get lb in cloud, return err if not found
func GetLoadBalancer(config *InCloud) (*LoadBalancer, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	lb, err := describeLoadBalancer(config.LbUrlPre, token, config.LbId)
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, ErrorNotFoundInCloud
	}
	return lb, nil
}

func ModifyLoadBalancer(config *InCloud, slbName string) (*SlbResponse, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	slbResponse, err := modifyLoadBalancer(config.LbUrlPre, token, config.LbId, slbName)
	if err != nil {
		return nil, err
	}
	if slbResponse == nil {
		return nil, ErrorNotFoundInCloud
	}
	return slbResponse, nil
}

// GetNodesInstanceIDs return resource ids for listener to create backends
func (lb *LoadBalancer) GetNodesInstanceIDs() []string {
	//if len(lb.Nodes) == 0 {
	//	return nil
	//}
	result := make([]string, 0)
	//for _, node := range lb.Nodes {
	//	result = append(result, instance.NodeNameToInstanceID(node.Name, lb.nodeLister))
	//}
	return result
}

// ClearNoUseListener delete uneccassary listeners in incloud, used when service ports changed
//func (l *LoadBalancer) ClearNoUseListener() error {
//	if l.Status.QcLoadBalancer == nil {
//		return nil
//	}
//	listeners, err := l.lbExec.GetListenersOfLB(*l.Status.QcLoadBalancer.LoadBalancerID, GetListenerPrefix(l.service))
//	if err != nil {
//		klog.Errorf("Failed to get incloud listeners of lb %s", l.Name)
//		return err
//	}
//
//	for _, listener := range listeners {
//		if util.IntIndex(l.TCPPorts, *listener.ListenerPort) == -1 {
//			err := l.lbExec.DeleteListener(*listener.LoadBalancerListenerID)
//			if err != nil {
//				klog.Errorf("Failed to delete listener %s", *listener.LoadBalancerListenerName)
//				return err
//			}
//		}
//	}
//	return nil
//}

/// -----Shared  functions-------

// GetLoadBalancerName,query slb name by slb id
func GetLoadBalancerName(config *InCloud) string {

}
