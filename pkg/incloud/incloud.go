// Copyright 2019 inspur Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package incloud

import (
	"fmt"
	"io"

	"gopkg.in/gcfg.v1"
	"k8s.io/client-go/informers"
	corev1informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/cloud-provider"
	"k8s.io/klog"
)

const (
	ProviderName = "incloud"
)

type LoadBalancerOpts struct {
	SubnetID string `gcfg:"subnet-id"` // overrides autodetection.
}

type Config struct {
	Global struct {
		AuthURL    string `gcfg:"auth-url"`
		Username   string
		UserID     string `gcfg:"user-id"`
		Password   string
		TenantID   string `gcfg:"tenant-id"`
		TenantName string `gcfg:"tenant-name"`
		TrustID    string `gcfg:"trust-id"`
		DomainID   string `gcfg:"domain-id"`
		DomainName string `gcfg:"domain-name"`

		ClientId    string `gcfg:"client-id"`
		KeycloakUrl string `gcfg:"keycloakUrl"`
		ProjectName string `gcfg:"projectName"`
		InstanceID  string `gcfg:"instanceID"`
		Port        string `gcfg:"port"`

		Region string
		CAFile string `gcfg:"ca-file"`

		EbsOpenApiUrl    string `gcfg:"ebs-openapi-url"`
		EbsServiceUrl    string `gcfg:"ebs-service-url"`
		ClientSecret     string `gcfg:"client-secret"`
		RequestedSubject string `gcfg:"requested-subject"`
		TokenClientID    string `gcfg:"token-client-id"`

		MasterAuthURL  string `gcfg:"master-auth-url"`
		OpenApiVersion string `gcfg:"openapi-version"`
	}
	LoadBalancer LoadBalancerOpts
}

var _ cloudprovider.Interface = &InCloud{}

// A single Kubernetes cluster can run in multiple zones,
// but only within the same region (and cloud provider).
type InCloud struct {
	//TODO
	zone            string
	clusterID       string
	nodeInformer    corev1informer.NodeInformer
	serviceInformer corev1informer.ServiceInformer
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := readConfig(config)
		if err != nil {
			return nil, err
		}
		return newInCloud(cfg)
	})
}

func readConfig(config io.Reader) (Config, error) {
	if config == nil {
		err := fmt.Errorf("no incloud provider config file given")
		return Config{}, err
	}

	var cfg Config
	err := gcfg.ReadInto(&cfg, config)
	return cfg, err
}

// newInCloud returns a new instance of InCloud cloud provider.
func newInCloud(config Config) (cloudprovider.Interface, error) {
	//TODO
	qc := InCloud{}

	klog.V(1).Infof("InCloud provider init done")
	return &qc, nil
}

func (qc *InCloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	clientset := clientBuilder.ClientOrDie("do-shared-informers")
	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)
	nodeinformer := sharedInformer.Core().V1().Nodes()
	go nodeinformer.Informer().Run(stop)
	qc.nodeInformer = nodeinformer

	serviceInformer := sharedInformer.Core().V1().Services()
	go serviceInformer.Informer().Run(stop)
	qc.serviceInformer = serviceInformer
}

func (qc *InCloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (qc *InCloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (qc *InCloud) ProviderName() string {
	return ProviderName
}

// HasClusterID returns true if the cluster has a clusterID
func (qc *InCloud) HasClusterID() bool {
	return true
}
