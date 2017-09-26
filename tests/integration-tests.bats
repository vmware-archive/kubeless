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
  test_kubeless_function_update get-python
}
@test "Test function: get-nodejs" {
  test_kubeless_function get-nodejs
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
@test "Test function: pubsub-nodejs" {
  test_kubeless_function pubsub-nodejs
}
@test "Test function: pubsub-ruby" {
  test_kubeless_function pubsub-ruby
}
@test "Test custom runtime image" {
  test_kubeless_function webserver
  test_kubeless_function_update webserver
}
# vim: ts=2 sw=2 si et syntax=sh
