package pkg

import (
	"fmt"
	"gitserver/kubernetes/inspur-cloud-controller-manager/cloud-controller-manager/pkg/common"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

type Protocol string

const (
	ProtocolTCP   Protocol = "TCP"
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTPS Protocol = "HTTPS"
)

//返回结构体
type Listener struct {
	SLBId         string `json:"slbId"`
	ListenerId    string `json:"listenerId"`
	ListenerName  string `json:"listenerName"`
	Protocol      string `json:"protocol"`
	Port          int    `json:"port"`
	ForwardRule   string `json:"forwardRule"`
	IsHealthCheck string `json:"isHealthCheck"`
	BackendServer []*BackendServer
}

//创建结构体,和Listener不一样
type CreateListenerOpts struct {
	SLBId              string   `json:"slbId"`
	ListenerName       string   `json:"listenerName"`
	Protocol           Protocol `json:"protocol"`
	Port               int32    `json:"port"`
	ForwardRule        string   `json:"forwardRule"`
	IsHealthCheck      string   `json:"isHealthCheck"`
	TypeHealthCheck    string   `json:"typeHealthCheck"`
	PortHealthCheck    int      `json:"portHealthCheck""`
	PeriodHealthCheck  int      `json:"periodHealthCheck"`
	TimeoutHealthCheck int      `json:"timeoutHealthCheck"`
	MaxHealthCheck     int      `json:"maxHealthCheck"`
	DomainHealthCheck  string   `json:"domainHealthCheck"`
	PathHealthCheck    string   `json:"pathHealthCheck"`
}

// GetListeners use should mannually load listener because sometimes we do not need load entire topology. For example, deletion
//GetListeners get listeners by slbid
func GetListeners(config *InCloud, service *corev1.Service) ([]Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		return nil, ErrorSlbIdNotDefined
	}
	ls, err := describeListenersBySlbId(config.LbUrlPre, token, slbid)
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
func GetListener(config *InCloud, service *corev1.Service, listenerId string) (*Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		return nil, ErrorSlbIdNotDefined
	}
	ls, err := describeListenerByListnerId(config.LbUrlPre, token, slbid, listenerId)
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
		if strings.ToLower(l.Protocol) == strings.ToLower(string(port.Protocol)) && l.Port == int(port.NodePort) {
			return &l
		}
	}
	return nil
}

func CreateListener(config *InCloud, opts CreateListenerOpts) (*Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	return createListener(config.LbUrlPre, token, opts)
}

func UpdateListener(config *InCloud, listenerid string, opts CreateListenerOpts) (*Listener, error) {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return nil, error
	}
	return modifyListener(config.LbUrlPre, token, listenerid, opts)
}

func (l *Listener) DeleteListener(config *InCloud, service *corev1.Service) error {
	token, error := getKeyCloakToken(config.RequestedSubject, config.TokenClientID, config.ClientSecret, config.KeycloakUrl, config)
	if error != nil {
		return error
	}
	slbid := getServiceAnnotation(service, common.ServiceAnnotationInternalSlbId, "")
	if slbid == "" {
		return ErrorSlbIdNotDefined
	}
	error = deleteListener(config.LbUrlPre, token, slbid, l.ListenerId)
	if nil != error {
		klog.Error("Deleting LoadBalancerListener:%v", error)
	}
	//if l.Status == nil {
	//	return fmt.Errorf("Could not delete noexit listener")
	//}
	//klog.Infof("Deleting LoadBalancerListener :'%s'", *l.Status.LoadBalancerListenerID)
	//return l.listenerExec.DeleteListener(*l.Status.LoadBalancerListenerID)
	return nil
}

func checkPortInService(service *corev1.Service, port int) *corev1.ServicePort {
	for index, p := range service.Spec.Ports {
		if int(p.NodePort) == port {
			return &service.Spec.Ports[index]
		}
	}
	return nil
}

func GetListenerPrefix(service *corev1.Service) string {
	return fmt.Sprintf("listener_%s_%s_", service.Namespace, service.Name)
}
