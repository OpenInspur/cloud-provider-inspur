# Copy the controller-manager into a thin image
FROM registry.icp.com:5000/service/devops/runtime/golang:alpine-3.10
MAINTAINER zhang yong
WORKDIR /
COPY main manager
ENV PATH=$PATH:/
CMD ["/manager"]
