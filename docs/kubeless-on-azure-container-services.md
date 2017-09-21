# Kubeless on Azure Container Services

## 1. Introduction

Azure Container Services (ACS) is a built-in service offered by Microsoft Azure platform to natively support complex container-based environments. By design, ACS supports Kubernetes as the main Docker containers orchestrator in its context. To know more about ACS, please, follow [this link](https://docs.microsoft.com/en-us/azure/container-service/kubernetes/container-service-intro-kubernetes).

Kubeless provides a native way to execute serverless functions on top of Kubernetes cluster so, ACS could become an ideal environment to run Kubeless on top of it.

This tutorial describes the entire process to install and execute Kubeless on top of Azure Container Services using Kubernetes.

## 2. Creating an ACS Kubernetes cluster

In order to get Kubeless up and running on top of ACS of course you'll need an ACS Kubernetes cluster. Fortunately, Microsoft already did a great job documenting the entire process to accomplish that. You can reach out that documentation following [this link](https://docs.microsoft.com/en-us/azure/container-service/kubernetes/container-service-tutorial-kubernetes-deploy-cluster).

### Important notes regarding the cluster creation itself

* The Microsoft's tutorial mentioned before creates a new cluster using Azure CLI 2.0. In order to get the things done in that way, you'll need to install this CLI. [This link](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) describes the process to do that in your environment.

* In the same document the property `--generate-ssh-keys` was used to generate the required SSH keys to the cluster deployment. If you would like to create your own keys, please use `--ssh-key-value` passing the path to your SSH pub file.

* ACS deploys at least one virtual machine to act as "master" and at least one virtual machine to act as "agent". If you'd like to access the master (where kubectl, docker and another stuff resides) for some reason, the usage of your own SSH keys is strongly recommended.

## 3. Installing "Kubeless-Controller"

Assuming that the Kubernetes cluster is up and running on top of ACS, its time to install kubeless-controller on master VM. To accomplish, please, follow the steps described on Kubeless documentation, available [here](https://github.com/kubeless/kubeless#installation).

## 4. Installing "Kubeless Client CLI"

Kubeless-Controller is the agent in charge to manage the requests from Kubernetes cluster side. The same way, there's an agent in charge to manage the occurrences from the client side (for instance: register a new function to be executed, call an existing function, etc.). This way, you'll need to install de Kubeless CLI.

Currently there's two different ways to accomplish that: manually (generating binaries manually) or in a automatized way (using a pre-compiled package provided by Bitnami).

### Option 1 - Automated by Bitnami

The is pretty simple. Please, do consider the snippets below.

For Linux environments:

``` 
# Step 1 - Download the package
curl -L https://github.com/kubeless/kubeless/releases/download/0.0.20/kubeless_linux-amd64.zip > kubeless.zip

# Step 2 - Unzip the package
unzip kubeless.zip

# Step 3 - Moving files to local path
sudo cp bundles/kubeless_linux-amd64/kubeless /usr/local/bin/
``` 

For OSX environments:

```
# Step 1 - Download the package
curl -L https://github.com/kubeless/kubeless/releases/download/0.0.20/kubeless_darwin-amd64.zip > kubeless.zip

# Step 2 - Unzip the package
unzip kubeless.zip

# Step 3 - Moving files to local path
sudo cp bundles/kubeless_darwin-amd64/kubeless /usr/local/bin/
TIP: For detailed installation instructions, visit the Kubeless releases page.
```

To verify if the installation was successful, please type `kubeless` in your terminal. The result should seems like the showed below.

```
Serverless framework for Kubernetes

Usage:
  kubeless [command]

Available Commands:
  function    function specific operations
  ingress     manage route to function on Kubeless
  topic       manage message topics in Kubeless
  version     Print the version of Kubeless

Flags:
  -h, --help   help for kubeless

Use "kubeless [command] --help" for more information about a command.
```
### Option 2 - Manually (do it yourself)

Anothe possibility is compiling and building kubeless CLI manually. If you would like to follow this path, please, follow the steps below.

**Step 1 - Install and configure Go**

Kubeless was written in Go lang. It means that if you would like make your own binaries, you'll need to have your environment appropriately settled. You can found an [step-by-step tutorial](https://golang.org/doc/install) that can drive you towards that. 

**Step 2 - Follow Dev Guide**

The next steps to build the kubeless CLI client were very well described on Kubeless official documentation under "dev-guide.md" file. It's available [here](https://github.com/kubeless/kubeless/blob/master/docs/dev-guide.md).

## 5. Usage and initial tests

Please follow Kubeless oficial documentation, available here: https://github.com/kubeless/kubeless#usage.