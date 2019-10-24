# Copy the controller-manager into a thin image
FROM registry.icp.com:5000/csf/busybox-curl
MAINTAINER zhang yong
WORKDIR /
COPY main manager
ENV PATH=$PATH:/
CMD ["/manager"]
