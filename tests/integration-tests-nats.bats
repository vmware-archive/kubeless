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

# 'bats' lacks loop support, unroll-them-all ->
@test "Deploy and wait for NATS" {
  deploy_nats_operator
  wait_for_kubeless_nats_operator_ready
  deploy_nats_trigger_controller
  wait_for_kubeless_nats_controller_ready
  deploy_nats_cluster
  wait_for_kubeless_nats_cluster_ready
  expose_nats_service
}
@test "Test function: pubsub-python-nats" {
  deploy_function python-nats
  verify_function python-nats
  kubeless_function_delete python-nats
}
@test "Test 1:n association between NATS trigger and functions" {
  deploy_function nats-python-func1-topic-test
  deploy_function nats-python-func2-topic-test
  deploy_nats_trigger nats-python-trigger-topic-test
  verify_function nats-python-func1-topic-test
  verify_function nats-python-func2-topic-test
  kubeless_function_delete nats-python-func1-topic-test
  kubeless_function_delete nats-python-func2-topic-test
}
@test "Test 1:n association between function and NATS triggers" {
  deploy_function nats-python-func-multi-topic
  deploy_nats_trigger nats-python-trigger-topic1
  deploy_nats_trigger nats-python-trigger-topic2
  verify_function nats-python-func-multi-topic
  kubeless_function_delete nats-python-func-multi-topic
}