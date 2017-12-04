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
@test "Verify k8s RBAC mode" {
  verify_rbac_mode
}
@test "Test simple function failure without RBAC rules" {
  test_must_fail_without_rbac_roles
}
@test "Redeploy with proper RBAC rules" {
  redeploy_with_rbac_roles
}
# 'bats' lacks loop support, unroll-them-all ->
@test "Deploy functions to evaluate" {
  deploy_function get-python
  deploy_function get-python-deps
  deploy_function scheduled-get-python
  deploy_function get-nodejs
  deploy_function get-nodejs-deps
  deploy_function timeout-nodejs
  deploy_function get-nodejs-multi
  deploy_function get-ruby
  deploy_function get-ruby-deps
  deploy_function get-python-metadata
  deploy_function post-python
  deploy_function post-nodejs
  deploy_function post-ruby
  deploy_function custom-get-python
}
@test "Test function: get-python" {
  verify_function get-python
}
@test "Test function: get-python-deps" {
  verify_function get-python-deps
}
@test "Test function update: get-python" {
  test_kubeless_function_update get-python
}
@test "Test function update: get-python-deps" {
  test_kubeless_function_update get-python-deps
}
@test "Test function ingress: get-python" {
  test_kubeless_ingress get-python
}
@test "Test function autoscale: get-python" {
  test_kubeless_autoscale get-python
}
@test "Test function: get-nodejs" {
  verify_function get-nodejs
}
@test "Test function: get-nodejs-deps" {
  verify_function get-nodejs-deps
}
@test "Test function: timeout-nodejs" {
  verify_function timeout-nodejs
}
@test "Test function: get-nodejs-multi" {
  verify_function get-nodejs-multi
}
@test "Test function: get-ruby" {
  verify_function get-ruby
}
@test "Test function: get-ruby-deps" {
  verify_function get-ruby-deps
}
@test "Test function: get-dotnetcore" {
  skip "This test is flaky until kubeless/kubeless/issues/395 is fixed"
  test_kubeless_function get-dotnetcore
}
@test "Test custom runtime image" {
  verify_function custom-get-python
  test_kubeless_function_update custom-get-python
}
@test "Test function: post-python" {
  verify_function post-python
}
@test "Test function: post-nodejs" {
  verify_function post-nodejs
}
@test "Test function: post-ruby" {
  verify_function post-ruby
}
@test "Test function: post-dotnetcore" {
  skip "This test is flaky until kubeless/kubeless/issues/395 is fixed"
  test_kubeless_function post-dotnetcore
}
@test "Test function: get-python-metadata" {
  verify_function get-python-metadata
}
@test "Deploy functions to evaluate (kafka dependent)" {
  wait_for_kubeless_kafka_server_ready
  deploy_function pubsub-python
  deploy_function pubsub-python34
  deploy_function pubsub-nodejs
  deploy_function pubsub-ruby
}
@test "Test function: pubsub-python" {
  verify_function pubsub-python
}
@test "Test function: pubsub-python34" {
  verify_function pubsub-python34
}
@test "Test function update: pubsub-python" {
  verify_function pubsub-python
}
@test "Test function: pubsub-nodejs" {
  verify_function pubsub-nodejs
}
@test "Test function: pubsub-ruby" {
  verify_function pubsub-ruby
}
@test "Test function: scheduled-get-python" {
  # Now we can verify the scheduled function
  # without having to wait
  verify_function scheduled-get-python
}
@test "Test topic list" {
  wait_for_kubeless_kafka_server_ready
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
@test "Verify Kafka after restart (if context=='minikube')" {
    local topic=$RANDOM
    kubeless topic create $topic
    sts_restart
    kubeless topic list | grep $topic
}
# vim: ts=2 sw=2 si et syntax=sh
