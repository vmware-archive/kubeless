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
@test "Deploy functions to evaluate" {
  deploy_function get-python
  deploy_function get-python-deps
  deploy_function get-python-custom-port
  deploy_function timeout-nodejs
  deploy_function get-nodejs-multi
  deploy_function get-python-metadata
  deploy_function get-python-secrets
  deploy_function post-python
  deploy_function custom-get-python
  deploy_function get-python-url-deps
  deploy_function get-node-url-zip
}
@test "Test function: get-python" {
  verify_function get-python
}
@test "Test function: get-python-deps" {
  verify_function get-python-deps
}
@test "Test function: get-python-custom-port" {
  verify_function get-python-custom-port
}
@test "Test function update: get-python" {
  test_kubeless_function_update get-python
}
@test "Test function update: get-python-deps" {
  test_kubeless_function_update get-python-deps
  kubeless_function_delete get-python-deps
}
@test "Test function autoscale: get-python" {
  if kubectl api-versions | tr '\n' ' ' | grep -q -v "autoscaling/v2beta1"; then
    skip "Autoscale is only supported for Kubernetes >= 1.8"
  fi
  test_kubeless_autoscale get-python
  kubeless_function_delete get-python
}
@test "Test function: timeout-nodejs" {
  verify_function timeout-nodejs
  kubeless_function_delete timeout-nodejs
}
@test "Test function: get-nodejs-multi" {
  verify_function get-nodejs-multi
  kubeless_function_delete get-nodejs-multi
}
@test "Test custom runtime image" {
  verify_function custom-get-python
  test_kubeless_function_update custom-get-python
  kubeless_function_delete custom-get-python
}
@test "Test function: post-python" {
  verify_function post-python
  kubeless_function_delete post-python
}
@test "Test function: get-python-metadata" {
  verify_function get-python-metadata
  kubeless_function_delete get-python-metadata
}
@test "Test function: get-python-secrets" {
  verify_function get-python-secrets
  kubeless_function_delete get-python-secrets
}
@test "Test no-errors" {
  if kubectl logs -n kubeless -l kubeless=controller | grep "level=error"; then
    echo "Found errors in the controller logs"
    false
  fi
}
@test "Test function: get-python-url-deps" {
  verify_function get-python-url-deps
  kubeless_function_delete get-python-url-deps
}
@test "Test function: get-node-url-zip" {
  verify_function get-node-url-zip
  kubeless_function_delete get-node-url-zip
}
# vim: ts=2 sw=2 si et syntax=sh
