#!/bin/bash

CLUSTER=${1:?}
ZONE=${2:?}
VERSION=${3:?}
ADMIN=${4:?}

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
        functions
        cronjobs
        jobs
        deployments
        horizontalpodautoscalers
    )
    for res in "${resources[@]}"; do
        clean $res
    done
    kubectl delete crd functions.k8s.io || true
else
    echo "Creating cluster $CLUSTER in $ZONE (v$VERSION)"
    # Bypass the warning about using alpha features
    echo 'y' | gcloud container clusters create --cluster-version=$VERSION --zone $ZONE $CLUSTER --num-nodes 5 --no-enable-legacy-authorization --enable-kubernetes-alpha --machine-type=n1-standard-2
    # Wait for the cluster to respond
    cnt=20
    until kubectl get pods; do
        ((cnt=cnt-1)) || (echo "Waited 20 seconds but cluster is not reachable" && return 1)
        sleep 1
    done
    kubectl create clusterrolebinding kubeless-cluster-admin --clusterrole=cluster-admin --user=$ADMIN
fi

