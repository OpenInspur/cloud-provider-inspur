# Copy the controller-manager into a thin image
FROM ubuntu
MAINTAINER zhang yong
WORKDIR /
COPY . .
ENV PATH=$PATH:/
CMD ["/manager"]
