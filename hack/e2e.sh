#!/bin/bash

set -e


SKIP_BUILD=no
tag=`git rev-parse --short HEAD`
IMG=magicsong/cloud-manager:$tag
DEST=test/manager.yaml
TEST_NS=cloud-test-$tag
config_file=test/config/incloud.yaml
#build binary

function cleanup(){
    result=$?
    set +e

    if [ $result != "0" ]; then
        echo "Detect test failure, save logs"
        podname=`kubectl get pod -n $TEST_NS -l app=cloud-controller-manager -o jsonpath="{.items[0].metadata.name}"`
        kubectl logs $podname -n $TEST_NS > cloud_provider-inspur.log
    fi

    echo "Cleaning Namespace"
    kubectl delete ns $TEST_NS > /dev/null
    if [ $SKIP_BUILD == "no" ]; then
        docker  rmi $IMG
    fi
    exit $result
}
set -e
trap cleanup EXIT SIGINT SIGQUIT

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -s|--skip-build)
    SKIP_BUILD=yes
    shift # past argument
    ;;
    -n|--NAMESPACE)
    TEST_NS=$2
    shift # past argument
    shift # past value
    ;;
    -t|--tag)
    tag="$2"
    shift # past argument
    shift # past value
    ;;
    --default)
    DEFAULT=YES
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done

if [ $SKIP_BUILD == "no" ]; then
    echo "Building binary"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -o bin/manager ./cmd/main.go

    echo "Building docker image"
    docker build -t $IMG  -f Dockerfile bin/
    echo "Push images"
    docker push $IMG
    echo "Generating yaml"

fi
sed -e 's@image: .*@image: '"${IMG}"'@' -e 's/kube-system/'"$TEST_NS"'/g' docs/examples/kube-cloud-controller-manager.yaml > $DEST

kubectl create ns $TEST_NS
kubectl create configmap lbconfig --from-file=$config_file -n $TEST_NS
kubectl apply -f $DEST
export TEST_NS
## prepare eips
export TEST_EIP="eip-vmldumvv"
export TEST_EIP_ADDR="139.198.121.161"
export TEST_ZONE="ap2a"

go test -timeout 20m -v -mod=vendor ./test/pkg/e2e/
