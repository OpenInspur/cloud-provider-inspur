# Copy the controller-manager into a thin image
FROM ubuntu
MAINTAINER zhang yong
WORKDIR /
COPY main manager
ENV PATH=$PATH:/
CMD ["/manager"]
