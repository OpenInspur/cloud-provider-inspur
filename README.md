# Loadbalancers

Inspur Cloud Controller Manager runs service controller,
which is responsible for watching services of type ```LoadBalancer```
and creating Inspur loadbalancers to satisfy its requirements.
Here are some examples of how it's used.
也可作为测试人员测试步骤,
按照slb组文档[http://git.inspur.com/inspurcloud-api-doc/slb-api-doc/blob/master/3-api-details.md#1-create-loadbalancer]
所说:[网络类型：取值:network(公网)、innernet(内网，暂未上线)],可以先测External HTTP loadbalancer

**step1:**

以测试用户来创建一个负载均衡sld，需要在负载均衡产品页面创建，如产线为https://console1.cloud.inspur.com/slb/#/slb?region=cn-north-3

**step2:**

创建service,type为loadbalancer

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
it means 是否开启健康检查.


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

也可以在负载均衡页面查看
