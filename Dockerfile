# Copy the controller-manager into a thin image
FROM registry.icp.com:5000/library/os/inspur-alpine-3.10:5.0.0

MAINTAINER zhang yong

WORKDIR /

COPY main manager

ENV PATH=$PATH:/

CMD ["/manager"]
