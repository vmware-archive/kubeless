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
@test "Deploy and wait for Kinesalite" {
  deploy_kinesis_trigger_controller
  wait_for_kubeless_kinesis_controller_ready
  deploy_kinesalite
  wait_for_kinesalite_pod
}
@test "Test function: stream-python-kinesis" {
  deploy_function python-kinesis
  verify_function python-kinesis
  kubeless_function_delete python-kinesis
}
@test "Test function: stream-multi-record-pubish-python-kinesis" {
  deploy_function python-kinesis-multi-record
  verify_function python-kinesis-multi-record
  kubeless_function_delete python-kinesis-multi-record
}
