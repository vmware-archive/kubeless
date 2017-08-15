#!/bin/bash
# Bring up kubeadm-dind-cluster (docker-in-docker k8s cluster)
DIND_CLUSTER_SH=dind-cluster-v1.7.sh
DIND_URL=https://cdn.rawgit.com/Mirantis/kubeadm-dind-cluster/master/fixed/${DIND_CLUSTER_SH}

rm -f ${DIND_CLUSTER_SH}
wget ${DIND_URL}
chmod +x ${DIND_CLUSTER_SH}
./${DIND_CLUSTER_SH} up
# vim: sw=4 ts=4 et si
