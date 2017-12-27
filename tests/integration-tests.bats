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
  deploy_function scheduled-get-python
  deploy_function get-python-custom-port
  deploy_function get-nodejs
  deploy_function get-nodejs-custom-port
  deploy_function get-nodejs-deps
  deploy_function timeout-nodejs
  deploy_function get-nodejs-multi
  deploy_function get-ruby
  deploy_function get-ruby-custom-port
  deploy_function get-ruby-deps
  deploy_function get-python-metadata
  deploy_function post-python
  deploy_function post-python-custom-port
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
@test "Test function ingress: get-python" {
  test_kubeless_ingress get-python
}
@test "Test function autoscale: get-python" {
  test_kubeless_autoscale get-python
  kubeless_function_delete get-python
}
@test "Test function: get-nodejs" {
  verify_function get-nodejs
  kubeless_function_delete get-nodejs
}
@test "Test function: get-nodejs-custom-port" {
  verify_function get-nodejs-custom-port
  kubeless_function_delete get-nodejs-custom-port
}
@test "Test function: get-nodejs-deps" {
  verify_function get-nodejs-deps
  kubeless_function_delete get-nodejs-deps
}
@test "Test function: timeout-nodejs" {
  verify_function timeout-nodejs
  kubeless_function_delete timeout-nodejs
}
@test "Test function: get-nodejs-multi" {
  verify_function get-nodejs-multi
  kubeless_function_delete get-nodejs-multi
}
@test "Test function: get-ruby" {
  verify_function get-ruby
  kubeless_function_delete get-ruby
}
@test "Test function: get-ruby-custom-port" {
  verify_function get-ruby-custom-port
  kubeless_function_delete get-ruby-custom-port
}
@test "Test function: get-ruby-deps" {
  verify_function get-ruby-deps
  kubeless_function_delete get-ruby-deps
}
@test "Test function: get-dotnetcore" {
  skip "This test is flaky until kubeless/kubeless/issues/395 is fixed"
  test_kubeless_function get-dotnetcore
  kubeless_function_delete get-dotnetcore
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
@test "Test function: post-python-custom-port" {
  verify_function post-python-custom-port
  kubeless_function_delete post-python-custom-port
}
@test "Test function: post-nodejs" {
  verify_function post-nodejs
  kubeless_function_delete post-nodejs
}
@test "Test function: post-ruby" {
  verify_function post-ruby
  kubeless_function_delete post-ruby
}
@test "Test function: post-dotnetcore" {
  skip "This test is flaky until kubeless/kubeless/issues/395 is fixed"
  test_kubeless_function post-dotnetcore
  kubeless_function_delete post-dotnetcore
}
@test "Test function: get-python-metadata" {
  verify_function get-python-metadata
  kubeless_function_delete get-python-metadata
}
@test "Test function: scheduled-get-python" {
  # Now we can verify the scheduled function
  # without having to wait
  verify_function scheduled-get-python
  kubeless_function_delete scheduled-get-python
}
# vim: ts=2 sw=2 si et syntax=sh
