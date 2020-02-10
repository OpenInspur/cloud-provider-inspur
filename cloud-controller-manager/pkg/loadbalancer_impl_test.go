package pkg

import (
	"context"
	. "github.com/agiledragon/gomonkey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
	"time"
)
var (
	vpcId                     = "vpc-2zeaybwqmvn6qgabfd3pe"
	regionId                  = "cn-beijing"
	zoneId                    = "cn-beijing-b"
	certID                    = "1745547945134207_157f665c830"
	listenPort1         int32 = 80
	listenPort2         int32 = 90
	targetPort1               = intstr.FromInt(8080)
	targetPort2               = intstr.FromInt(9090)
	nodePort1           int32 = 8080
	nodePort2           int32 = 9090
	protocolTcp               = v1.ProtocolTCP
	protocolUdp               = v1.ProtocolUDP
	node1                     = "i-bp1bcl00jhxr754tw8vx"
	node2                     = "i-bp1bcl00jhxr754tw8vy"
	clusterName               = "clusterName-random"
	serviceUIDNoneExist       = "UID-1234567890-0987654321-1234556"
	serviceUIDExist           = "c83f8bed-812e-11e9-a0ad-00163e0a3984"
	nodeName                  = "iZuf694l8lw6xvdx6gh7tkZ"

)
type Context2 struct{
   key string
}
func (c *Context2)Va(r context.Context)interface{}{
	return r.Value(c.key)
}
func (c *Context2)Er(r context.Context)error{
	return r.Err()
}
func (c *Context2)Do(r context.Context)<-chan struct{}{
	return r.Done()
}
func (c *Context2)DL(r context.Context)(deadline time.Time, ok bool){
	return r.Deadline()
}
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
	defer patch1.Reset()
	defer patch2.Reset()
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
