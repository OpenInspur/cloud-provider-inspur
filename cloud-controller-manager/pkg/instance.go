package pkg

import (
	"gitserver/kubernetes/inspur-cloud-controller-manager/cloud-controller-manager/pkg/common"
	"k8s.io/api/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
)

type Instance struct {
	Name string
	//instanceApi *qcservice.InstanceService
	nodeLister corev1lister.NodeLister
	isApp      bool
	//Status      *qcservice.Instance
}

type InstanceSpec struct {
}

//func NewInstance(api *qcservice.InstanceService, nodeLister corev1lister.NodeLister, name string, isApp bool) *Instance {
//	return &Instance{
//		Name:        name,
//		instanceApi: api,
//		nodeLister:  nodeLister,
//		isApp:       isApp,
//	}
//}

func (i *Instance) GetInstanceID() string {
	return ""
	//return GetNodeInstanceID(i.Name, i.nodeLister)
}

//func (i *Instance) LoadQcInstance() error {
//	id := i.GetInstanceID()
//	return i.LoadQcInstanceByID(id)
//}

//func (i *Instance) LoadQcInstanceByID(id string) error {
//	status := []*string{qcservice.String("pending"), qcservice.String("running"), qcservice.String("stopped")}
//	input := &qcservice.DescribeInstancesInput{
//		Instances: []*string{&id},
//		Status:    status,
//		Verbose:   qcservice.Int(1),
//	}
//	if i.isApp {
//		input.IsClusterNode = qcservice.Int(1)
//	}
//	output, err := i.instanceApi.DescribeInstances(input)
//	if err != nil {
//		return err
//	}
//	if len(output.InstanceSet) == 0 {
//		return cloudprovider.InstanceNotFound
//	}
//	i.Status = output.InstanceSet[0]
//	return nil
//}
//
func (i *Instance) GetK8sAddress() ([]v1.NodeAddress, error) {
	//if i.Status == nil {
	//	err := i.LoadQcInstance()
	//	if err != nil {
	//		klog.Errorf("error getting instance '%v'", i.Name)
	//		return nil, err
	//	}
	//}
	addrs := []v1.NodeAddress{}
	//for _, vxnet := range i.Status.VxNets {
	//	// vxnet.Role 1 main nic, 0 slave nic. skip slave nic for hostnic cni plugin
	//	if vxnet.PrivateIP != nil && *vxnet.PrivateIP != "" && *vxnet.Role == 1 {
	//		addrs = append(addrs, v1.NodeAddress{Type: v1.NodeInternalIP, Address: *vxnet.PrivateIP})
	//	}
	//}
	//
	//if i.Status.EIP != nil && i.Status.EIP.EIPAddr != nil && *i.Status.EIP.EIPAddr != "" {
	//	addrs = append(addrs, v1.NodeAddress{Type: v1.NodeExternalIP, Address: *i.Status.EIP.EIPAddr})
	//}
	//if len(addrs) == 0 {
	//	err := fmt.Errorf("The instance %s maybe broken because it has no ip", *i.Status.InstanceID)
	//	return nil, err
	//}
	return addrs, nil
}

// Make sure incloud instance hostname or override-hostname (if provided) is equal to InstanceId
// Recommended to use override-hostname
func GetNodeInstanceID(node *v1.Node) string {
	if instanceid, ok := node.GetAnnotations()[common.NodeAnnotationInstanceID]; ok {
		return instanceid
	}
	return node.Name
}
