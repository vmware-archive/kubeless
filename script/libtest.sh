#!/bin/bash
# k8s helpers, specially "wait"-ers on pod ready/deleted/etc

set -u

T_BOLD=$(test -t 0 && tput bold)
T_SGR0=$(test -t 0 && tput sgr0)

typeset -i TOTAL_PASS=0
typeset -i TOTAL_FAIL=0

KUBECTL_BIN=$(which kubectl)
KUBECFG_BIN=$(which kubecfg)

# Wrapup kubectl, kubecfg for --context
test_reset() {
    TEST_CONTEXT=${1:?missing k8s context, e.g.: minikube}
    TOTAL_PASS=0
    TOTAL_FAIL=0
}
test_report() {
    info "$* ${T_BOLD}PASS=$TOTAL_PASS FAIL=$TOTAL_FAIL${T_SGR0}"
    return $TOTAL_FAIL
}

kubectl() {
    ${KUBECTL_BIN:?} --context=${TEST_CONTEXT:?} "$@"
}
kubecfg() {
    ${KUBECFG_BIN:?} --context=${TEST_CONTEXT:?} "$@"
}

## Generic helper functions
info () {
    echo "INFO: $@"
}
spin() {
    local on_tty
    test -t 1 && on_tty=true || on_tty=false
    echo -n .; sleep $1; $on_tty && echo -ne "\r"; sleep $1
}

pass_or_fail() {
    local rc=${1:?} exp_rc=${2:?} msg="${3:?}" status
    [[ ${rc} == ${exp_rc} ]] \
        && { status=PASS; TOTAL_PASS=TOTAL_PASS+1; } \
        || { status=FAIL; TOTAL_FAIL=TOTAL_FAIL+1; }
    echo "${T_BOLD}${status}${T_SGR0}: ${msg}"
    [[ ${status} = PASS ]]
}

## Pre-run verifications
verify_k8s_tools() {
    local tools="kubectl kubecfg kubeless"
    info "VERIFY: k8s tools installed: $tools"
    for exe in $tools; do
        which ${exe} >/dev/null && continue
        echo "ERROR: '${exe}' needs to be installed"
        return 1
    done
}

verify_minikube_running () {
    [[ $TEST_CONTEXT == minikube ]] || return 0
    info "VERIFY: minikube running ..."
    minikube status | grep -q "minikube: Running" && return 0
    echo "ERROR: minikube not running."
    return 1
}
verify_rbac_mode() {
    info "VERIFY: context=${TEST_CONTEXT} running with RBAC ... "
    kubectl api-versions |&grep -q rbac && return 0
    echo "ERROR: Please run w/RBAC, eg minikube as: minikube start --extra-config=apiserver.Authorization.Mode=RBAC"
    return 1
}

## k8s specific Helper functions
k8s_wait_for_pod_ready() {
    info "Waiting for pod '${@}' to be ready ... "
    until kubectl get pod "${@}" |&grep -q Running; do
        spin 0.5
    done
}
k8s_wait_for_pod_gone() {
    info "Waiting for pod '${@}' to be gone ... "
    until kubectl get pod "${@}" |&grep -q No.resources.found; do
        spin 0.5
    done
}
k8s_wait_for_pod_logline() {
    local string="${1:?}"; shift
    info "Waiting for '${@}' to show logline '${string}' ..."
    until kubectl logs --tail=10 "${@}"|&grep -q "${string}"; do
        spin 0.5
    done
    pass_or_fail $? 0 "Found logline: '$string'"
}
kubeless_recreate() {
    local jsonnet_del=${1:?missing jsonnet delete manifest} jsonnet_upd=${2:?missing jsonnet update manifest}
    info "Delete kubeless namespace, wait to be gone ... "
    kubecfg delete ${jsonnet_del}
    kubectl delete namespace kubeless >& /dev/null || true
    while kubectl get namespace kubeless >& /dev/null; do
        spin 0.5
    done
    kubectl create namespace kubeless
    kubecfg update ${jsonnet_upd}
}
kubeless_function_delete() {
    local func=${1:?}; shift
    info "Deleting function "${func}" in case still present ... "
    kubeless function delete "${func}" >& /dev/null || true
    kubectl delete all -l function="${func}" > /dev/null || true
    k8s_wait_for_pod_gone -l function="${func}"
}
kubeless_function_deploy() {
    local func=${1:?}; shift
    info "Deploying function ..."
    kubeless function deploy ${func} ${@}
}
kubeless_function_exp_regex_call() {
    local exp_rc=${1:?} regex=${2:?} func=${3:?}; shift 3
    info "Calling function ${func}, expecting rc=${exp_rc} "
    kubeless function call ${func} "${@}"|&egrep ${regex}
    pass_or_fail $? ${exp_rc} "called function ${func} got rc=$?"
}
k8s_context_save() {
    TEST_CONTEXT_SAVED=$(${KUBECTL_BIN} config current-context)
    # Kubeless doesn't support contexts yet, save+restore it
    # Don't save current_context if it's the same already
    [[ $TEST_CONTEXT_SAVED == $TEST_CONTEXT ]] && TEST_CONTEXT_SAVED=""

    # Save current_context
    [[ $TEST_CONTEXT_SAVED != "" ]] && \
        info "Saved context: '${TEST_CONTEXT_SAVED}'" && \
        ${KUBECTL_BIN} config use-context ${TEST_CONTEXT}
}
k8s_context_restore() {
    # Restore saved context
    [[ $TEST_CONTEXT_SAVED != "" ]] && \
        info "Restoring context: '${TEST_CONTEXT_SAVED}'" && \
        ${KUBECTL_BIN} config use-context ${TEST_CONTEXT_SAVED}
}

# vim: sw=4 ts=4 et si
