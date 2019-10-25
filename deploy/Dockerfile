# Copy the controller-manager into a thin image
FROM registry.icp.com:5000/library/curl:test
MAINTAINER zhang yong
WORKDIR /
COPY main manager
ENV PATH=$PATH:/
CMD ["/manager"]
