# Cloud Provider 帮助文档（中文版）

## inspur Cloud Provider 介绍

**Cloud Provider**

提供作为树外云控制管理器的Cloud Provider程序接口。它允许Kubernetes集群利用Inspur云的基础设施服务。

**先决条件**

- 一个可用的ACS-kubernetes集群
- 用kubectl连接到kubernetes集群
- 创建nginx部署。下面的示例基于nginx部署

## 注意事项

- Cloud Controller Manager（简称CCM）会为`Type=LoadBalancer`类型的Service创建或配置浪潮云负载均衡（SLB），包含**SLB**、**监听**、**虚拟服务器组**等资源。
- 对于非LoadBalancer类型的service则不会为其配置负载均衡，这包含如下场景：当用户将`Type=LoadBalancer`的service变更为`Type!=LoadBalancer`时，CCM也会删除其原先为该Service创建的SLB.
- 自动刷新配置：CCM使用声明式API，会在一定条件下自动根据service的配置刷新浪潮云负载均衡配置，所有用户自行在SLB控制台上修改的配置均存在被覆盖的风险（使用已有SLB同时不覆盖监听的场景除外），因此不能在SLB控制台手动修改Kubernetes创建并维护的SLB的任何配置，否则有配置丢失的风险。
- 只支持为serivce指定一个已有的负载均衡。
- 指定已有SLB
  - 需要为Service设置`service.beta.kubernetes.io/inspur-load-balancer-slbid` annotation。
  - SLB配置：此时CCM会使用该SLB做为Service的SLB，并根据其他annotation配置SLB，并且自动的为SLB创建多个虚拟服务器组（当集群节点变化的时候，也会同步更新虚拟服务器组里面的节点）。
  - 转发规则配置：通过添加`loadbalancer.inspur.com/forward-rule`来配置转发规则，例如WR是加权轮循，RR是轮循。
  - 健康检查配置配置：是否配置监听取决于`loadbalancer.inspur.com/is-healthcheck`是否设置为true。 如果设置为false，那么CCM不会为SLB管理任何健康检查。如果设置为true，那么CCM会采用健康检查。
  - SLB的删除： 当Service删除的时候CCM不会删除用户通过id指定的已有SLB。
- 后端服务器更新
  - CCM会自动的为该Service对应的SLB刷新后端虚拟服务器组。当Service对应的后端Endpoint发生变化的时候或者集群节点变化的时候都会自动的更新SLB的后端Server。
  - 任何情况下CCM不会将Master节点作为SLB的后端。
## 如何使用

浪潮云控制器管理器运行服务控制器，负责监视loadbalancer类型的服务，并创建浪潮loadbalancer以满足其需求。
下面是一些如何使用它的例子。

**步骤1:**

要通过测试用户创建负载平衡器SLD，需要在负载平衡产品页上创建它。例如，生产线是 https://console1.cloud.inspur.com/slb/#/slb?region=cn-north-3

**步骤2:**

当您使用"type:load balancer"创建服务时，将创建一个Inspur负载平衡器。
下面的示例将创建一个nginx部署，并通过Inspur外部负载平衡器公开它.

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

```service.beta.kubernetes.io/inspur-load-balancer-slbid```
用于指示我们要使用的负载平衡器id。

```loadbalancer.inspur.com/forward-rule``` 
指示要使用的转发规则，例如WRR、RR。

```loadbalancer.inspur.com/is-healthcheck``` 默认值是false.
意思是是否开启健康检查。

将yaml模板保存到test/loadbalancers/external-http-nginx.yaml，然后使用“kubectl create-f test/loadbalancers/external http nginx.yaml”创建服务。
监视服务并通过以下命令等待```EXTERNAL-IP```。
这将是负载平衡器IP，您可以使用它连接到您的服务。

```bash
$ watch kubectl get service
NAME                 CLUSTER-IP     EXTERNAL-IP       PORT(S)        AGE
http-nginx-service   10.0.0.10      122.112.219.229   80:30000/TCP   5m
```

**step3:**

现在您可以通过设置的负载平衡器访问您的服务。

```bash
$ curl http://122.112.219.229
```


