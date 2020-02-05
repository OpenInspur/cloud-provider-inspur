# Cloud Provider 帮助文档（中文版）

# 通过负载均衡（Server Load Balancer）访问服务

您可以使用浪潮云负载均衡来访问服务。  
详细信息请参考官方文档：[通过负载均衡（Server Load Balancer）访问服务](http://)

## 背景信息

如果您的集群的cloud-controller-manager版本大于等于v1.9.3，对于指定已有SLB的时候，系统默认不再为该SLB处理监听，用户需要手动配置该SLB的监听规则。

执行以下命令，可查看cloud-controller-manager的版本。

```
root@master # kubectl get po -n kube-system -o yaml|grep image:|grep cloud-con|uniq

image: registry-vpc.cn-hangzhou.inspur.com/acs/cloud-controller-manager-amd64:v1.9.3
```

## 注意事项

- Cloud Controller Manager（简称CCM）会为`Type=LoadBalancer`类型的Service创建或配置浪潮云负载均衡（SLB），包含**SLB**、**监听**、**虚拟服务器组**等资源。
- 对于非LoadBalancer类型的service则不会为其配置负载均衡，这包含如下场景：当用户将`Type=LoadBalancer`的service变更为`Type!=LoadBalancer`时，CCM也会删除其原先为该Service创建的SLB（用户通过`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id annotation`指定的已有SLB除外）。
- 自动刷新配置：CCM使用声明式API，会在一定条件下自动根据service的配置刷新浪潮云负载均衡配置，所有用户自行在SLB控制台上修改的配置均存在被覆盖的风险（使用已有SLB同时不覆盖监听的场景除外），因此不能在SLB控制台手动修改Kubernetes创建并维护的SLB的任何配置，否则有配置丢失的风险。
- 同时支持为serivce指定一个已有的负载均衡，或者让CCM自行创建新的负载均衡。但两种方式在SLB的管理方面存在一些差异：指定已有SLB
  - 需要为Service设置`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id` annotation。
  - SLB配置：此时CCM会使用该SLB做为Service的SLB，并根据其他annotation配置SLB，并且自动的为SLB创建多个虚拟服务器组（当集群节点变化的时候，也会同步更新虚拟服务器组里面的节点）。
  - 监听配置：是否配置监听取决于`service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners: `是否设置为true。 如果设置为false，那么CCM不会为SLB管理任何监听配置。如果设置为true，那么CCM会尝试为SLB更新监听，此时CCM会根据监听名称判断SLB上的监听是否为k8s维护的监听（名字的格式为k8s/Port/ServiceName/Namespace/ClusterID），若Service声明的监听与用户自己管理的监听端口冲突，那么CCM会报错。
  - SLB的删除： 当Service删除的时候CCM不会删除用户通过id指定的已有SLB。
- CCM管理的SLB  
  - CCM会根据service的配置自动的创建配置**SLB**、**监听**、**虚拟服务器组**等资源，所有资源归CCM管理，因此用户不得手动在SLB控制台更改以上资源的配置，否则CCM在下次Reconcile的时候将配置刷回service所声明的配置，造成非用户预期的结果。
  - SLB的删除：当Service删除的时候CCM会删除该SLB。
- 后端服务器更新
  - CCM会自动的为该Service对应的SLB刷新后端虚拟服务器组。当Service对应的后端Endpoint发生变化的时候或者集群节点变化的时候都会自动的更新SLB的后端Server。
  - `spec.ExternalTraffic = Cluster`模式的Service，CCM默认会将所有节点挂载到SLB的后端（使用BackendLabel标签配置后端的除外）。由于SLB限制了每个ECS上能够attach的SLB的个数（quota），因此这种方式会快速的消耗该quota,当quota耗尽后，会造成Service Reconcile失败。解决的办法，可以使用Local模式的Service。
  - `spec.ExternalTraffic = Local`模式的Service，CCM默认只会讲Service对应的Pod所在的节点加入到SLB后端。这会明显降低quota的消耗速度。同时支持四层源IP保留。
  - 任何情况下CCM不会将Master节点作为SLB的后端。
  - CCM默认不会从SLB后端移除被kubectl drain/cordon的节点。如需移除节点，请设置service.beta.kubernetes.io/alibaba-cloud-loadbalancer-remove-unscheduled-backend为on。
  >> 说明 
     如果是v1.9.3.164-g2105d2e-aliyun之前的版本，CCM默认会从SLB后端移除被kubectl drain/cordon的节点。
  
## 操作步骤

浪潮云控制器管理器运行服务控制器，它负责监视“LoadBalancer”类型的服务”并创建浪潮负载均衡器以满足其需求。

**步骤1:**

要通过测试用户创建负载平衡器SLD，需要在负载平衡产品页上创建它。例如，生产线是 https://console1.cloud.inspur.com/slb/#/slb?region=cn-north-3

**步骤2:**

创建service,type是loadbalancer

_**External HTTP loadbalancer**_

当您使用"type:load balancer"创建服务时，将创建一个Inspur负载平衡器。

下面的示例将创建一个nginx部署，并通过Inspur外部负载平衡器公开它.

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

```service.beta.kubernetes.io/inspur-load-balancer-slbid```
用于指示我们要使用的负载平衡器id。

```loadbalancer.inspur.com/forward-rule``` 
指示要使用的转发规则，例如WRR、RR。

```loadbalancer.inspur.com/is-healthcheck``` 默认值是false.
意思是是否开启健康检查。

```bash
$ kubectl create -f examples/loadbalancers/external-http-nginx.yaml
```

监视服务并通过以下命令等待“EXTERNAL-IP”。

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


