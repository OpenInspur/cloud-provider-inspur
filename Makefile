# tab space is 4
# GitHub viewer defaults to 8, change with ?ts=4 in URL

# Vars describing project
NAME= cloud-provider-inspur
GIT_REPOSITORY= github.com/OpenInspur/cloud-provider-inspur
REGISTRY?= registry.inspurcloud.cn:5000/csf
TAG?=$(shell git describe --tags)

# Set defaults for needed vars in case version_info script did not set
# Revision set to number of commits ahead
VERSION?= 0.0
COMMITS?= 0
REVISION?= $(COMMITS)
BUILD_LABEL?= unknown_build
BUILD_DATE?= $(shell date -u +%Y%m%d.%H%M%S)
GIT_SHA1?= unknown_sha1

# default just build binary
default: go-build

# target for debugging / printing variables
print-%:
    @echo '$*=$($*)'

# perform go build on project
go-build: bin/cloud-provider-inspur

bin/cloud-provider-inspur:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -o bin/manager ./cmd/main.go

bin/.docker-images-build-timestamp: bin/cloud-provider-inspur Makefile Dockerfile
    docker build -q -t  $(REGISTRY)/$(NAME):$(TAG) -t dockerhub.inspur.cloud.com/ $(REGISTRY)/$(NAME):$(TAG) . > bin/.docker-images-build-timestamp

image: bin/cloud-provider-inspur Makefile Dockerfile
    docker build -t $(REGISTRY)/$(NAME):$(TAG)  .

publish: test go-build
    docker build -t $(REGISTRY)/$(NAME):$(TAG)  -f ./Dockerfile bin/
    docker push $(REGISTRY)/$(NAME):$(TAG)

clean:
    rm -rf bin/ && if -f bin/.docker-images-build-timestamp then docker rmi `cat bin/.docker-images-build-timestamp`

test: fmt vet
    go test -v -cover -mod=vendor ./cloud-controller-manager/pkg/...

fmt:
    go fmt ./cloud-controller-manager/pkg/... ./cmd/... ./test/pkg/...

vet:
    go vet ./cloud-controller-manager/pkg/... ./cmd/... ./test/pkg/...

.PHONY: default all go-build clean install-docker test

e2e:
    ./hack/e2e.sh