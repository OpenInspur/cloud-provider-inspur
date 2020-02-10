# Kubernetes Cloud Controller Manager for inspur Cloud

Thank you for visiting the cloud-provider-inspur repository!


`cloud-provider-inspur` is the external Kubernetes cloud controller manager implementation for inspur Cloud. Running `cloud-provider-inspur` allows you build your kubernetes clusters leverage on many cloud services on inspur Cloud. You can read more about Kubernetes cloud controller manager [here](https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/).

## Development

Build an image with the command

```bash
# for example REGISTRY=registry.inspurcloud.cn/
# YExecute the following command in the directory where the dockerfile is located.This will build an docker image from binary 
$ docker build -t $(REGISTRY):$(port)/cloud-provider-inspur:$(TAG) .
# push to your specified registry.
$ docker push 
# Query the created image
$ docker images |grep cloud-provider-inspur
```

## QuickStart

- [Getting-started](docs/getting-started.md)
- [Usage Guide](docs/cloud-controller-manager.md)


## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack channel](https://kubernetes.slack.com/messages/sig-cloud-provider)
- [Mailing list](https://groups.google.com/forum/#!forum/kubernetes-sig-cloud-provider)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

## Testing
See more info in page [Test](https://github.com/kubernetes/cloud-provider-inspur/tree/master/docs/testing.md)