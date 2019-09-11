package loadbalance

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type Protocol string

const (
	ProtocolTCP   Protocol = "TCP"
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTPS Protocol = "HTTPS"
)

var (
	ErrorListenerPortConflict = fmt.Errorf("Port has been occupied")
	ErrorReuseEIPButNoName    = fmt.Errorf("If you want to reuse an eip , you must specify the name of each port in service")
	ErrorListenerNotFound     = fmt.Errorf("Failed to get listener in cloud")
)

//返回结构体
type Listener struct {
	SLBId         string `json:"slbId"`
	ListenerId    string `json:"listenerId"`
	ListenerName  string `json:"listenerName"`
	Protocol      string `json:"protocol"`
	Port          int    `json:"port"`
	ForwardRule   string `json:"forwardRule"`
	IsHealthCheck bool   `json:"isHealthCheck"`
	BackendServer []*BackendServer
}

//创建结构体,和Listener不一样
type CreateListenerOpts struct {
	SLBId         string   `json:"slbId"`
	ListenerName  string   `json:"listenerName"`
	Protocol      Protocol `json:"protocol"`
	Port          int32    `json:"port"`
	ForwardRule   string   `json:"forwardRule"`
	IsHealthCheck bool     `json:"isHealthCheck"`
}

// GetListeners use should mannually load listener because sometimes we do not need load entire topology. For example, deletion
//GetListeners get listeners by slbid
func GetListeners(config *InCloud) ([]Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	ls, err := describeListenersBySlbId(config.LbUrlPre, token, config.LbId)
	if err != nil {
		return nil, err
	}
	if ls == nil {
		return nil, ErrorNotFoundInCloud
	}
	//TODO: get listener's backend

	return ls, nil
}

//GetListener get listener by listenerid
func GetListener(config *InCloud, listenerId string) (*Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	ls, err := describeListenerByListnerId(config.LbUrlPre, token, config.LbId, listenerId)
	if err != nil {
		return nil, err
	}
	if ls == nil {
		return nil, ErrorNotFoundInCloud
	}
	return ls, nil
}

// get listener for a port or nil if does not exist
func GetListenerForPort(existingListeners []Listener, port corev1.ServicePort) *Listener {
	for _, l := range existingListeners {
		if Protocol(l.Protocol) == toListenersProtocol(port.Protocol) && l.Port == int(port.Port) {
			return &l
		}
	}

	return nil
}

func CreateListener(config *InCloud, opts CreateListenerOpts) (*Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	return createListener(config.LbUrlPre, token, opts)
}

func UpdateListener(config *InCloud, listenerid string, opts CreateListenerOpts) (*Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl)
	if error != nil {
		return nil, error
	}
	return modifyListener(config.LbUrlPre, token, listenerid, opts)
}

func toListenersProtocol(protocol corev1.Protocol) Protocol {
	switch protocol {
	case corev1.ProtocolTCP:
		return ProtocolTCP
	default:
		return Protocol(string(protocol))
	}
}

func (l *Listener) CheckPortConflict() (bool, error) {
	//if l.lb.EIPStrategy != ReuseEIP {
	//	return false, nil
	//}
	//listeners, err := l.listenerExec.GetListenersOfLB(*l.lb.Status.QcLoadBalancer.LoadBalancerID, "")
	//if err != nil {
	//	return false, err
	//}
	//for _, list := range listeners {
	//	if *list.ListenerPort == l.ListenerPort {
	//		return true, nil
	//	}
	//}
	return false, nil
}

func (l *Listener) CreateListenerWithBackends() error {
	//err := l.CreateListener()
	//if err != nil {
	//	return err
	//}
	//l.LoadBackends()
	//err = l.backendList.CreateBackends()
	//if err != nil {
	//	klog.Errorf("Failed to create backends of listener %s", l.Name)
	//	return err
	//}
	return nil
}

func (l *Listener) CreateListener() error {
	//if l.Status != nil {
	//	klog.Warningln("Create listener even have a listener")
	//}
	//yes, err := l.CheckPortConflict()
	//if err != nil {
	//	klog.Errorf("Failed to check port conflicts")
	//	return err
	//}
	//if yes {
	//	return ErrorListenerPortConflict
	//}
	//input := &qcservice.AddLoadBalancerListenersInput{
	//	LoadBalancer: l.lb.Status.QcLoadBalancer.LoadBalancerID,
	//	Listeners: []*qcservice.LoadBalancerListener{
	//		{
	//			ListenerProtocol:         &l.Protocol,
	//			BackendProtocol:          &l.Protocol,
	//			BalanceMode:              &l.BalanceMode,
	//			ListenerPort:             &l.ListenerPort,
	//			LoadBalancerListenerName: &l.Name,
	//		},
	//	},
	//}
	//if l.Protocol == "udp" {
	//	input.Listeners[0].HealthyCheckMethod = qcservice.String("udp")
	//}
	//listener, err := l.listenerExec.CreateListener(input)
	//if err != nil {
	//	return err
	//}
	//l.Status = listener
	return nil
}

func (l *Listener) DeleteListener() error {
	//if l.Status == nil {
	//	return fmt.Errorf("Could not delete noexit listener")
	//}
	//klog.Infof("Deleting LoadBalancerListener :'%s'", *l.Status.LoadBalancerListenerID)
	//return l.listenerExec.DeleteListener(*l.Status.LoadBalancerListenerID)
	return nil
}

func (l *Listener) NeedUpdate() bool {
	//if l.Status == nil {
	//	return false
	//}
	//if l.BalanceMode != *l.Status.BalanceMode {
	//	return true
	//}
	return false
}

func (l *Listener) UpdateBackends() error {
	//l.LoadBackends()
	//useless, err := l.backendList.LoadAndGetUselessBackends()
	//if err != nil {
	//	klog.Errorf("Failed to load backends of listener %s", l.Name)
	//	return err
	//}
	//if len(useless) > 0 {
	//	klog.V(1).Infof("Delete useless backends")
	//	err := l.backendExec.DeleteBackends(useless...)
	//	if err != nil {
	//		klog.Errorf("Failed to delete useless backends of listener %s", l.Name)
	//		return err
	//	}
	//}
	//for _, b := range l.backendList.Items {
	//	err := b.LoadQcBackend()
	//	if err != nil {
	//		if err == ErrorBackendNotFound {
	//			err = b.Create()
	//			if err != nil {
	//				klog.Errorf("Failed to create backend of instance %s of listener %s", b.Spec.InstanceID, l.Name)
	//				return err
	//			}
	//		}
	//		return err
	//	} else {
	//		err = b.UpdateBackend()
	//		if err != nil {
	//			klog.Errorf("Failed to update backend %s of listener %s", b.Name, l.Name)
	//			return err
	//		}
	//	}
	//}
	return nil
}

func (l *Listener) UpdateListener() error {
	//err := l.LoadListener()
	////create if not exist
	//if err == ErrorListenerNotFound {
	//	err = l.CreateListenerWithBackends()
	//	if err != nil {
	//		klog.Errorf("Failed to create backends of listener %s of loadbalancer %s", l.Name, l.lb.Name)
	//		return err
	//	}
	//	return nil
	//}
	//if err != nil {
	//	klog.Errorf("Failed to load listener %s in incloud", l.Name)
	//	return err
	//}
	//err = l.UpdateBackends()
	//if err != nil {
	//	return err
	//}
	//if !l.NeedUpdate() {
	//	return nil
	//}
	//klog.Infof("Modifying balanceMode of LoadBalancerTCPListener :'%s'", *l.Status.LoadBalancerListenerID)
	//return l.listenerExec.ModifyListener(*l.Status.LoadBalancerListenerID, l.BalanceMode)
	return nil
}

func checkPortInService(service *corev1.Service, port int) *corev1.ServicePort {
	for index, p := range service.Spec.Ports {
		if int(p.Port) == port {
			return &service.Spec.Ports[index]
		}
	}
	return nil
}

func GetListenerPrefix(service *corev1.Service) string {
	return fmt.Sprintf("listener_%s_%s_", service.Namespace, service.Name)
}
