package common

const (
	/*LoadBalancer
	 */

	//service定义时使用
	ServiceAnnotationInternalSlbId = "service.beta.kubernetes.io/inspur-load-balancer-slbid"
	//Listener forwardRule
	ServiceAnnotationLBForwardRule = "loadbalancer.inspur.com/forward-rule"
	//Listener isHealthCheck
	ServiceAnnotationLBHealthCheck = "loadbalancer.inspur.com/is-healthcheck"
	//Listener typeHealthCheck
	ServiceAnnotationLBtypeHealthCheck = "loadbalancer.inspur.com/healthcheck-type"
	//Listener portHealthCheck
	ServiceAnnotationLBportHealthCheck = "loadbalancer.inspur.com/healthcheck-port"
	//Listener periodHealthCheck
	ServiceAnnotationLBperiodHealthCheck = "loadbalancer.inspur.com/healthcheck-period"
	//Listener timeoutHealthCheck
	ServiceAnnotationLBtimeoutHealthCheck = "loadbalancer.inspur.com/healthcheck-timeout"
	//Listener maxHealthCheck
	ServiceAnnotationLBmaxHealthCheck = "loadbalancer.inspur.com/healthcheck-max"
	//Listener domainHealthCheck
	ServiceAnnotationLBdomainHealthCheck = "loadbalancer.inspur.com/healthcheck-domain"
	//Listener pathHealthCheck
	ServiceAnnotationLBpathHealthCheck = "loadbalancer.inspur.com/healthcheck-path"

	/*Instances
	 */

	NodeAnnotationInstanceID = "node.beta.kubernetes.io/instance-id"
)
