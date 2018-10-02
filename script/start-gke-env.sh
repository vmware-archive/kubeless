#!/bin/bash

CLUSTER=${1:?}
ZONE=${2:?}
BRANCH=${3:?}
ADMIN=${4:?}

# Resolve latest version from a branch
VERSION=$(gcloud container get-server-config --zone $ZONE --format='yaml(validMasterVersions)' 2> /dev/null | grep $BRANCH | awk '{print $2}' | head -n 1)

function clean() {
    local resource=${1:?}
    kubectl get $resource | awk '{print $1}' | xargs kubectl delete $resource || true
}
if ! gcloud container clusters list; then
    echo "Unable to access gcloud project"
    exit 1
fi

if gcloud container clusters list | grep -q $CLUSTER; then
    echo "GKE cluster already exits. Deleting resources"
    # Cluster already exists, make sure it is clean
    gcloud container clusters get-credentials $CLUSTER --zone $ZONE
    kubectl delete ns kubeless || true
    resources=(
        cronjobs
        jobs
        deployments
        horizontalpodautoscalers
    )
    for res in "${resources[@]}"; do
        clean $res
    done

    echo "Removing clusterroles"  >&9
    kubectl delete clusterrole kubeless-controller-deployer || true
    kubectl delete clusterrole kafka-controller-deployer || true
    kubectl delete clusterrolebindings kubeless-controller-deployer || true
    kubectl delete clusterrolebindings kafka-controller-deployer || true

    echo "Removing customresourcecleanup.apiextensions.k8s.io finalizer from CRD's"  >&9
    kubectl patch crd/functions.kubeless.io -p '{"metadata":{"finalizers":[]}}' --type=merge       || true
    kubectl patch crd/cronjobtriggers.kubeless.io -p '{"metadata":{"finalizers":[]}}' --type=merge || true
    kubectl patch crd/httptriggers.kubeless.io -p '{"metadata":{"finalizers":[]}}' --type=merge    || true
    kubectl patch crd/kafkatriggers.kubeless.io -p '{"metadata":{"finalizers":[]}}' --type=merge   || true

    echo "Removing finalizers from CRD object's and deleting the CRD objects"  >&9
    functions=$(kubectl get functions -o name)
    for func in $functions; do
        kubectl patch $func -p '{"metadata":{"finalizers":[]}}' --type=merge || true
        kubectl delete $func
    done
    cronjobtriggers=$(kubectl get cronjobtriggers -o name)
    for trigger in $cronjobtriggers; do
        kubectl patch $trigger -p '{"metadata":{"finalizers":[]}}' --type=merge || true
        kubectl delete $trigger
    done
    httptriggers=$(kubectl get httptriggers -o name)
    for trigger in $httptriggers; do
        kubectl patch $trigger -p '{"metadata":{"finalizers":[]}}' --type=merge || true
        kubectl delete $trigger        
    done
    kafkatriggers=$(kubectl get kafkatriggers -o name)
    for trigger in $kafkatriggers; do
        kubectl patch $trigger -p '{"metadata":{"finalizers":[]}}' --type=merge || true
        kubectl delete $trigger        
    done

    echo "Deleting CRD's"  >&9
    kubectl delete crd functions.kubeless.io       || true
    kubectl delete crd cronjobtriggers.kubeless.io || true
    kubectl delete crd httptriggers.kubeless.io    || true
    kubectl delete crd kafkatriggers.kubeless.io   || true
else
    echo "Creating cluster $CLUSTER in $ZONE (v$VERSION)"
    gcloud container clusters create --cluster-version=$VERSION --zone $ZONE $CLUSTER --num-nodes 5 --machine-type=n1-standard-2
    # Wait for the cluster to respond
    cnt=20
    until kubectl get pods; do
        ((cnt=cnt-1)) || (echo "Waited 20 seconds but cluster is not reachable" && return 1)
        sleep 1
    done
    kubectl create clusterrolebinding kubeless-cluster-admin --clusterrole=cluster-admin --user=$ADMIN
fi

