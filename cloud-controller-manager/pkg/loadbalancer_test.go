package pkg

import (
	. "github.com/agiledragon/gomonkey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

var (
	listenPort1         int32 = 80
	targetPort1               = intstr.FromInt(8080)
	nodePort1           int32 = 8080
	clusterName               = "clusterName-random"
	serviceUIDNoneExist       = "UID-1234567890-0987654321-1234556"
)
func TestGetLoadBalancer(t *testing.T) {
	config :=&InCloud{}
	service :=&v1.Service{}
	patch1:=ApplyFunc(getServiceAnnotation,func (service *v1.Service, annotationKey string, defaultSetting string) string {
		return "123"
	})
	patch2:=ApplyFunc(getKeyCloakToken,func (requestedSubject, tokenClientId, clientSecret, keycloakUrl string, ic *InCloud) (string, error) {
		return "",nil
	})
	patch3:=ApplyFunc( describeLoadBalancer,func(url, token, slbId string) (*LoadBalancer, error) {
		return &LoadBalancer{
			RegionId:"1",
			VpcId:"2",
		},nil
	})
	defer patch1.Reset()
	defer patch2.Reset()
	defer patch3.Reset()
	lb,err:=GetLoadBalancer(config,service)
	if lb==nil||err!=nil{
		t.Fatal("get load balancer failed")
	}
}

func TestDeleteLoadBalancer(t *testing.T) {
	config :=&InCloud{}
	service := &v1.Service{}
	patch1:=ApplyFunc(getServiceAnnotation,func (service *v1.Service, annotationKey string, defaultSetting string) string {
		return "12"
	})
	patch2:=ApplyFunc(getKeyCloakToken,func (requestedSubject, tokenClientId, clientSecret, keycloakUrl string, ic *InCloud) (string, error) {
		return "",nil
	})
	patch3:=ApplyFunc(deleteLoadBalancer,func (url, token, slbId string) error {
		return nil
	})
	defer patch1.Reset()
	defer patch2.Reset()
	defer patch3.Reset()
	err := DeleteLoadBalancer(config,service)
	if err !=nil{
		t.Fatal(err)
	}
}