package pkg

import (
	."github.com/agiledragon/gomonkey"
	v1 "k8s.io/api/core/v1"
	"testing"
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
	GetLoadBalancer(config,service)
}