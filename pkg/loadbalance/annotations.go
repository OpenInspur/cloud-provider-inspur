package loadbalance

// EIPStrategy is the type representing eip strategy
type EIPStrategy string

type EIPAllocateSource string

const (
	// ServiceAnnotationLoadBalancerEipIds is the annotation which specifies a list of eip ids.
	// The ids in list are separated by ',', e.g. "eip-j38f2h3h,eip-ornz2xq7". And this annotation should
	// NOT be used with ServiceAnnotationLoadBalancerVxnetId. Please make sure there is one and only one
	// of them being set
	ServiceAnnotationLoadBalancerEipIds = "service.beta.kubernetes.io/incloud-load-balancer-eip-ids"

	// ServiceAnnotationLoadBalancerType is the annotation used on the
	// service to indicate that we want a incloud loadBalancer type.
	// value "0" means the LB can max support 5000 concurrency connections, it's default type.
	// value "1" means the LB can max support 20000 concurrency connections.
	// value "2" means the LB can max support 40000 concurrency connections.
	// value "3" means the LB can max support 100000 concurrency connections.
	// value "4" means the LB can max support 200000 concurrency connections.
	// value "5" means the LB can max support 500000 concurrency connections.
	ServiceAnnotationLoadBalancerType = "service.beta.kubernetes.io/incloud-load-balancer-type"

	// ServiceAnnotationLoadBalancerEipStrategy is usd to specify EIP use strategy
	// reuse represent the EIP can be shared with other service which has no port conflict
	// exclusive is the default value, means every service has its own EIP
	ServiceAnnotationLoadBalancerEipStrategy = "service.beta.kubernetes.io/incloud-load-balancer-eip-strategy"
)
