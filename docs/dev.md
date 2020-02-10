
Developers may want to build the image from scratch. We provide a simple command ```make image``` to accomplish this. 
Be advise that build it from source requires docker to be installed.
A valid tag is required to build your image. Tag with ```git tag v1.1.0```

## Build `cloud-controller-manager` image

```bash
# for example REGISTRY=registry.inspurcloud.cn:5000/csf/
# This will build an docker image from binary and push to your specified registry. 
$ make publish
# You can also use `make image` if you don't want push this image to your registry
# Query the created image
$ docker images |grep cloud-provider-inspur
```

## Testing

### UnitTest

See [Testing UnitTest](https://github.com/kubernetes/cloud-provider-inspur/tree/master/docs/testing.md)
