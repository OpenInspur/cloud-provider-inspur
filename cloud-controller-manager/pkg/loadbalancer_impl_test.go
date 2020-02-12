package pkg

import (
	"context"
	. "github.com/agiledragon/gomonkey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestEnsureLoadBalancerDeleted1(t *testing.T) {
	clusterName := ""
	service := &v1.Service{}
	c := &InCloud{

	}
	patch1 := ApplyFunc(GetLoadBalancer, func(config *InCloud, service *v1.Service) (*LoadBalancer, error) {
		return &LoadBalancer{
			RegionId: "123",
		}, nil
	})
	patch2 := ApplyFunc(GetListeners, func(config *InCloud, service *v1.Service) ([]Listener, error) {
		return []Listener{
			{SLBId: "1234"},
		}, nil
	})
	patch3 := ApplyFunc(GetListenerForPort,func (existingListeners []Listener, port v1.ServicePort) *Listener {
		return &Listener{
			SLBId:"1",
		}
	})
	patch4:=ApplyFunc(GetBackends,func (config *InCloud, slbid, listenerId string) ([]Backend, error) {
		return []Backend{
			{BackendId:"11"},
		},nil
	})
	patch5:=ApplyFunc(DeleteBackends,func (config *InCloud, slbid, listenerId string, backendIdList []string) error {
		return nil
	})
	defer patch1.Reset()
	defer patch2.Reset()
	defer patch3.Reset()
	defer patch4.Reset()
	defer patch5.Reset()
	c.EnsureLoadBalancerDeleted(context.TODO(), clusterName, service)
}

func TestEnsureLoadBalancer(t *testing.T) {
	c :=&InCloud{}
	ss := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "https-service",
			UID:         types.UID(serviceUIDNoneExist),
			Annotations: map[string]string{},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
			},
			Type:            v1.ServiceTypeLoadBalancer,
			SessionAffinity: v1.ServiceAffinityNone,
		},
	}
	nn := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "222"},
			Spec: v1.NodeSpec{
				ProviderID: "222",
			},
		},
	}
  c.EnsureLoadBalancer(context.TODO(),clusterName,ss,nn)
}
