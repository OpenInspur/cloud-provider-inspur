package common

const (
	/*LoadBalancer
	 */

	//service定义时使用
	ServiceAnnotationLBInternal = "service.beta.kubernetes.io/inspur-internal-load-balancer"
	//Listener forwardRule
	ServiceAnnotationLBForwardRule = "loadbalancer.inspur.com/forward-rule"
	//Listener isHealthCheck
	ServiceAnnotationLBHealthCheck = "loadbalancer.inspur.com/is-healthcheck"

	/*Instances
	 */

	NodeAnnotationInstanceID = "node.beta.kubernetes.io/instance-id"
)
