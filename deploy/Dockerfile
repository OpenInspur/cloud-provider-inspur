# Copy the controller-manager into a thin image
FROM ubuntu
MAINTAINER zhang yong
WORKDIR /
COPY ./bin/ .
ENV PATH=$PATH:/
CMD ["/manager"]
