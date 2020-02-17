
# Prerequisites
- Version: kubernetes version > 1.7.2 is required.
- CloudNetwork: Only inspur Cloud VPC network is supported.

## API requirements for SLB
1. CKE does not need to pay attention to SLB API and does not need to charge
2. The existing create loadbalancer logic is in front, and CKE starts from the listener management interface (such as listener management interface) and the backend server management interface (add backendservers)
3. Listener management interface and back-end server management interface are synchronous interfaces

# Deploy CloudProvider in inspur Cloud.

## Preparation

1. Modify kubelet configuration parameters of all nodes and add parameters. 
```
- --cloud-config=/etc/kubernetes/cloud-config
- --cloud-provider=external
```
You may also need to add other startup items.

2. Generate a configuration file in the Master node，```kubernetes-deploy/etc/kubernetes/cloud-config```Configuration needs to be added，```incloud_slbUrl_prefix```
Node needs annotation (implemented in kube-deploy)：```NodeAnnotationInstanceID = “node.beta.kubernetes.io/instance-id”```

3. Create LB service and annotate.
```
ServiceAnnotationInternalSlbId = "service.beta.kubernetes.io/inspur-internal-load-balancer-slbid"
ServiceAnnotationLBForwardRule = "loadbalancer.inspur.com/forward-rule"
ServiceAnnotationLBHealthCheck = "loadbalancer.inspur.com/is-healthcheck"
```

## cloud-config
```
slburl-pre=https://service.cloud.inspur.com/regionsvc-cn-north/slb/v1/slbs
slbId=slb-0000000898
```

## Try With Simple Example

### Create Service
```
kind: Service
apiVersion: v1
metadata:
  name: external-http-nginx-service
  annotations:
    service.beta.kubernetes.io/inspur-load-balancer-slbid:
    loadbalancer.inspur.com/forward-rule: "RR"
    loadbalancer.inspur.com/is-healthcheck: "1"
    loadbalancer.inspur.com/healthcheck-type: "tcp"
    loadbalancer.inspur.com/healthcheck-port: "0"
    loadbalancer.inspur.com/healthcheck-period: "30"
    loadbalancer.inspur.com/healthcheck-timeout: "1"
    loadbalancer.inspur.com/healthcheck-max: "1"
    loadbalancer.inspur.com/healthcheck-domain: ""
    loadbalancer.inspur.com/healthcheck-path: "/"
spec:
  selector:
    app: nginx
  type: LoadBalancer
  ports:
  - name: http
    port: 80
    targetPort: 80
```

### Deployment
Once `cloud-controller-manager` is up and running, run a external nginx deployment:
```bash
$ cat <<EOF >external-http-nginx.yaml
---
---
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
Put the yaml example above into external-http-nginx.yaml and execute the following command.
```
$ kubectl create -f external-http-nginx.yaml
```
