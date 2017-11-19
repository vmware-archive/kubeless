#!/usr/bin/env bats

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

load ../script/libtest

@test "Verify TEST_CONTEXT envvar" {
  : ${TEST_CONTEXT:?}
}
@test "Verify needed kubernetes tools installed" {
  verify_k8s_tools
}
@test "Verify minikube running (if context=='minikube')" {
  verify_minikube_running
}
@test "Verify k8s RBAC mode" {
  verify_rbac_mode
}
@test "Test simple function failure without RBAC rules" {
  test_must_fail_without_rbac_roles
}
@test "Test simple function success with proper RBAC rules" {
  test_must_pass_with_rbac_roles
}
# 'bats' lacks loop support, unroll-them-all ->
@test "Test function: get-python" {
  test_kubeless_function get-python
}
@test "Test function: get-python-deps" {
  test_kubeless_function get-python-deps
}
@test "Test function update: get-python" {
  test_kubeless_function_update get-python
}
@test "Test function ingress: get-python" {
  test_kubeless_ingress get-python
}
@test "Test function autoscale: get-python" {
  test_kubeless_autoscale get-python
}
@test "Test function: get-nodejs" {
  test_kubeless_function get-nodejs
}
@test "Test function: get-nodejs-deps" {
  test_kubeless_function get-nodejs-deps
}
@test "Test function: get-nodejs-multi" {
  test_kubeless_function get-nodejs-multi
}
@test "Test function: get-ruby" {
  test_kubeless_function get-ruby
}
@test "Test function: get-ruby-deps" {
  test_kubeless_function get-ruby-deps
}
@test "Test function: get-dotnetcore" {
  test_kubeless_function get-dotnetcore
}
@test "Test function: post-python" {
  test_kubeless_function post-python
}
@test "Test function: post-nodejs" {
  test_kubeless_function post-nodejs
}
@test "Test function: post-ruby" {
  test_kubeless_function post-ruby
}
@test "Test function: post-dotnetcore" {
  test_kubeless_function post-dotnetcore
}
@test "Test function: get-python-metadata" {
  test_kubeless_function get-python-metadata
}
@test "Test function: pubsub-python" {
  test_kubeless_function pubsub-python
}
@test "Test function update: pubsub-python" {
  test_kubeless_function_update pubsub-python
}
@test "Test function: pubsub-nodejs" {
  test_kubeless_function pubsub-nodejs
}
@test "Test function: pubsub-ruby" {
  test_kubeless_function pubsub-ruby
}
@test "Test topic list" {
  _wait_for_kubeless_kafka_server_ready
  for topic in topic1 topic2; do
    kubeless topic create $topic
    _wait_for_kubeless_kafka_topic_ready $topic
  done

  kubeless topic list >$BATS_TMPDIR/kubeless-topic-list
  grep -qxF topic1 $BATS_TMPDIR/kubeless-topic-list
  grep -qxF topic2 $BATS_TMPDIR/kubeless-topic-list
}
@test "Test topic deletion" {
  test_topic_deletion
}
@test "Test custom runtime image" {
  deploy_function webserver
  wait_for_endpoint webserver
  verify_function webserver
  update_function webserver
  wait_for_endpoint webserver
  verify_update_function webserver
}
@test "Verify Kafka after restart (if context=='minikube')" {
    sts_restart
}
# vim: ts=2 sw=2 si et syntax=sh
