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
	ListenerCount     int `json:"listenerCount"`
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
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl,config)
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
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl,config)
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

func DeleteLoadBalancer(config *InCloud)error{
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl,config)
	if error != nil {
		return error
	}
	error = deleteLoadBalancer(config.LbUrlPre,token,config.LbId)
	return error
}
//// CreateQingCloudLB do create a lb in incloud
//func (l *LoadBalancer) CreateQingCloudLB() error {
//	err := l.EnsureEIP()
//	if err != nil {
//		return err
//	}
//	err = l.EnsureLoadBalancerSecurityGroup()
//	if err != nil {
//		return err
//	}
//
//	createInput := &qcservice.CreateLoadBalancerInput{
//		EIPs:             qcservice.StringSlice(l.EIPs),
//		LoadBalancerType: &l.Type,
//		LoadBalancerName: &l.Name,
//		SecurityGroup:    l.Status.QcSecurityGroup.SecurityGroupID,
//	}
//	lb, err := l.lbExec.Create(createInput)
//	if err != nil {
//		klog.Errorf("Failed to create a lb %s in incloud", l.Name)
//		return err
//	}
//	l.Status.QcLoadBalancer = lb
//	err = l.LoadListeners()
//	if err != nil {
//		klog.Errorf("Failed to generate listener of loadbalancer %s", l.Name)
//		return err
//	}
//	for _, listener := range l.listeners {
//		err = listener.CreateListenerWithBackends()
//		if err != nil {
//			klog.Errorf("Failed to create listener %s of loadbalancer %s", listener.Name, l.Name)
//			return err
//		}
//	}
//	err = l.lbExec.Confirm(*lb.LoadBalancerID)
//	if err != nil {
//		klog.Errorf("Failed to make loadbalancer %s go into effect", l.Name)
//		return err
//	}
//	l.GenerateK8sLoadBalancer()
//	klog.V(1).Infof("Loadbalancer %s created succeefully", l.Name)
//	return nil
//}
//
//// UpdateQingCloudLB update some attrs of incloud lb
//func (l *LoadBalancer) UpdateQingCloudLB() error {
//	if l.Status.QcLoadBalancer == nil {
//		klog.Warningf("Nothing can do before loading incloud loadBalancer %s", l.Name)
//		return nil
//	}
//	lbid := *l.Status.QcLoadBalancer.LoadBalancerID
//	if l.NeedResize() {
//		klog.V(2).Infof("Detect lb size changed, begin to resize the lb %s", l.Name)
//		err := l.lbExec.Resize(*l.Status.QcLoadBalancer.LoadBalancerID, l.Type)
//		if err != nil {
//			klog.Errorf("Failed to resize lb %s", l.Name)
//			return err
//		}
//	}
//
//	if yes, toadd, todelete := l.NeedChangeIP(); yes {
//		klog.V(2).Infof("Adding eips %s to and deleting %s from lb %s", toadd, todelete, l.Name)
//		err := l.lbExec.AssociateEip(lbid, toadd...)
//		if err != nil {
//			klog.Errorf("Failed to add eips %s to lb %s", toadd, l.Name)
//			return err
//		}
//		err = l.lbExec.DissociateEip(lbid, todelete...)
//		if err != nil {
//			klog.Errorf("Failed to add eips %s to lb %s", todelete, l.Name)
//			return err
//		}
//	}
//
//	if l.NeedUpdate() {
//		modifyInput := &qcservice.ModifyLoadBalancerAttributesInput{
//			LoadBalancerName: &l.Name,
//			LoadBalancer:     l.Status.QcLoadBalancer.LoadBalancerID,
//		}
//		err := l.lbExec.Modify(modifyInput)
//		if err != nil {
//			klog.Errorf("Failed to update lb %s in incloud", l.Name)
//			return err
//		}
//	}
//	err := l.LoadListeners()
//	if err != nil {
//		klog.Errorf("Failed to generate listener of loadbalancer %s", l.Name)
//		return err
//	}
//	for _, listener := range l.listeners {
//		err = listener.UpdateListener()
//		if err != nil {
//			klog.Errorf("Failed to create/update listener %s of loadbalancer %s", listener.Name, l.Name)
//			return err
//		}
//	}
//	klog.V(2).Infoln("Clear useless listeners")
//	err = l.ClearNoUseListener()
//	if err != nil {
//		klog.Errorf("Failed to clear listeners of service %s", l.service.Name)
//		return err
//	}
//	err = l.lbExec.Confirm(*l.Status.QcLoadBalancer.LoadBalancerID)
//	if err != nil {
//		klog.Errorf("Failed to make loadbalancer %s go into effect", l.Name)
//		return err
//	}
//	return nil
//}

//func (l *LoadBalancer) deleteListenersOnlyIfOK() (bool, error) {
//	if l.Status.QcLoadBalancer == nil {
//		return false, nil
//	}
//	listeners, err := l.lbExec.GetListenersOfLB(*l.Status.QcLoadBalancer.LoadBalancerID, "")
//	if err != nil {
//		klog.Errorf("Failed to check current listeners of lb %s", l.Name)
//		return false, err
//	}
//	prefix := GetListenerPrefix(l.service)
//	toDelete := make([]*qcservice.LoadBalancerListener, 0)
//	isUsedByAnotherSevice := false
//	for _, listener := range listeners {
//		if !strings.HasPrefix(*listener.LoadBalancerListenerName, prefix) {
//			isUsedByAnotherSevice = true
//		} else {
//			toDelete = append(toDelete, listener)
//		}
//	}
//	if isUsedByAnotherSevice {
//		for _, listener := range toDelete {
//			err = l.lbExec.DeleteListener(*listener.LoadBalancerListenerID)
//			if err != nil {
//				klog.Errorf("Failed to delete listener %s", *listener.LoadBalancerListenerName)
//				return false, err
//			}
//		}
//	}
//	return isUsedByAnotherSevice, nil
//}
//
//func (l *LoadBalancer) DeleteQingCloudLB() error {
//	if l.Status.QcLoadBalancer == nil {
//		err := l.GetLoadBalancer()
//		if err != nil {
//			if err == ErrorNotFoundInCloud {
//				klog.V(1).Infof("Cannot find the lb %s in cloud, maybe is deleted", l.Name)
//				err = l.deleteSecurityGroup()
//				if err != nil {
//					if err == ErrorSGNotFoundInCloud {
//						return nil
//					}
//					klog.Errorf("Failed to delete SecurityGroup of lb %s ", l.Name)
//					return err
//				}
//				return nil
//			}
//			return err
//		}
//	}
//	ok, err := l.deleteListenersOnlyIfOK()
//	if err != nil {
//		return err
//	}
//	if ok {
//		klog.Infof("Detect lb %s is used by another service, delete listeners only", l.Name)
//		return nil
//	}
//	//record eip id before deleting
//	ip := l.Status.QcLoadBalancer.Cluster[0]
//	err = l.lbExec.Delete(*l.Status.QcLoadBalancer.LoadBalancerID)
//	if err != nil {
//		klog.Errorf("Failed to excute deletion of lb %s", *l.Status.QcLoadBalancer.LoadBalancerName)
//		return err
//	}
//	err = l.deleteSecurityGroup()
//	if err != nil {
//		if err == ErrorSGNotFoundInCloud {
//			klog.Warningf("Detect sg %s is deleted", l.Name)
//		} else {
//			klog.Errorf("Failed to delete SecurityGroup of lb %s err '%s' ", l.Name, err)
//			return err
//		}
//	}
//
//	if l.EIPAllocateSource != ManualSet && *ip.EIPName == eip.AllocateEIPName {
//		klog.V(2).Infof("Detect eip %s of lb %s is allocated, release it", *ip.EIPID, l.Name)
//		err := l.eipExec.ReleaseEIP(*ip.EIPID)
//		if err != nil {
//			klog.Errorf("Fail to release  eip %s of lb %s err '%s' ", *ip.EIPID, l.Name, err)
//		}
//	}
//	klog.Infof("Successfully delete loadBalancer '%s'", l.Name)
//	return nil
//}
//
//// NeedUpdate tell us whether an update to loadbalancer is needed
//func (l *LoadBalancer) NeedUpdate() bool {
//	if l.Status.QcLoadBalancer == nil {
//		return false
//	}
//	if l.Name != *l.Status.QcLoadBalancer.LoadBalancerName {
//		return true
//	}
//	return false
//}
//
//// GenerateK8sLoadBalancer get a corev1.LoadBalancerStatus for k8s
//func (l *LoadBalancer) GenerateK8sLoadBalancer() error {
//	if l.Status.QcLoadBalancer == nil {
//		err := l.GetLoadBalancer()
//		if err != nil {
//			if err == ErrorNotFoundInCloud {
//				return nil
//			}
//			klog.Errorf("Failed to load qc loadbalance of %s", l.Name)
//			return err
//		}
//	}
//	status := &corev1.LoadBalancerStatus{}
//	for _, eip := range l.Status.QcLoadBalancer.Cluster {
//		status.Ingress = append(status.Ingress, corev1.LoadBalancerIngress{IP: *eip.EIPAddr})
//	}
//	for _, ip := range l.Status.QcLoadBalancer.EIPs {
//		status.Ingress = append(status.Ingress, corev1.LoadBalancerIngress{IP: *ip.EIPAddr})
//	}
//	if len(status.Ingress) == 0 {
//		return fmt.Errorf("Have no ip yet")
//	}
//	l.Status.K8sLoadBalancerStatus = status
//	return nil
//}

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
