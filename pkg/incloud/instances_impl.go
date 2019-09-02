//暂不实现

package incloud

import (
	"k8s.io/cloud-provider"
)

//var _ cloudprovider.Instances = &InCloud{}

// Instances returns an implementation of Instances for InCloud.
func (qc *InCloud) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

//func (qc *InCloud) newInstance(name string) *instance.Instance {
//	return nil
//}

//NodeAddresses returns the addresses of the specified instance.
//TODO(roberthbailey): This currently is only used in such a way that it
//returns the address of the calling instance. We should do a rename to
//make this clearer.
//func (qc *InCloud) NodeAddresses(_ context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
//	return qc.newInstance(string(name)).GetK8sAddress()
//}
//
//func (qc *InCloud) NodeAddressesByProviderID(_ context.Context, providerID string) ([]v1.NodeAddress, error) {
//	if providerID == "" {
//		err := fmt.Errorf("Call NodeAddresses with an empty ProviderID")
//		return nil, err
//	}
//	ins := qc.newInstance(providerID)
//	err := ins.LoadQcInstanceByID(providerID)
//	if err != nil {
//		klog.Errorf("Failed to load InCloud instance of %s", providerID)
//		return nil, err
//	}
//	return ins.GetK8sAddress()
//}
//
//func (qc *InCloud) InstanceID(_ context.Context, nodeName types.NodeName) (string, error) {
//	return instance.NodeNameToInstanceID(string(nodeName), qc.nodeInformer.Lister()), nil
//}
//
//func (qc *InCloud) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
//	ins := qc.newInstance(string(name))
//	err := ins.LoadQcInstance()
//	if err != nil {
//		klog.Errorf("Failed to load InCloud instance of %s", name)
//		return "", err
//	}
//	return *ins.Status.InstanceType, nil
//}
//
//func (qc *InCloud) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
//	ins := qc.newInstance(providerID)
//	err := ins.LoadQcInstanceByID(providerID)
//	if err != nil {
//		klog.Errorf("Failed to load InCloud instance of %s", providerID)
//		return "", err
//	}
//	return *ins.Status.InstanceType, nil
//}
//
//// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances.
//// The method is currently only used in gce.
//func (qc *InCloud) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
//	return errors.New("Unimplemented")
//}
//
//func (qc *InCloud) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
//	return types.NodeName(hostname), nil
//}
//
//func (qc *InCloud) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
//	ins := qc.newInstance(providerID)
//	err := ins.LoadQcInstanceByID(providerID)
//	if err != nil {
//		klog.Errorf("Failed to load InCloud instance of %s", providerID)
//		return false, err
//	}
//	if *ins.Status.Status == "terminated" || *ins.Status.Status == "ceased" {
//		return false, nil
//	}
//	return true, nil
//}
//
//func (qc *InCloud) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
//	ins := qc.newInstance(providerID)
//	err := ins.LoadQcInstanceByID(providerID)
//	if err != nil {
//		klog.Errorf("Failed to load InCloud instance of %s", providerID)
//		return false, err
//	}
//	if *ins.Status.Status == "stopped" || *ins.Status.Status == "suspended" {
//		return true, nil
//	}
//	return false, err
//}
