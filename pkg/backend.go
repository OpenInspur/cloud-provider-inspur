package pkg

import (
	"fmt"
	"gitserver/kubernetes/inspur-cloud-controller-manager/pkg/common"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"reflect"
	"time"
)

var ErrorBackendNotFound = fmt.Errorf("Cannot find backend")

type Backend struct {
	BackendId   string `json:"backendId"`
	ListenerId  string `json:"listenerId"`
	ServerId    string `json:"ServerId"`
	Port        int    `json:"port"`
	ServerName  string `json:"serverName"`
	ServerIp    string `json:"serverIp"`
	ServierType string `json:"type"`
	Weight      int    `json:"weight"`
}

type CreateBackendOpts struct {
	SLBId      string           `json:"slbId"`
	ListenerId string           `json:"listenerId"`
	Servers    []*BackendServer `json:"servers"`
}

type BackendServer struct {
	ServerId    string `json:"serverId"`
	Port        int    `json:"port"`
	ServerName  string `json:"serverName"`
	ServerIp    string `json:"serverIp"`
	ServierType string `json:"type"`
	Weight      int    `json:"weight"`
}

type BackendList struct {
	code    string           `json:"code"`
	Message string           `json:"message"`
	Data    []*BackendServer `json:"data"`
}

type SlbResponse struct {
	RegionId          string    `json:"regionId"`
	CreatedTime       time.Time `json:"createdTime"`
	ExpiredTime       time.Time `json:"expiredTime"`
	SpecificationId   string    `json:"specificationId"`
	SpecificationName string    `json:"specificationName"`
	SlbId             string    `json:"slbId"`
	SlbName           string    `json:"slbName"`
	Scheme            string    `json:"scheme"`
	BusinessIp        string    `json:"businessIp"`
	VpcId             string    `json:"vpcId"`
	SubnetId          string    `json:"subnetId"`
	EipId             string    `json:"eipId"`
	EipAddress        string    `json:"eipAddress"`
	ListenerCount     int       `json:"listenerCount"`
	ChargeType        string    `json:"chargeType"`
	PurchaseTime      string    `json:"purchaseTime"`
	Type              string    `json:"type"`
	State             string    `json:"state"`
}

func CreateBackends(config *InCloud, opts CreateBackendOpts) (*BackendList, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	return createBackend(config.LbUrlPre, token, opts)
}

func UpdateBackends(config *InCloud, service *v1.Service, listener *Listener, backends interface{}) error {
	//先查询listenner关联的backends
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return error
	}
	backs, error := describeBackendservers(config.LbUrlPre, token, listener.SLBId, listener.ListenerId)
	if error != nil {
		klog.Errorf("describeBackendservers failed ", error)
		return error
	}
	nodes, ok := backends.([]*v1.Node)
	if !ok {
		klog.Errorf("skip default backends update for type %s", reflect.TypeOf(backends))
		return nil
	}
	add, del := []*BackendServer{}, []string{}
	// checkout for newly added servers
	for _, node := range nodes {
		found := false
		for _, back := range backs {
			if back.ServerId == string(node.UID) {
				found = true
				break
			}
		}
		if !found {
			addr, err := nodeAddressForLB(node)
			if err != nil {
				if err == ErrorBackendNotFound {
					// Node failure, do not create member
					klog.Warningf("Failed to create LB backend for node %s: %v", node.Name, err)
					continue
				} else {
					return fmt.Errorf("error getting address for node %s: %v", node.Name, err)
				}
			}
			server := new(BackendServer)
			server.ServerId = GetNodeInstanceID(node)
			server.ServerIp = addr
			server.Port = listener.Port
			server.ServerName = node.Name
			server.ServierType = "ECS"
			server.Weight = 10
			add = append(add, server)
			klog.Info("add backend server:%v", add)
		}
	}
	opts := CreateBackendOpts{
		SLBId:      listener.SLBId,
		ListenerId: listener.ListenerId,
		Servers:    add,
	}
	_, err := CreateBackends(config, opts)
	if nil != err {
		klog.Infof("CreateBackends failed: ", err)
		return err
	}
	// check for removed backend servers
	for _, back := range backs {
		found := false
		for _, node := range nodes {
			if back.ServerId == string(node.UID) {
				found = true
				break
			}
		}
		if !found {
			del = append(del, back.ServerId)
		}
	}
	if len(del) > 0 {
		DeleteBackends(config, service, listener.ListenerId, del)
		if nil != err {
			klog.Infof("DeleteBackends failed: ", err)
			return err
		}
	}
	return nil
}

func DeleteBackends(config *InCloud, service *v1.Service, listenerId string, backendIdList []string) error {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return error
	}
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		slbid = config.LbId
	}
	error = removeBackendServers(config.LbUrlPre, token, slbid, listenerId, backendIdList)

	return error
}

func GetBackends(config *InCloud, service *v1.Service, listenerId string) ([]Backend, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		slbid = config.LbId
	}
	backends, error := describeBackendservers(config.LbUrlPre, token, slbid, listenerId)
	if nil != error {
		klog.Infof("GetBackends failed: ", error)
		return nil, error
	}
	return backends, nil
}

func NewBackendList(lb *LoadBalancer, listener *Listener) *BackendList {
	//list := make([]*Backend, 0)
	//instanceIDs := lb.GetNodesInstanceIDs()
	//exec := lb.lbExec.(executor.QingCloudListenerBackendExecutor)
	//for _, instance := range instanceIDs {
	//	b := &Backend{
	//		backendExec: exec,
	//		Name:        fmt.Sprintf("backend_%s_%s", listener.Name, instance),
	//		Spec: BackendSpec{
	//			Listener:   listener,
	//			Weight:     1,
	//			Port:       listener.NodePort,
	//			InstanceID: instance,
	//		},
	//	}
	//	list = append(list, b)
	//}
	//return &BackendList{
	//	backendExec: exec,
	//	Listener:    listener,
	//	Items:       list,
	//}
	return nil
}

//func (b *Backend) Create() error {
//	backends := make([]*qcservice.LoadBalancerBackend, 0)
//	backends = append(backends, b.toQcBackendInput())
//	input := &qcservice.AddLoadBalancerBackendsInput{
//		LoadBalancerListener: b.Spec.Listener.Status.LoadBalancerListenerID,
//		Backends:             backends,
//	}
//	return b.backendExec.CreateBackends(input)
//}
//func (b *Backend) DeleteBackend() error {
//	if b.Status == nil {
//		return fmt.Errorf("Backend %s Not found", b.Name)
//	}
//	return b.backendExec.DeleteBackends(*b.Status.LoadBalancerBackendID)
//}
//
//func (b *BackendList) CreateBackends() error {
//	if len(b.Items) == 0 {
//		return fmt.Errorf("No backends to create")
//	}
//	backends := make([]*qcservice.LoadBalancerBackend, 0)
//	for _, item := range b.Items {
//		backends = append(backends, item.toQcBackendInput())
//	}
//	input := &qcservice.AddLoadBalancerBackendsInput{
//		LoadBalancerListener: b.Listener.Status.LoadBalancerListenerID,
//		Backends:             backends,
//	}
//	return b.backendExec.CreateBackends(input)
//}
//
//func (b *BackendList) LoadAndGetUselessBackends() ([]string, error) {
//	backends, err := b.backendExec.GetBackendsOfListener(*b.Listener.Status.LoadBalancerListenerID)
//	if err != nil {
//		return nil, err
//	}
//	useless := make([]string, 0)
//	for _, back := range backends {
//		useful := false
//		for _, i := range b.Items {
//			if *back.LoadBalancerBackendName == i.Name {
//				i.Status = back
//				useful = true
//				break
//			}
//		}
//		if !useful {
//			useless = append(useless, *back.LoadBalancerBackendID)
//		}
//	}
//	return useless, nil
//}
//func (b *Backend) LoadQcBackend() error {
//	backends, err := b.backendExec.GetBackendsOfListener(*b.Spec.Listener.Status.LoadBalancerListenerID)
//	if err != nil {
//		return err
//	}
//	for _, item := range backends {
//		if *item.ResourceID == b.Spec.InstanceID {
//			b.Status = item
//			return nil
//		}
//	}
//	return ErrorBackendNotFound
//}
//
//func (b *Backend) NeedUpdate() bool {
//	if b.Status == nil {
//		err := b.LoadQcBackend()
//		if err != nil {
//			klog.Errorf("Unable to get qc backend %s when check updatable", b.Name)
//			return false
//		}
//	}
//	if b.Spec.Weight != *b.Status.Weight || b.Spec.Port != *b.Status.Port {
//		return true
//	}
//	return false
//}
//
//func (b *Backend) UpdateBackend() error {
//	if !b.NeedUpdate() {
//		return nil
//	}
//	return b.backendExec.ModifyBackend(*b.Status.LoadBalancerBackendID, b.Spec.Weight, b.Spec.Port)
//}
