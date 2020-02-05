package pkg

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog"
)

var _ cloudprovider.Zones = &InCloud{}

func (ic *InCloud) Zones() (cloudprovider.Zones, bool) {
	return ic, true
}
func (ic *InCloud) GetZone(ctx context.Context) (cloudprovider.Zone, error) {
	klog.Infof("GetZone() called, current zone is %v", ic.zone)

	return cloudprovider.Zone{Region: ic.zone}, nil
}

// GetZoneByNodeName implements Zones.GetZoneByNodeName
// This is particularly useful in external cloud providers where the kubelet
// does not initialize node data.
func (ic *InCloud) GetZoneByNodeName(ctx context.Context, nodeName types.NodeName) (cloudprovider.Zone, error) {
	klog.Infof("GetZoneByNodeName() called, current zone is %v, and return zone directly as temporary solution", ic.zone)
	return cloudprovider.Zone{Region: ic.zone}, nil
}

// GetZoneByProviderID implements Zones.GetZoneByProviderID
// This is particularly useful in external cloud providers where the kubelet
// does not initialize node data.
func (ic *InCloud) GetZoneByProviderID(ctx context.Context, providerID string) (cloudprovider.Zone, error) {
	klog.Infof("GetZoneByProviderID() called, current zone is %v, and return zone directly as temporary solution", ic.zone)
	return cloudprovider.Zone{Region: ic.zone}, nil
}
