# Testing

### UnitTest

Inspur Cloud Controller use mocked cloud SDK to implement code unit test.

```go
// Test Http configuration.
func TestEnsureLoadBalancerVswitchID(t *testing.T) {

	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "https-service",
				UID:  types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerVswitch:     VSWITCH_ID,
					ServiceAnnotationLoadBalancerAddressType: string(slb.IntranetAddressType),
				},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
		nil,
		nil,
	)

	f.RunDefault(t, "Create Loadbalancer With VswitchID")
}
```

Here is an example of self defined test point.
```go
func TestEnsureLoadbalancerDeleted(t *testing.T) {
	prid := nodeid(string(REGION), INSTANCEID)
	f := NewDefaultFrameWork(
		// initial service based on your definition
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "https-service",
				UID:         types.UID(serviceUIDNoneExist),
				Annotations: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{Port: listenPort1, TargetPort: targetPort1, Protocol: v1.ProtocolTCP, NodePort: nodePort1},
				},
				Type:            v1.ServiceTypeLoadBalancer,
				SessionAffinity: v1.ServiceAffinityNone,
			},
		},
		// initial node based on your definition.
		// backend of the created loadbalancer
		[]*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: prid},
				Spec: v1.NodeSpec{
					ProviderID: prid,
				},
			},
		},
		nil,
		nil,
	)

	f.Run(
		t,
		"Delete Loadbalancer", "ecs",
		func() {
			_, err := f.Cloud.EnsureLoadBalancer(CLUSTER_ID, f.SVC, f.Nodes)
			if err != nil {
				t.Fatalf("delete loadbalancer error: create %s", err.Error())
			}
			err = f.Cloud.EnsureLoadBalancerDeleted(CLUSTER_ID, f.SVC)
			if err != nil {
				t.Fatalf("ensure loadbalancer delete error, %s", err.Error())
			}
			exist, _, err := f.LoadBalancer().findLoadBalancer(f.SVC)
			if err != nil || exist {
				t.Fatalf("Delete LoadBalancer error: %v, %t", err, exist)
			}
		},
	)
}
```

Faked SDK made unit test easier.

Use ```make test``` to run unit test.

### Integration Test

Inspur Cloud Controller Manager integration test is expected to follow the kubernetes testgrid rule addressed in https://github.com/kubernetes/community/pull/2224#issuecomment-395410751 for consistency.

### System Test

Inspur Cloud Controller Manager runs service controller,
which is responsible for watching services of type ```LoadBalancer```
and creating Inspur loadbalancers to satisfy its requirements.
[http://git.inspur.com/inspurcloud-api-doc/slb-api-doc/blob/master/3-api-details.md#1-create-loadbalancer]

**step1:**

To create a load balancer SLD by testing users, you need to create it on the load balancing product page. For example, the production line is https://console1.cloud.inspur.com/slb/#/slb?region=cn-north-3

**step2:**

create service,type is loadbalancer

_**External HTTP loadbalancer**_

When you create a service with ```type: LoadBalancer```, an Inspur load balancer will be created.
The example below will create a nginx deployment and expose it via an Inspur External load balancer.

_**yaml**_



```
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: external-http-nginx-deployment
  annotations:
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: registry.inspurcloud.cn/library/iop/nginx:1.17
        ports:
        - containerPort: 80
```

```
kind: Service
apiVersion: v1
metadata:
  name: external-http-nginx-service
  annotations:
    service.beta.kubernetes.io/inspur-load-balancer-slbid: #这里填写step1创建的slbid
    loadbalancer.inspur.com/forward-rule: "RR"
    loadbalancer.inspur.com/is-healthcheck: "0"
spec:
  selector:
    app: nginx
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: 80
```

---

The ```service.beta.kubernetes.io/inspur-load-balancer-slbid``` annotation
is used on the service to indicate the loadbalancer id we want to use.

The ```loadbalancer.inspur.com/forward-rule``` annotation
indicates which forwardRule we want to use,such as WRR,RR 

The ```loadbalancer.inspur.com/is-healthcheck``` default is false.
it means Whether to turn on health check.


```bash
$ kubectl create -f examples/loadbalancers/external-http-nginx.yaml
```

Watch the service and await an ```EXTERNAL-IP``` by the following command.
This will be the load balancer IP which you can use to connect to your service.

```bash
$ watch kubectl get service
NAME                 CLUSTER-IP     EXTERNAL-IP       PORT(S)        AGE
http-nginx-service   10.0.0.10      122.112.219.229   80:30000/TCP   5m
```

**step3:**

You can now access your service via the provisioned load balancer.

```bash
$ curl http://122.112.219.229
```