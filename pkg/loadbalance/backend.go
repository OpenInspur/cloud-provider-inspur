package loadbalance

import (
	"fmt"
	_ "k8s.io/api/core/v1"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
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
	RegionId string `json:"regionId"`
	CreatedTime time.Time `json:"createdTime"`
	ExpiredTime time.Time `json:"expiredTime"`
	SpecificationId string `json:"specificationId"`
	SpecificationName string `json:"specificationName"`
	SlbId string `json:"slbId"`
	SlbName string `json:"slbName"`
	Scheme string   `json:"scheme"`
	BusinessIp string `json:"businessIp"`
	VpcId string `json:"vpcId"`
	SubnetId string `json:"subnetId"`
	EipId string `json:"eipId"`
	EipAddress string `json:"eipAddress"`
	ListenerCount int `json:"listenerCount"`
	ChargeType string `json:"chargeType"`
	PurchaseTime string `json:"purchaseTime"`
	Type string `json:"type"`
	State string `json:"state"`
} 

func CreateBackend(config *InCloud, opts CreateBackendOpts) (*BackendList, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	return createBackend(config.LbUrlPre, token, opts)
}

func UpdateBackends(config *InCloud, listener *Listener, backends interface{}) error {

	//先查询listenner关联的backends
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return error
	}
	backs,error := describeBackendservers(config.LbUrlPre,token,listener.SLBId,listener.ListenerId)
	if error != nil {
		glog.Infof("describeBackendservers failed ", error)
		return error
	}
	nodes := backends.([]*v1.Node)
	//剔除nodes里面已经存在的backend
	for index,node := range nodes {
		for i,b := range backs {
			if b.ServerName == node.Name {
				nodes = append(nodes[:index],nodes[index+1:]...)
				backs = append(backs[:i],backs[i+1:]...)
			}
		}
	}
	var bs []*BackendServer
	//根据nodes创建backend
	for in,nod := range nodes {
		server := new(BackendServer)
		server.ServerId =  (string)(nod.UID)
		server.ServerName = nod.Name
		server.ServierType = "ECS"
		server.Weight = 10
		bs[in] = server
	}
	opts :=  CreateBackendOpts {
		SLBId:listener.SLBId,
		ListenerId:listener.ListenerId,
		Servers:bs,
	}

	_,err := CreateBackend(config,opts)
	if nil != err {
		glog.Infof("createBackend failed: ", err)
		return err
	}
	//nodes, ok := backends.([]*v1.Node)
	//if !ok {
	//	glog.Infof("skip default server group update for type %s", reflect.TypeOf(backends))
	//	return nil
	//}
	//TODO:遍历nodes

	// checkout for newly added servers

	// check for removed backend servers

	return nil
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
