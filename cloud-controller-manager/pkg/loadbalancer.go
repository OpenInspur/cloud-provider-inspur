package pkg

import (
	"context"
	"fmt"
	"gitserver/kubernetes/inspur-cloud-controller-manager/cloud-controller-manager/pkg/common"
	"k8s.io/api/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
)

var (
	ErrorNotFoundInCloud = fmt.Errorf("Cannot find lb in incloud")
	ErrorSlbIdNotDefined = fmt.Errorf("Could not find Service SLB Id ")
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
	ListenerCount     int    `json:"listenerCount"`
	SlbType           string `json:"slbType"`
	State             string `json:"state"`
	UserId            string `json:"userId"`
}

type LoadBalancerStatus struct {
	K8sLoadBalancerStatus *v1.LoadBalancerStatus
}

type NewLoadBalancerOption struct {
	NodeLister corev1lister.NodeLister

	K8sNodes    []*v1.Node
	K8sService  *v1.Service
	Context     context.Context
	ClusterName string
}

//GetLoadBalancer by slbid,use incloud api to get lb in cloud, return err if not found
func GetLoadBalancer(config *InCloud, service *v1.Service) (*LoadBalancer, error) {
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		return nil, ErrorSlbIdNotDefined
	}
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}

	lb, err := describeLoadBalancer(config.LbUrlPre, token, slbid)
	if err != nil {
		return nil, err
	}
	if lb == nil {
		return nil, ErrorNotFoundInCloud
	}
	return lb, nil
}

func ModifyLoadBalancer(config *InCloud, service *v1.Service, slbName string) (*SlbResponse, error) {
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		return nil, ErrorSlbIdNotDefined
	}
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	slbResponse, err := modifyLoadBalancer(config.LbUrlPre, token, slbid, slbName)
	if err != nil {
		return nil, err
	}
	if slbResponse == nil {
		return nil, ErrorNotFoundInCloud
	}
	return slbResponse, nil
}

func DeleteLoadBalancer(config *InCloud, service *v1.Service) error {
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		return ErrorSlbIdNotDefined
	}
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return error
	}
	error = deleteLoadBalancer(config.LbUrlPre, token, slbid)
	return error
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
