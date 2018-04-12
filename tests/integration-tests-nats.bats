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
  deploy_function pubsub-python-nats
  verify_function pubsub-python-nats
  kubeless_function_delete pubsub-python-nats
}