# inspur Cloud Provider

## inspur Cloud Provider introduction

**CloudProvider** 

Provides the Cloud Provider interface implementation as an out-of-tree cloud-controller-manager. It allows Kubernetes clusters to leverage the infrastructure services of Inspur Cloud . 

**pre-requirement**

- An available ACS kubernetes cluster。
- Connect to your kubernetes cluster with kubectl。
- Create an nginx deployment。 The example below is based on then nginx deployment。

## Matters needing attention

- Cloud Controller Manager（Abbreviation CCM）,Inspur cloud SLB will be configured for services of type `Type=LoadBalancer`, including **SLB**, **monitor**, **virtual server group** and other resources.
- For non loadbalancer type services, load balancer will not be configured, which includes the following scenarios: When the user changes the service of `Type=LoadBalancer` to `Type=!LoadBalancer`, CCM will also delete the SLB it originally created for the service.
- Auto refresh configuration: CCM will automatically refresh the Inspur cloud load balancer configuration according to the service configuration under certain conditions by using the declarative API. There is a risk that the configuration modified by all users on the SLB console will be overwritten (except for the scene where the existing SLB is used and the monitoring is not overwritten). Therefore, it is not allowed to manually modify any SLB created and maintained by kubernetes on the SLB console Configuration, otherwise there is a risk of configuration loss.
- Only one existing load balancing service can be specified for service.
- Specify existing SLB
  - Need to be set for service`service.beta.kubernetes.io/inspur-load-balancer-slbid` annotation。
  - SLB configuration: at this time, CCM will use this SLB as the SLB of the service, configure SLB according to other annotations, and automatically create multiple virtual server groups for SLB (when the cluster nodes change, the nodes in the virtual server group will also be updated synchronously).
  - Forwarding rule configuration: configure the forwarding rule by adding `loadbalancer.inspur.com/forward-rule`. For example, WR is weighted round robin and RR is round robin.
  - Health check configuration configuration: whether to configure listening depends on whether `loadbalancer.inspur.com/is-healthcheck` is set to true. If set to false, CCM does not manage any health checks for SLB.如果设置为true，那么CCM会采用健康检查。
  - SLB deletion: when the service is deleted, CCM will not delete the existing SLB specified by user ID.
- Backend server update
  - CCM will automatically refresh the backend virtual server group for the SLB corresponding to the service. When the backend endpoint corresponding to the service changes or the cluster node changes, the backend server of SLB will be updated automatically.
  - In any case, CCM will not use the master node as the back end of SLB.

## How to used 

Inspur Cloud Controller Manager runs service controller,which is responsible for watching services of type LoadBalancerand creating Inspur loadbalancers to satisfy its requirements.
Here are some examples of how it's used.

**step1:**

To create a load balancer SLD by testing users, you need to create it on the load balancing product page. For example, the production line is https://console1.cloud.inspur.com/slb/#/slb?region=cn-north-3

**step2:**

When you create a service with ```type: LoadBalancer```, an Inspur load balancer will be created.
The example below will create a nginx deployment and expose it via an Inspur External load balancer.

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
    service.beta.kubernetes.io/inspur-load-balancer-slbid: #Fill in slbid created by step1 here
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

Save the yaml template to test/loadbalancers/external-http-nginx.yaml ， and then use `kubectl create -f test/loadbalancers/external-http-nginx.yaml` to create your service.
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