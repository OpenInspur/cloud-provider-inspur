# inspur Cloud Provider

## Accessing services through server load balancer

You can use Inspur cloud load balancer to access services.
For details, please refer to the official document: [access the service through server load balancer] (http:// )

## background information

- CloudProvider would not deal with your LoadBalancer(which was provided by user) listener by default if your cloud-controller-manager version is great equal then v1.9.3. User need to config their listener by themselves or using `service.beta.kubernetes.io/inspur-cloud-loadbalancer-force-override-listeners: "true"` to force overwrite listeners.<br />
Using the following command to find the version of your cloud-controller-manager

```
root@master # kubectl get po -n kube-system -o yaml|grep image:|grep cloud-con|uniq
image: registry.icp.cn-....-controller-manager-amd64:v1.9.3
```

## How to create service with Type=LoadBalancer

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