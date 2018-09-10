# Kubeless on Azure Kubernetes Service

## 1. Introduction

This guide goes over the required steps for deploying Kubeless in Azure AKS (Azure Kubernetes Service). The steps in this guide require for you to install:

 - [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) (`az`): This CLI will be used to create the cluster in AKS.
 - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/): Used for installing Kubeless.

## 2. Creating an AKS cluster

In order to get Kubeless up and running on top of AKS of course you'll need an AKS cluster. Fortunately, Microsoft already did a great job documenting the entire process to accomplish that. You can reach out that documentation following [this link](https://docs.microsoft.com/en-us/azure/aks/kubernetes-walkthrough#create-aks-cluster).

### Important notes regarding the cluster creation itself

* In the same document the property `--generate-ssh-keys` was used to generate the required SSH keys to the cluster deployment. If you would like to create your own keys, please use `--ssh-key-value` passing the path to your SSH pub file.

## 3. Installing "Kubeless-Controller"

Assuming that the Kubernetes cluster is up and running on top of ACS, its time to install Kubeless. To accomplish, please, follow the steps described on Kubeless [Quick-Start Guide](/docs/quick-start).
