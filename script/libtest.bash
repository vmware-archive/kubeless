#!/bin/bash
# k8s and kubeless helpers, specially "wait"-ers on pod ready/deleted/etc

KUBELESS_JSONNET=kubeless.jsonnet
KUBELESS_JSONNET_RBAC=kubeless-rbac-novols.jsonnet

KUBECTL_BIN=$(which kubectl)
KUBECFG_BIN=$(which kubecfg)

export TEST_MAX_WAIT_SEC=120

# Workaround 'bats' lack of forced output support, dup() stderr fd
exec 9>&2
echo_info() {
    test -z "$TEST_DEBUG" && return 0
    echo "INFO: $*" >&9
}
export -f echo_info

kubectl() {
    ${KUBECTL_BIN:?} --context=${TEST_CONTEXT:?} "$@"
}
kubecfg() {
    ${KUBECFG_BIN:?} --context=${TEST_CONTEXT:?} "$@"
}

## k8s specific Helper functions
k8s_wait_for_pod_ready() {
    echo_info "Waiting for pod '${@}' to be ready ... "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    until kubectl get pod "${@}" |&grep -q Running; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
k8s_wait_for_pod_gone() {
    echo_info "Waiting for pod '${@}' to be gone ... "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    until kubectl get pod "${@}" |&grep -q No.resources.found; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
k8s_wait_for_pod_logline() {
    local string="${1:?}"; shift
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for '${@}' to show logline '${string}' ..."
    until kubectl logs --tail=10 "${@}"|&grep -q "${string}"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
k8s_context_save() {
    TEST_CONTEXT_SAVED=$(${KUBECTL_BIN} config current-context)
    # Kubeless doesn't support contexts yet, save+restore it
    # Don't save current_context if it's the same already
    [[ $TEST_CONTEXT_SAVED == $TEST_CONTEXT ]] && TEST_CONTEXT_SAVED=""

    # Save current_context
    [[ $TEST_CONTEXT_SAVED != "" ]] && \
        echo_info "Saved context: '${TEST_CONTEXT_SAVED}'" && \
        ${KUBECTL_BIN} config use-context ${TEST_CONTEXT}
}
k8s_context_restore() {
    # Restore saved context
    [[ $TEST_CONTEXT_SAVED != "" ]] && \
        echo_info "Restoring context: '${TEST_CONTEXT_SAVED}'" && \
        ${KUBECTL_BIN} config use-context ${TEST_CONTEXT_SAVED}
}
_wait_for_cmd_ok() {
    local cmd="${*:?}"; shift
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for '${*}' to successfully exit ..."
    until env ${cmd}; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}

## Specific for kubeless
kubeless_recreate() {
    local jsonnet_del=${1:?missing jsonnet delete manifest} jsonnet_upd=${2:?missing jsonnet update manifest}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Delete kubeless namespace, wait to be gone ... "
    kubecfg delete ${jsonnet_del}
    kubectl delete namespace kubeless >& /dev/null || true
    while kubectl get namespace kubeless >& /dev/null; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
    kubectl create namespace kubeless
    kubecfg update ${jsonnet_upd}
}
kubeless_function_delete() {
    local func=${1:?}; shift
    echo_info "Deleting function "${func}" in case still present ... "
    kubeless function delete "${func}" >& /dev/null || true
    kubectl delete all -l function="${func}" > /dev/null || true
    k8s_wait_for_pod_gone -l function="${func}"
}
kubeless_function_deploy() {
    local func=${1:?}; shift
    echo_info "Deploying function ..."
    kubeless function deploy ${func} ${@}
}
kubeless_function_exp_regex_call() {
    local exp_rc=${1:?} regex=${2:?} func=${3:?}; shift 3
    echo_info "Calling function ${func}, expecting rc=${exp_rc} "
    kubeless function call ${func} "${@}"|&egrep ${regex}
}
_wait_for_kubeless_controller_ready() {
    echo_info "Waiting for kubeless controller to be ready ... "
    k8s_wait_for_pod_ready -n kubeless -l kubeless=controller
    _wait_for_cmd_ok kubectl get functions 2>/dev/null
}
_wait_for_controller_logline() {
    local string="${1:?}"
    k8s_wait_for_pod_logline "${string}" -n kubeless -l kubeless=controller
}
_wait_for_simple_function_pod_ready() {
    k8s_wait_for_pod_ready -l function=get-python
}
_deploy_simple_function() {
    make -C examples get-python
}
_call_simple_function() {
    # Artifact to dodge 'bats' lack of support for positively testing _for_ errors
    case "${1:?}" in
        1) make -C examples get-python-verify |& egrep Error.1;;
        0) make -C examples get-python-verify;;
    esac
}
_delete_simple_function() {
    kubeless_function_delete get-python
}

## Entry points used by 'bats' tests:
verify_k8s_tools() {
    local tools="kubectl kubecfg kubeless"
    for exe in $tools; do
        which ${exe} >/dev/null && continue
        echo "ERROR: '${exe}' needs to be installed"
        return 1
    done
}
verify_minikube_running () {
    [[ $TEST_CONTEXT == minikube ]] || return 0
    minikube status | grep -q "minikube: Running" && return 0
    echo "ERROR: minikube not running."
    return 1
}
verify_rbac_mode() {
    kubectl api-versions |&grep -q rbac && return 0
    echo "ERROR: Please run w/RBAC, eg minikube as: minikube start --extra-config=apiserver.Authorization.Mode=RBAC"
    return 1
}
test_must_fail_without_rbac_roles() {
    echo_info "RBAC TEST: function deploy/call must fail without RBAC roles"
    _delete_simple_function
    kubeless_recreate $KUBELESS_JSONNET_RBAC $KUBELESS_JSONNET
    _wait_for_kubeless_controller_ready
    _deploy_simple_function
    _wait_for_controller_logline "User.*cannot"
    _call_simple_function 1
}
test_must_pass_with_rbac_roles() {
    echo_info "RBAC TEST: function deploy/call must succeed with RBAC roles"
    _delete_simple_function
    kubeless_recreate $KUBELESS_JSONNET_RBAC $KUBELESS_JSONNET_RBAC
    _wait_for_kubeless_controller_ready
    _deploy_simple_function
    _wait_for_controller_logline "controller synced and ready"
    _wait_for_simple_function_pod_ready
    _call_simple_function 0
}

test_kubeless_function() {
    local func=${1:?}
    echo_info "TEST: $func"
    kubeless_function_delete ${func}
    make -sC examples ${func}
    k8s_wait_for_pod_ready -l function=${func}
    make -sC examples ${func}-verify
}
# vim: sw=4 ts=4 et si
