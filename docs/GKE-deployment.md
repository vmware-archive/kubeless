# Deploying Kubeless to Google Kubernetes Engine (GKE)

This guide goes over the required steps for deploying Kubeless in GKE. There are a few pain points that you need to know in order to successfully deploy Kubeless in a GKE environment. First your google cloud account should have privileges enough to create and manage clusters. You can login into your account using the `gcloud` CLI tool:

```
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

```
gcloud container clusters create \
  --cluster-version=1.7.8-gke.0 \
  my-cluster \
  --num-nodes 5 \
  --no-enable-legacy-authorization \
  --enable-kubernetes-alpha
```

Right now the version that is being tested in the Kubeless CI/CD system is GKE 1.7 so that's the one we are specifying as the desired version.

The default number of nodes is 3. That default number is enough for small deployments but it is recommened to use at least 4 or 5 nodes so you don't run out of resources after deploying a few functions.

In the case of GKE 1.7 the legacy authorization system sets every RBAC account by default as admin so it is not really restricting the access. For disabling it we set the flag `--no-enable-legacy-authorization`. 

Finally if you want to use Alpha features like autoscaling or scheduled actions in GKE 1.7 you need to set the flag `--enable-kubernetes-alpha`. Note that this will cause that the new cluster is marked to be deleted after 30 days. For non-testing environments this flag should be ommited.

After a few minutes you should be able to see your cluster running:
```
$ gcloud container clusters list
NAME           ZONE        MASTER_VERSION                    MASTER_IP      MACHINE_TYPE   NODE_VERSION  NUM_NODES  STATUS
my-cluster  us-east1-c  1.7.8-gke.0 ALPHA (29 days left)  35.227.111.43  n1-standard-1  1.7.8-gke.0   5          RUNNING
```

## Creating the admin clusterrolebinding

For deploying Kubeless in your cluster, your user should have permissions enough for creating cluster roles and cluster role bindings. For doing so you need to give your current GKE account admin privileges in the new cluster. This is not being done by default so you need to do it manually:

```
kubectl create clusterrolebinding kubeless-cluster-admin --clusterrole=cluster-admin --user=<your-gke-user>
```

The above command may fail with `Error from server (Forbidden): User "your-gke-user" cannot create clusterrolebindings.rbac.authorization.k8s.io at the cluster scope`. This error is showed since your account doesn't have privileges to create `clusterrolebindings` (even if you are able to create clusters). If that is the case you can still perform the above operation but using as user the `admin` user that is created by default in the cluster. You can retrieve the admin password executing:

```
gcloud container clusters describe my-cluster --zone <my-cluster-zone>
```

Once you have the admin password you can retry the command above:

```
kubectl --username=admin --password=<admin_password> create clusterrolebinding kubeless-cluster-admin --clusterrole=cluster-admin --user=<your-gke-user>
```

## Deploying Kubeless

After that your are finally able to deploy Kubeless. Get the latest release from the [release page](https://github.com/kubeless/kubeless/releases) and deploy the RBAC version of the Kubeless manifest:

```
export KUBELESS_VERSION=<latest version>
kubectl create namespace kubeless
kubectl create -f https://github.com/kubeless/kubeless/releases/download/v$KUBELESS_VERSION/kubeless-rbac-v$KUBELESS_VERSION.yaml
```
