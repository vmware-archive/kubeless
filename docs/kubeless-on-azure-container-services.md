# Kubeless on Azure Container Services

## 1. Introduction

Azure Container Services (ACS) is a built-in service offered by Microsoft Azure platform to natively support complex container-based environments. By design, ACS supports Kubernetes as the main Docker conrainers orchestrator in its context. To know more about ACS, please, follow [this link](https://docs.microsoft.com/en-us/azure/container-service/kubernetes/container-service-intro-kubernetes).

Kubeless provides a native way to execute serverless functions on top of Kubernetes cluster so, ACS could become an ideal environment to run Kubeless on top of it.

This tutorial describes the entire process to install and execute Kubeless on top of Azure Container Services using Kubernetes.

## 2. Creating an ACS Kubernetes cluster

In order to get Kubeless up and running on top of ACS of course you'll need an ACS Kubernetes cluster. Fortunately, Microsoft already did a great job documenting the entire process to accomplish that. You can reach out that documentation following [this link](https://docs.microsoft.com/en-us/azure/container-service/kubernetes/container-service-tutorial-kubernetes-deploy-cluster).

### Important notes regarding the cluster creation itself

* The Microsoft's tutorial mentioned before creates a new cluster using Azure CLI 2.0. In order to get the things done in that way, you'll need to install this CLI. [This link](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) describes the process to do that in your environment.

* In the same document the property `--generate-ssh-keys` was used to generate the required SSH keys to the cluster deployment. If you would like to create your own keys, please use `--ssh-key-value` passing the path to your SSH pub file.

* ACS deploys at least one virtual machine to act as "master" and at least one virtual machine to act as "agent". If you'd like to access the master (where kubectl, docker and another stuff resides) for some reason, the usage of your own SSH keys is strongly recommended.

## 3. Installing "kubeless-controller"

