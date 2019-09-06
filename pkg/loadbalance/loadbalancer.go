package loadbalance

import (
	"context"
	"fmt"
	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/incloud"
	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/instance"
	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
)

var (
	ErrorLBNotFoundInCloud = fmt.Errorf("Cannot find lb in incloud")
	ErrorSGNotFoundInCloud = fmt.Errorf("Cannot find security group in incloud")
)

type LoadBalancer struct {
	nodeLister corev1lister.NodeLister
	listeners  []*Listener

	LoadBalancerSpec
	incloud.InCloud
}

type LoadBalancerSpec struct {
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

// NewLoadBalancer create loadbalancer in memory, not in cloud, call api to create a real loadbalancer in incloud
func NewLoadBalancer(opt *NewLoadBalancerOption, config *incloud.InCloud) (*LoadBalancer, error) {
	result := &LoadBalancer{
		nodeLister: opt.NodeLister,
	}
	//result.Name = GetLoadBalancerName(opt.ClusterName, opt.K8sService)
	//lbType := opt.K8sService.Annotations[ServiceAnnotationLoadBalancerType]
	//if lbType == "" {
	//	result.Type = 0
	//} else {
	//	t, err := strconv.Atoi(lbType)
	//	if err != nil {
	//		err = fmt.Errorf("Pls spec a valid value of loadBalancer for service %s, accept values are '0-3',err: %s", opt.K8sService.Name, err.Error())
	//		return nil, err
	//	}
	//	if t > 3 || t < 0 {
	//		err = fmt.Errorf("Pls spec a valid value of loadBalancer for service %s, accept values are '0-3'", opt.K8sService.Name)
	//		return nil, err
	//	}
	//	result.Type = t
	//}
	//if strategy, ok := opt.K8sService.Annotations[ServiceAnnotationLoadBalancerEipStrategy]; ok && strategy == string(ReuseEIP) {
	//	result.EIPStrategy = ReuseEIP
	//} else {
	//	result.EIPStrategy = Exclusive
	//}
	//if source, ok := opt.K8sService.Annotations[ServiceAnnotationLoadBalancerEipSource]; ok {
	//	switch source {
	//	case string(AllocateOnly):
	//		result.EIPAllocateSource = AllocateOnly
	//	case string(UseAvailableOnly):
	//		result.EIPAllocateSource = UseAvailableOnly
	//	case string(UseAvailableOrAllocateOne):
	//		result.EIPAllocateSource = UseAvailableOrAllocateOne
	//	default:
	//		result.EIPAllocateSource = ManualSet
	//	}
	//} else {
	//	result.EIPAllocateSource = ManualSet
	//}
	//t, n := util.GetPortsOfService(opt.K8sService)
	//result.TCPPorts = t
	//result.NodePorts = n
	//result.service = opt.K8sService
	//result.Nodes = opt.K8sNodes
	//result.clusterName = opt.ClusterName
	//if result.EIPAllocateSource == ManualSet {
	//	lbEipIds, hasEip := opt.K8sService.Annotations[ServiceAnnotationLoadBalancerEipIds]
	//	if hasEip {
	//		result.EIPs = strings.Split(lbEipIds, ",")
	//	}
	//}
	result.InCloud = *config
	return result, nil
}

//GetLoadBalancer by slbid,use incloud api to get lb in cloud, return err if not found
func (lb *LoadBalancer) GetLoadBalancer() error {
	config := lb.InCloud
	token, error := incloud.GetKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return error
	}
	lbs, err := incloud.DescribeLoadBalancers(config.LbUrlPre, token, config.LbId)
	if err != nil {
		return err
	}
	if lbs == nil {
		return ErrorLBNotFoundInCloud
	}
	lb.LoadBalancerSpec = *lbs
	return nil
}

// LoadListeners use should mannually load listener because sometimes we do not need load entire topology. For example, deletion
//LoadListeners get listeners by slbid
func (lb *LoadBalancer) LoadListeners() error {
	config := lb.InCloud
	token, error := incloud.GetKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return error
	}
	lbs, err := incloud.DescribeListeners(config.LbUrlPre, token, config.LbId)
	if err != nil {
		return err
	}
	if lbs == nil {
		return ErrorLBNotFoundInCloud
	}
	lb.listeners = lbs
	return nil
}

// GetListeners return listeners of this service
func (lb *LoadBalancer) GetListeners() []*Listener {
	return lb.listeners
}

// NeedResize tell us if we should resize the lb in incloud
//func (l *LoadBalancer) NeedResize() bool {
//	if l.Status.QcLoadBalancer == nil {
//		return false
//	}
//	if l.Type != *l.Status.QcLoadBalancer.LoadBalancerType {
//		return true
//	}
//	return false
//}
//
//func (l *LoadBalancer) NeedChangeIP() (yes bool, toadd []string, todelete []string) {
//	if l.Status.QcLoadBalancer == nil || l.EIPAllocateSource != ManualSet {
//		return
//	}
//	yes = true
//	new := strings.Split(l.service.Annotations[ServiceAnnotationLoadBalancerEipIds], ",")
//	old := make([]string, 0)
//	for _, ip := range l.Status.QcLoadBalancer.Cluster {
//		old = append(old, *ip.EIPID)
//	}
//	for _, ip := range new {
//		if util.StringIndex(old, ip) == -1 {
//			toadd = append(toadd, ip)
//		}
//	}
//	for _, ip := range old {
//		if util.StringIndex(new, ip) == -1 {
//			todelete = append(todelete, ip)
//		}
//	}
//	if len(toadd) == 0 && len(todelete) == 0 {
//		yes = false
//	}
//	return
//}

func (lb *LoadBalancer) EnsureQingCloudLB() error {
	//err := lb.GetLoadBalancer()
	//if err != nil {
	//	if err == ErrorLBNotFoundInCloud {
	//		err = lb.CreateQingCloudLB()
	//		if err != nil {
	//			klog.Errorf("Failed to create lb in incloud of service %s", lb.service.Name)
	//			return err
	//		}
	//		return nil
	//	}
	//	return err
	//}
	//err = lb.UpdateQingCloudLB()
	//if err != nil {
	//	klog.Errorf("Failed to update lb %s in incloud of service %s", lb.Name, lb.service.Name)
	//	return err
	//}
	//lb.GenerateK8sLoadBalancer()
	return nil
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

// GetService return service of this loadbalancer
func (lb *LoadBalancer) GetService() *corev1.Service {
	return lb.service
}

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
//			if err == ErrorLBNotFoundInCloud {
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
//			if err == ErrorLBNotFoundInCloud {
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
	if len(lb.Nodes) == 0 {
		return nil
	}
	result := make([]string, 0)
	for _, node := range lb.Nodes {
		result = append(result, instance.NodeNameToInstanceID(node.Name, lb.nodeLister))
	}
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

// GetLoadBalancerName generate lb name for each service. The name of a service is fixed and predictable
func GetLoadBalancerName(clusterName string, service *corev1.Service) string {
	defaultName := fmt.Sprintf("k8s_lb_%s_%s_%s", clusterName, service.Name, util.GetFirstUID(string(service.UID)))
	annotation := service.GetAnnotations()
	if annotation == nil {
		return defaultName
	}
	if strategy, ok := annotation[ServiceAnnotationLoadBalancerEipStrategy]; ok {
		if strategy == string(ReuseEIP) {
			return fmt.Sprintf("k8s_lb_%s_%s", clusterName, annotation[ServiceAnnotationLoadBalancerEipIds])
		}
	}
	return defaultName
}
