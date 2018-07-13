# Copyright (c) 2016-2017 Bitnami
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# k8s and kubeless helpers, specially "wait"-ers on pod ready/deleted/etc

KUBELESS_MANIFEST=kubeless-non-rbac.yaml
KUBELESS_MANIFEST_RBAC=kubeless.yaml
KAFKA_MANIFEST=kafka-zookeeper.yaml
NATS_MANIFEST=nats.yaml
KINESIS_MANIFEST=kinesis.yaml

KUBECTL_BIN=$(which kubectl)
: ${KUBECTL_BIN:?ERROR: missing binary: kubectl}

export TEST_MAX_WAIT_SEC=300

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

## k8s specific Helper functions
k8s_wait_for_pod_ready() {
    echo_info "Waiting for pod '${@}' to be ready ... "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}

    # Retries just in case it is not stable
    local -i successCount=0
    while [ "$successCount" -lt "3" ]; do
        if kubectl get pod "${@}" | grep -q Running; then
            ((successCount=successCount+1))
        fi
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
k8s_wait_for_pod_count() {
    local pod_cnt=${1:?}; shift
    echo_info "Waiting for pod '${@}' to have count==${pod_cnt} running ... "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    # Retries just in case it is not stable
    local -i successCount=0
    while [ "$successCount" -lt "3" ]; do
        if [[ $(kubectl get pod "${@}" -ogo-template='{{.items|len}}') == ${pod_cnt} ]]; then
            ((successCount=successCount+1))
        fi
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
    k8s_wait_for_pod_ready "${@}"
    echo "Finished waiting"
}
k8s_wait_for_uniq_pod() {
    k8s_wait_for_pod_count 1 "$@"
}
k8s_wait_for_pod_gone() {
    echo_info "Waiting for pod '${@}' to be gone ... "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    until kubectl get pod "${@}" | grep -q No.resources.found; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
k8s_wait_for_pod_logline() {
    local string="${1:?}"; shift
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for '${@}' to show logline '${string}' ..."
    until kubectl logs "${@}"| grep -q "${string}"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
k8s_wait_for_cluster_ready() {
    echo_info "Waiting for k8s cluster to be ready (context=${TEST_CONTEXT}) ..."
    _wait_for_cmd_ok kubectl get po 2>/dev/null && \
    k8s_wait_for_pod_ready -n kube-system -l component=kube-addon-manager && \
    k8s_wait_for_pod_ready -n kube-system -l k8s-app=kube-dns && \
        return 0
    return 1
}
k8s_log_all_pods() {
    local namespaces=${*:?} ns
    for ns in ${*}; do
        echo "### namespace: ${ns} ###"
        kubectl get pod -n ${ns} -oname|xargs -I@ sh -xc "kubectl logs -n ${ns} @|sed 's|^|@: |'"
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
    local manifest_del=${1:?missing delete manifest} manifest_upd=${2:?missing update manifest}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Delete kubeless namespace, wait to be gone ... "
    kubectl delete -f ${manifest_del} || true
    kubectl delete namespace kubeless >& /dev/null || true
    while kubectl get namespace kubeless >& /dev/null; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
    kubectl create namespace kubeless
    kubectl create -f ${manifest_upd}
}
kubeless_function_delete() {
    local func=${1:?}; shift
    echo_info "Deleting function "${func}" in case still present ... "
    kubeless function ls |grep -w "${func}" && kubeless function delete "${func}" >& /dev/null || true
    echo_info "Wait for function "${func}" to be deleted "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    while kubectl get functions "${func}" >& /dev/null; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
kubeless_kafka_trigger_delete() {
    local trigger=${1:?}; shift
    echo_info "Deleting kafka trigger "${trigger}" in case still present ... "
    kubeless trigger kafka list |grep -w "${trigger}" && kubeless trigger kafka delete "${trigger}" >& /dev/null || true
}
kubeless_nats_trigger_delete() {
    local trigger=${1:?}; shift
    echo_info "Deleting NATS trigger "${trigger}" in case still present ... "
    kubeless trigger nats list |grep -w "${trigger}" && kubeless trigger nats delete "${trigger}" >& /dev/null || true
}    
kubeless_function_deploy() {
    local func=${1:?}; shift
    echo_info "Deploying function ..."
    kubeless function deploy ${func} ${@}
}
_wait_for_kubeless_controller_ready() {
    echo_info "Waiting for kubeless controller to be ready ... "
    k8s_wait_for_pod_ready -n kubeless -l kubeless=controller
    _wait_for_cmd_ok kubectl get functions 2>/dev/null
}
_wait_for_kubeless_controller_logline() {
    local string="${1:?}"
    k8s_wait_for_pod_logline "${string}" -n kubeless -l kubeless=controller -c kubeless-function-controller
}
wait_for_ingress() {
    echo_info "Waiting until Nginx pod is ready ..."
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    until kubectl get pods -l name=nginx-ingress-controller -n kube-system>& /dev/null; do
        ((cnt=cnt-1)) || exit 1
        sleep 1
    done
}
wait_for_kubeless_kafka_server_ready() {
    [[ $(kubectl get pod -n kubeless kafka-0 -ojsonpath='{.metadata.annotations.ready}') == true ]] && return 0
    echo_info "Waiting for kafka-0 to be ready ..."
    k8s_wait_for_pod_logline "Kafka.*Server.*started" -n kubeless kafka-0
    echo_info "Waiting for kafka-trigger-controller pod to be ready ..."
    k8s_wait_for_pod_ready -n kubeless -l kubeless=kafka-trigger-controller
    _wait_for_cmd_ok kubectl get kafkatriggers 2>/dev/null
    kubectl annotate pods --overwrite -n kubeless kafka-0 ready=true
}
wait_for_kubeless_nats_operator_ready() {
    echo_info "Waiting for NATS operator pod to be ready ..."
    k8s_wait_for_pod_ready -n nats-io -l name=nats-operator
}
wait_for_kubeless_nats_cluster_ready() {
    echo_info "Waiting for NATS cluster pods to be ready ..."
    k8s_wait_for_pod_ready -n nats-io -l nats_cluster=nats
}
wait_for_kubeless_nats_controller_ready() {
    echo_info "Waiting for NATS controller pods to be ready ..."
    k8s_wait_for_pod_ready -n kubeless -l kubeless=nats-trigger-controller
}
_wait_for_kubeless_kafka_topic_ready() {
    local topic=${1:?}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for kafka-0 topic='${topic}' to be ready ..."
    # zomg enter kafka-0 container to peek for topic already present
    until \
        kubectl exec -n kubeless kafka-0 -- sh -c \
        '/opt/bitnami/kafka/bin/kafka-topics.sh --list --zookeeper $(
            sed -n s/zookeeper.connect=//p /bitnami/kafka/conf/server.properties)'| \
                grep -qw ${topic}
        do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
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
        1) make -C examples get-python-verify |  egrep Error.1;;
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
verify_rbac_mode() {
    kubectl api-versions | grep -q rbac && return 0
    echo "ERROR: Please run w/RBAC, eg minikube as: minikube start --extra-config=apiserver.Authorization.Mode=RBAC"
    return 1
}

wait_for_endpoint() {
    local func=${1:?}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    local endpoint=$(kubectl get endpoints -l function=$func | grep $func | awk '{print $2}')
    echo_info "Waiting for the endpoint ${endpoint}' to be ready ..."
    until curl -s $endpoint; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
wait_for_autoscale() {
    local func=${1:?}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    local hap=$()
    echo_info "Waiting for HAP ${func} to be ready ..."
    until kubectl get horizontalpodautoscalers | grep $func; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
wait_for_job() {
    local func=${1:?}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for build job of ${func} to be finished ..."
    until kubectl get job -l function=${func} -o yaml | grep "succeeded: 1"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
test_must_fail_without_rbac_roles() {
    echo_info "RBAC TEST: function deploy/call must fail without RBAC roles"
    _delete_simple_function
    kubeless_recreate $KUBELESS_MANIFEST_RBAC $KUBELESS_MANIFEST
     _wait_for_kubeless_controller_logline "User.*cannot"
}
redeploy_with_rbac_roles() {
    kubeless_recreate $KUBELESS_MANIFEST_RBAC $KUBELESS_MANIFEST_RBAC
    _wait_for_kubeless_controller_ready
    _wait_for_kubeless_controller_logline "controller synced and ready"
}

deploy_kafka() {
    echo_info "Deploy kafka ... "
    kubectl create -f $KAFKA_MANIFEST
}

deploy_nats_operator() {
    echo_info "Deploy NATS operator ... "
    kubectl apply -f https://raw.githubusercontent.com/nats-io/nats-operator/master/example/deployment-rbac.yaml
}

deploy_nats_cluster() {
    echo_info "Deploy NATS cluster ... "
    kubectl apply -f ./manifests/nats/nats-cluster.yaml -n nats-io
}

deploy_nats_trigger_controller() {
    echo_info "Deploy NATS trigger controller ... "
    kubectl create -f $NATS_MANIFEST
}

expose_nats_service() {
    kubectl get svc nats -n nats-io -o yaml | sed 's/ClusterIP/NodePort/' | kubectl replace -f -
}

deploy_kinesis_trigger_controller() {
    echo_info "Deploy Kinesis trigger controller ... "
    kubectl create -f $KINESIS_MANIFEST
}

wait_for_kubeless_kinesis_controller_ready() {
    echo_info "Waiting for Kinesis trigger controller pods to be ready ..."
    k8s_wait_for_pod_ready -n kubeless -l kubeless=kinesis-trigger-controller
}

deploy_kinesalite() {
    echo_info "Deploy Kinesalite a AWS Kinesis mock server ... "
    kubectl apply -f ./manifests/kinesis/kinesalite.yaml
}

wait_for_kinesalite_pod() {
    echo_info "Waiting for Kinesalite pod to be ready ..."
    k8s_wait_for_pod_ready -l app=kinesis
}

deploy_function() {
    local func=${1:?} func_topic
    echo_info "TEST: $func"
    kubeless_function_delete ${func}
    make -sC examples ${func}
}

deploy_kafka_trigger() {
    local trigger=${1:?}
    echo_info "TEST: $trigger"
    kubeless_kafka_trigger_delete ${trigger}
    make -sC examples ${trigger}
}

deploy_nats_trigger() {
    local trigger=${1:?}
    echo_info "TEST: $trigger"
    kubeless_nats_trigger_delete ${trigger}
    make -sC examples ${trigger}
}

verify_function() {
    local func=${1:?}
    local make_task=${2:-${func}-verify}
    echo_info "Init logs: $(kubectl logs -l function=${func} -c prepare)"
    k8s_wait_for_pod_ready -l function=${func}
    case "${func}" in
        *pubsub*)
            func_topic=$(kubectl get kafkatrigger "${func}" -o yaml|sed -n 's/topic: //p')
            echo_info "FUNC TOPIC: $func_topic"
    esac
    local -i counter=0
    until make -sC examples ${make_task}; do
        echo_info "FUNC ${func} failed. Retrying..."
        ((counter=counter+1))
        if [ "$counter" -ge 3 ]; then
            echo_info "FUNC ${func} failed ${counter} times. Exiting"
            return 1;
        fi
        sleep `expr 10 \* $counter`
    done
}
test_kubeless_function() {
    local func=${1:?}
    deploy_function $func
    verify_function $func
}
update_function() {
    local func=${1:?} func_topic
    echo_info "UPDATE: $func"
    make -sC examples ${func}-update
    sleep 10
    k8s_wait_for_uniq_pod -l function=${func}
}
restart_function() {
    local func=${1:?}
    echo_info "Restarting: $func"
    kubectl delete pod -l function=${func}
    k8s_wait_for_uniq_pod -l function=${func}
}
test_kubeless_function_update() {
    local func=${1:?}
    update_function $func
    verify_function $func ${func}-update-verify
}
create_basic_auth_secret() {
    local secret=${1:?}; shift
    htpasswd -cb auth foo bar
    kubectl create secret generic $secret --from-file=auth
}
create_tls_secret_from_key_cert() {
    local secret=${1:?}; shift
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /tmp/tls.key -out /tmp/tls.crt -subj "/CN=foo.bar.com"
    kubectl create secret tls $secret --key /tmp/tls.key --cert /tmp/tls.crt
}
create_http_trigger_with_tls_secret(){
    local func=${1:?}; shift
    local domain=${1-""};
    local subpath=${2-""};
    local secret=${3-""};
    delete_http_trigger ${func}
    echo_info "TEST: Creating HTTP trigger"
    local command="kubeless trigger http create ing-${func} --function-name ${func}"
    if [ -n "$domain" ]; then
        command="$command --hostname ${domain}"
    fi
    if [ -n "$subpath" ]; then
        command="$command --path ${subpath}"
    fi
    if [ -n "$secret" ]; then
        command="$command --tls-secret ${secret}"
    fi
    eval $command
}
create_http_trigger(){
    local func=${1:?}; shift
    local domain=${1-""};
    local subpath=${2-""};
    local basicauth=${3-""};
    local gateway=${4-""};
    delete_http_trigger ${func}
    echo_info "TEST: Creating HTTP trigger"
    local command="kubeless trigger http create ing-${func} --function-name ${func}"
    if [ -n "$domain" ]; then
        command="$command --hostname ${domain}"
    fi
    if [ -n "$subpath" ]; then
        command="$command --path ${subpath}"
    fi
    if [ -n "$basicauth" ]; then
        command="$command --basic-auth-secret ${basicauth}"
    fi
    if [ -n "$gateway" ]; then
        command="$command --gateway ${gateway}"
    fi
    eval $command
}
update_http_trigger(){
    local func=${1:?}; shift
    local domain=${1:-""}
    local subpath=${2:-""};
    echo_info "TEST: Updating HTTP trigger"
    local command="kubeless trigger http update ing-${func} --function-name ${func}"
    if [ -n "$domain" ]; then
        command="$command --hostname ${domain}"
    fi
    if [ -n "$subpath" ]; then
        command="$command --path ${subpath}"
    fi
    eval $command
}
verify_http_trigger(){
    local func=${1:?}; shift
    local ip=${1:?}; shift
    local expected_response=${1:?}; shift
    local domain=${1:?}; shift
    local subpath=${1:-""};
    kubeless trigger http list | grep ${func}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for ingress to be ready..."
    until kubectl get ingress | grep $func | grep "$domain" | awk '{print $3}' | grep "$ip"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
    sleep 3
    curl -vv --header "Host: $domain" $ip\/$subpath | grep "${expected_response}"
}
verify_http_trigger_basic_auth(){
    local func=${1:?}; shift
    local ip=${1:?}; shift
    local expected_response=${1:?}; shift
    local domain=${1:?}; shift
    local subpath=${1:?}; shift
    local auth=${1:-""};
    kubeless trigger http list | grep ${func}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for ingress to be ready..."
    until kubectl get ingress | grep $func | grep "$domain" | awk '{print $3}' | grep "$ip"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
    sleep 3
    curl -v --header "Host: $domain" $ip\/$subpath | grep "401 Authorization Required"
    curl -v --header "Host: $domain" -u $auth $ip\/$subpath | grep "${expected_response}"
}
verify_https_trigger(){
    local func=${1:?}; shift
    local ip=${1:?}; shift
    local expected_response=${1:?}; shift
    local domain=${1:?}; shift
    local subpath=${1:-""};
    kubeless trigger http list | grep ${func}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    echo_info "Waiting for ingress to be ready..."
    until kubectl get ingress | grep $func | grep "$domain" | awk '{print $3}' | grep "$ip"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
    sleep 3
    curl -k -vv --header "Host: $domain" https:\/\/$ip\/$subpath | grep "${expected_response}"
}
delete_http_trigger() {
    local func=${1:?}; shift
    kubeless trigger http list |grep -w ing-${func} && kubeless trigger http delete ing-${func} >& /dev/null || true
}
create_cronjob_trigger(){
    local func=${1:?}; shift
    local schedule=${1:?};
    delete_cronjob_trigger ${func}
    echo_info "TEST: Creating CronJob trigger"
    kubeless trigger cronjob create ${func} --function ${func} --schedule "${schedule}"
}
update_cronjob_trigger(){
    local func=${1:?}; shift
    local schedule=${1:?};
    echo_info "TEST: Updating CronJob trigger"
    kubeless trigger cronjob update ${func} --function ${func} --schedule "${schedule}"
}
verify_cronjob_trigger(){
    local func=${1:?}; shift
    local schedule=${1:?}; shift
    local expected_log=${1:?}
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    kubeless trigger cronjob list | grep ${func} | grep "${schedule}"
    echo_info "Waiting for CronJob to be executed..."
    until kubectl logs -l function=${func} | grep "$expected_log"; do
        ((cnt=cnt-1)) || return 1
        sleep 1
    done
}
delete_cronjob_trigger() {
    local func=${1:?}; shift
    kubeless trigger cronjob list |grep -w ${func} && kubeless trigger cronjob delete ${func} >& /dev/null || true
}
test_kubeless_autoscale() {
    local func=${1:?} exp_autoscale act_autoscale
    # Use some fixed values
    local val=10 num=3
    echo_info "TEST: autoscale ${func}"
    kubeless autoscale create ${func} --value ${val:?} --min ${num:?} --max ${num:?}
    wait_for_autoscale ${func}
    kubeless autoscale list | fgrep -w ${func}
    act_autoscale=$(kubectl get horizontalpodautoscaler -ojsonpath='{range .items[*].spec}{@.scaleTargetRef.name}:{@.targetCPUUtilizationPercentage}:{@.minReplicas}:{@.maxReplicas}{end}')
    exp_autoscale="${func}:${val}:${num}:${num}"
    [[ ${act_autoscale} == ${exp_autoscale} ]]
    k8s_wait_for_pod_count ${num} -l function="${func}"
    kubeless autoscale delete ${func}
}
test_topic_deletion() {
    local topic=$RANDOM
    local topic_count=0
    kubeless topic create $topic
    kubeless topic delete $topic
    topic_count=$(kubeless topic list | grep $topic | wc -l)
    if [ ${topic_count} -gt 0 ] ; then
     echo_info "Topic $topic still exists"
     exit 200
    fi
}
sts_restart() {
    local num=1
    kubectl delete pod kafka-0 -n kubeless
    kubectl delete pod zoo-0 -n kubeless
    k8s_wait_for_uniq_pod -l kubeless=zookeeper -n kubeless
    k8s_wait_for_uniq_pod -l kubeless=kafka -n kubeless
    wait_for_kubeless_kafka_server_ready
}
verify_clean_object() {
    local type=${1:?}; shift
    local name=${1:?}; shift
    echo_info "Checking if "${type}" exists for function "${name}"... "
    local -i cnt=${TEST_MAX_WAIT_SEC:?}
    until [[ ! $(kubectl get ${type} 2>&1 | grep ${name}) ]]; do
        ((cnt=cnt-1)) || return 1
        sleep 1
        echo_info "$(kubectl get ${type} 2>&1 | grep ${name})"
    done
    echo_info "${type}/${name} is gone"
}
# vim: sw=4 ts=4 et si
