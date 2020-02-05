// Copyright 2019 inspur Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"k8s.io/client-go/informers"
	corev1informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/cloud-provider"
	"k8s.io/klog"
	"os"
	"strings"
)

const (
	ProviderName           = "incloud"
	DefaultCloudConfigPath = "/etc/kubernetes/cloud-config"
)

type Config struct {
	KeycloakUrl      string `gcfg:"keycloakUrl"`
	ClientSecret     string `gcfg:"client-secret"`
	RequestedSubject string `gcfg:"requested-subject"`
	TokenClientID    string `gcfg:"token-client-id"`
	SubnetID         string `gcfg:"subnet-id"`  //由于loadbalancer本身不在cloud controller manager创建，不需要
	SlbUrlPre        string `gcfg:"slbUrl-pre"` //cloud-config中配置slb url前缀；
	KeycloakToken    string `gcfg:"kktoken"`
}

var _ cloudprovider.Interface = &InCloud{}

// A single Kubernetes cluster can run in multiple zones,
// but only within the same region (and cloud provider).
type InCloud struct {
	zone            string
	clusterID       string
	nodeInformer    corev1informer.NodeInformer
	serviceInformer corev1informer.ServiceInformer

	LbUrlPre         string
	KeycloakToken    string
	RequestedSubject string
	TokenClientID    string
	ClientSecret     string
	KeycloakUrl      string
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

	return LoadCloudCfg()
}

func LoadCloudCfg() (Config, error) {
	fi, err := os.Open(DefaultCloudConfigPath)
	if err != nil {
		return Config{}, fmt.Errorf("load cloud config file: %v", err)
	}
	defer fi.Close()

	var slbUrlPre, requestedSubject, tokenClientID, clientSecret, keycloakUrl, keycloakToken string
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		arg := string(a)
		if arg != "" && arg[0:1] != "#" && strings.Index(arg, "=") > 0 {
			key := arg[0:strings.Index(arg, "=")]
			value := arg[strings.Index(arg, "=")+1 : len(arg)]
			switch key {
			case "slbUrl-pre":
				slbUrlPre = value
			case "client-secret":
				clientSecret = value
			case "requested-subject":
				requestedSubject = value
			case "token-client-id":
				tokenClientID = value
			case "keycloakUrl":
				keycloakUrl = value
			case "kktoken":
				keycloakToken = value
			default:
			}
		}
	}
	config := Config{keycloakUrl, clientSecret, requestedSubject,
		tokenClientID, "", slbUrlPre, keycloakToken}
	klog.Info(config)
	return config, nil
}

// newInCloud returns a new instance of InCloud cloud provider.
func newInCloud(config Config) (cloudprovider.Interface, error) {
	qc := InCloud{
		LbUrlPre:         config.SlbUrlPre,
		KeycloakToken:    config.KeycloakToken,
		RequestedSubject: config.RequestedSubject,
		TokenClientID:    config.TokenClientID,
		ClientSecret:     config.ClientSecret,
		KeycloakUrl:      config.KeycloakUrl,
	}

	klog.Infof("InCloud provider init done")
	b, _ := json.Marshal(&qc)
	klog.Infof("InCloud is ", string(b))
	return &qc, nil
}

func (ic *InCloud) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	clientset := clientBuilder.ClientOrDie("do-shared-informers")
	sharedInformer := informers.NewSharedInformerFactory(clientset, 0)
	nodeinformer := sharedInformer.Core().V1().Nodes()
	go nodeinformer.Informer().Run(stop)
	ic.nodeInformer = nodeinformer

	serviceInformer := sharedInformer.Core().V1().Services()
	go serviceInformer.Informer().Run(stop)
	ic.serviceInformer = serviceInformer
}

func (ic *InCloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (ic *InCloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (ic *InCloud) ProviderName() string {
	return ProviderName
}

// HasClusterID returns true if the cluster has a clusterID
func (ic *InCloud) HasClusterID() bool {
	return true
}
