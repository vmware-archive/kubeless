# Deploying Kubeless to Google Kubernetes Engine (GKE)

This guide goes over the required steps for deploying Kubeless in GKE. There are a few pain points that you need to know in order to successfully deploy Kubeless in a GKE environment. First your google cloud account should have enough privileges to create and manage clusters. You can login to your account using the `gcloud` CLI tool:

```console
$ gcloud auth login
Go to the following link in your browser:

    https://accounts.google.com/o/oauth2/auth?redirect_uri=...

Enter verification code: ...
You are now logged in as [your@mail.com].
Your current project is [your-project].  You can change this setting by running:
  $ gcloud config set project PROJECT_ID
```

You can also follow the initialization process executing `gcloud init`.

## Creating a cluster

Once you are logged in, you can create the cluster:

```console
gcloud container clusters create \
  --cluster-version=1.8.10-gke.0 \
  my-cluster \
  --num-nodes 5
```

At the moment of writing this document, the CI/CD system is testing Kubeless against GKE 1.8 so that's the one we are specifying as the desired version. You can check the current version tested in [the Travis file](../.travis.yml).

The default number of nodes is 3. That default number is enough for small deployments but it is recommended to use at least 5 or 7 nodes so you don't run out of resources after deploying a few functions.

After a few minutes you should be able to see your cluster running:

```console
$ gcloud container clusters list
NAME        ...          STATUS
my-cluster  ...          RUNNING
```

## Creating the admin clusterrolebinding

For deploying Kubeless in your cluster, your user should have enough permissions for creating cluster roles and cluster role bindings. For doing so you need to give your current GKE account admin privileges in the new cluster. This is not being done by default so you need to do it manually:

```console
kubectl create clusterrolebinding kubeless-cluster-admin --clusterrole=cluster-admin --user=<your-gke-user>
```

The above command may fail with:

```console
Error from server (Forbidden): User "your-gke-user" cannot create
clusterrolebindings.rbac.authorization.k8s.io at the cluster scope
```

This error is shown since your account doesn't have privileges to create `clusterrolebindings` (even if you are able to create clusters). If that is the case you can still perform the above operation using the default `admin` user. You can retrieve the admin password executing:

```console
gcloud container clusters describe my-cluster --zone <my-cluster-zone>
```

Once you have the admin password you can retry the command above:

```console
kubectl --username=admin --password=<admin_password> \
  create clusterrolebinding kubeless-cluster-admin \
  --clusterrole=cluster-admin --user=<your-gke-user>
```

## Deploying Kubeless

After that your are finally able to deploy Kubeless. Get the latest release from the [release page](https://github.com/kubeless/kubeless/releases) and deploy the RBAC version of the Kubeless manifest.

## Kubeless on GKE 1.8.x with Alpha features

On GKE 1.8.x, when you have finished the above steps, there is still one step required to make the Kafka/Zookeeper PVC bounded if you enable alpha features when creating your cluster. Checking PVC you will see they are pending:

```
kubectl get pvc -n kubeless
NAME              STATUS    VOLUME    CAPACITY   ACCESSMODES   STORAGECLASS   AGE
datadir-kafka-0   Pending                                                     2m
zookeeper-zoo-0   Pending                                                     2m
```

Because there are no correlative PV available, you have to create them. On GKE, you might want to go with [GKE Persistent Disk](https://kubernetes.io/docs/concepts/storage/volumes/#gcepersistentdisk). First, create two PD with this command:

```console
gcloud compute disks create --size=1GB --zone=<your_GKE_zone> kubeless-kafka
gcloud compute disks create --size=1GB --zone=<your_GKE_zone> kubeless-zookeeper
```

Then create Kafka and Zookeeper PV:

```console
kubectl create -f docs/misc/kafka-pv.yaml
kubectl create -f docs/misc/zookeeper-pv.yaml
```

Once both PV are created, the PVC will be bounded shortly and you will see Kafka and Zookeeper running:

```console
kubectl get pod -n kubeless
NAME                                   READY     STATUS    RESTARTS   AGE
kafka-0                                1/1       Running   1          30m
kubeless-controller-659755588f-bwch6   1/1       Running   0          30m
zoo-0                                  1/1       Running   0          30m
```
