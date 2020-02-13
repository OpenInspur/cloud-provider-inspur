| title                           | authors   | owning-sig         | reviewers    | approvers    | editor | creation-date | last-updated | status      |
| ------------------------------- | --------- | ------------------ | ------------ | ------------ | ------ | ------------- | ------------ | ----------- |
| Cloud Provider for Inspur Cloud | @huyuqing | sig-cloud-provider | @zhangyongn1 | @zhangyong01 | TBD    | 2020-02-11    | 2020-02-20   | provisional |

# Cloud Provider for Inspur Cloud

This is a KEP for adding `Cloud Provider for Inspur Cloud` into the Kubernetes ecosystem.

## Table of Contents

- [Summary](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-Inspur.md#summary)
- Motivation
  - [Goals](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#goals)
  - [Non-Goals](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#non-goals)
- Prerequisites
  - [Repository Requirements](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#repository-requirements)
  - [User Experience Reports](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#user-experience-reports)
  - [Testgrid Integration](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#testgrid-integration)
  - [CNCF Certified Kubernetes](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#cncf-certified-kubernetes)
  - [Documentation](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#documentation)
  - [Technical Leads are members of the Kubernetes Organization](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#technical-leads-are-members-of-the-kubernetes-organization)
- Proposal
  - [Repositories](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#repositories)
  - [Meetings](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#meetings)
  - [Others](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cloud-provider/providers/0020-cloud-provider-inspur.md#others)

## Summary

Inspur Cloud provides the Cloud Provider interface implementation as an out-of-tree cloud-controller-manager. It allows Kubernetes clusters to leverage the infrastructure services of Inspur Cloud . 

## Motivation

### Goals

Cloud Provider of Inspur Cloud implements interoperability between Kubernetes cluster and Inspur Cloud. In this project, we will dedicated in:

- Provide reliable, secure and optimized integration with Inspur Cloud for Kubernetes
- Help on the improvement for decoupling cloud provider specifics from Kubernetes implementation.

### Non-Goals

- Identify domain knowledge and work that can be contributed back to Kubernetes and related CNCF projects.

- Mentor CCE developers to contribute to CNCF projects.

- Focus on Kubernetes and CNCF related projects, the discussion of development issue for CCE  will not be included in the SIG.

## Prerequisites

1. The VPC network is supported in this project. The support for classic network or none ECS environment will be out-of-scope.
2. The existing create loadbalancer logic is in front, and CCM starts from the listener management interface (such as listener management interface) and the backend server management interface (add backendservers).
3. Kubernetes version v1.7 or higher

### Repository Requirements

The repository url which meets all the requirements is https://github.com/inspurcsg/cloud-provider-inspur

### User Experience Reports

1. The cloud-controller-manager Plug-in unit is created successfully and running normally.

![1581493267842](https://raw.githubusercontent.com/OpenInspur/cloud-provider-inspur/master/docs/media/1.png)

2. The deployment of nginx and its service is successful.

![1581493316451](https://raw.githubusercontent.com/OpenInspur/cloud-provider-inspur/master/docs/media/2.png)

3. Load balancing available.

![1581493330883](https://raw.githubusercontent.com/OpenInspur/cloud-provider-inspur/master/docs/media/3.png)

4. Update the member of the loadbalancer when the pod number changes (the node increases),the number of nginx copies is 2, which are all on the slave1 node. Query the member of the loadbalancer.

   ![1581493686508](https://raw.githubusercontent.com/OpenInspur/cloud-provider-inspur/master/docs/media/4.png)

5. Modify the current number of replicas to 4. On the slave1 / slave2 node, query the member of the   loadbalancer.

   ![](https://raw.githubusercontent.com/OpenInspur/cloud-provider-inspur/master/docs/media/5.png)

   ![](https://raw.githubusercontent.com/OpenInspur/cloud-provider-inspur/master/docs/media/6.png)

### Testgrid Integration

Inspur cloud provider is reporting conformance test results to TestGrid as per the [Reporting Conformance Test Results to Testgrid KEP](https://github.com/kubernetes/community/blob/master/keps/sig-cloud-provider/0018-testgrid-conformance-e2e.md). See [report](https://k8s-testgrid.appspot.com/conformance-alibaba-cloud-provider#Alibaba Cloud Provider, v1.10) for more details.

### CNCF Certified Kubernetes

Inspur cloud provider is accepted as part of the [Certified Kubernetes Conformance Program](https://github.com/cncf/k8s-conformance). For v1.14 See [inspur-iop-amd64](https://github.com/cncf/k8s-conformance/tree/master/v1.14/inspur-iop-amd64 ) [inspur-iop-arm64](https://github.com/cncf/k8s-conformance/tree/master/v1.14/inspur-iop-arm64)For v1.15 See [inspur-iop-amd64](https://github.com/cncf/k8s-conformance/tree/master/v1.15/inspur-iop-amd64) [inspur-iop-arm64](https://github.com/cncf/k8s-conformance/tree/master/v1.15/inspur-iop-arm64) For v1.16 See [inspur-iop-amd64](https://github.com/cncf/k8s-conformance/tree/master/v1.16/inspur-iop-amd64) [inspur-iop-arm64](https://github.com/cncf/k8s-conformance/tree/master/v1.16/inspur-iop-arm64) V1.17 has been proposed and is still in the process of authentication and Inspur has also submitted CNCF Certified of mipes64 architecture after v1.16.

### Documentation

Inspur CloudProvider provide users with multiple documentation on build & deploy & utilize CCM. Please refer tohttps://github.com/inspurcsg/cloud-provider-inspur/docs/ for more details.

### Technical Leads are members of the Kubernetes Organization

The Leads run operations and processes governing this subproject.

- @TimYin Special Tech Leader, Inspur Cloud. Kubernetes Member

## Proposal

Here we propose a repository from Kubernetes organization to host our cloud provider implementation. Cloud Provider of Inspur Cloud would be a subproject under Kubernetes community.

### Repositories

Cloud Provider of Inspur Cloud will need a repository under Kubernetes org named `kubernetes/cloud-provider-inspur to host any cloud specific code. The initial owners will be indicated in the initial OWNER files.

Additionally, SIG-cloud-provider take the ownership of the repo but Inspur Cloud should have the fully autonomy permission to operator on this subproject.

### Meetings

Cloud Provider meetings is expected to have biweekly. SIG Cloud Provider will provide zoom/youtube channels as required. We will have our first meeting after repo has been settled.

Recommended Meeting Time: Wednesdays at 20:00 PT (Pacific Time) (biweekly). [Convert to your timezone](http://www.thetimezoneconverter.com/?t=20:00&tz=PT (Pacific Time)).

- Meeting notes and Agenda.
- Meeting recordings.

### Others

NA at this moment.
