
Developers may want to build the image from scratch. We provide a simple command ```make image``` to accomplish this. 
Be advise that build it from source requires docker to be installed.
A valid tag is required to build your image. Tag with ```git tag v1.9.3```

## Build `cloud-controller-manager` image

```bash
# for example REGISTRY=registry.inspurcloud.cn/
# YExecute the following command in the directory where the dockerfile is located.This will build an docker image from binary 
$ docker build -t $(REGISTRY):$(port)/cloud-provider-inspur:$(TAG) .
# push to your specified registry.
$ docker push 
# Query the created image
$ docker images |grep cloud-provider-inspur
```

## Testing

### UnitTest

See [Testing UnitTest](https://github.com/kubernetes/cloud-provider-inspur/tree/master/docs/testing.md)
