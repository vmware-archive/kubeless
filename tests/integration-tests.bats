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
  deploy_function get-nodejs
  deploy_function get-nodejs-custom-port
  deploy_function get-nodejs-stream
  deploy_function get-nodejs-deps
  deploy_function timeout-nodejs
  deploy_function get-nodejs-multi
  deploy_function get-ruby
  deploy_function get-ruby-custom-port
  deploy_function get-ruby-deps
  deploy_function get-php
  deploy_function get-php-deps
  deploy_function timeout-php
  deploy_function get-go
  deploy_function get-go-custom-port
  deploy_function get-go-deps
  deploy_function timeout-go
  deploy_function get-python-metadata
  deploy_function get-python-secrets
  deploy_function post-python
  deploy_function post-python-custom-port
  deploy_function post-nodejs
  deploy_function post-ruby
  deploy_function post-php
  deploy_function post-go
  deploy_function custom-get-python
  deploy_function get-python-url-deps
  deploy_function get-node-url-zip
  deploy_function get-java
  deploy_function post-java
  deploy_function get-java-deps
  deploy_function get-jvm-java
  deploy_function get-nodejs-distroless
  deploy_function get-nodejs-distroless-deps
  deploy_function get-ballerina
  deploy_function get-ballerina-custom-port
  deploy_function get-ballerina-data
  deploy_function get-ballerina-conf
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
@test "Test function: get-nodejs" {
  verify_function get-nodejs
  kubeless_function_delete get-nodejs
}
@test "Test function: get-nodejs-custom-port" {
  verify_function get-nodejs-custom-port
  kubeless_function_delete get-nodejs-custom-port
}
@test "Test function: get-nodejs-stream" {
  verify_function get-nodejs-stream
  kubeless_function_delete get-nodejs-stream
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
@test "Test function: get-php" {
  verify_function get-php
}
@test "Test function update: get-php" {
  test_kubeless_function_update get-php
  kubeless_function_delete get-php
}
@test "Test function: get-php-deps" {
  verify_function get-php-deps
}
@test "Test function update: get-php-deps" {
  test_kubeless_function_update get-php-deps
  kubeless_function_delete get-php-deps
}
@test "Test function: timeout-php" {
  verify_function timeout-php
  kubeless_function_delete timeout-php
}
@test "Test function: get-go" {
  verify_function get-go
  kubeless_function_delete get-go
}
@test "Test function: get-go-deps" {
  verify_function get-go-deps
}
@test "Test function: get-go-custom-port" {
  verify_function get-go-custom-port
  kubeless_function_delete get-go-custom-port
}
@test "Test function: timeout-go" {
  verify_function timeout-go
  kubeless_function_delete timeout-go
}
@test "Test function: post-go" {
  verify_function post-go
  kubeless_function_delete post-go
}
@test "Test function: get-dotnetcore" {
  test_kubeless_function get-dotnetcore
  kubeless_function_delete get-dotnetcore
}
@test "Test function: get-dotnetcore-dependency" {
  test_kubeless_function get-dotnetcore-dependency
  kubeless_function_delete get-dotnetcore-dependency
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
@test "Test function: post-php" {
  verify_function post-php
  kubeless_function_delete post-php
}
@test "Test function: post-dotnetcore" {
  test_kubeless_function post-dotnetcore
  kubeless_function_delete post-dotnetcore
}
@test "Test function: get-python-metadata" {
  verify_function get-python-metadata
  kubeless_function_delete get-python-metadata
}
@test "Test function: get-python-secrets" {
  verify_function get-python-secrets
  kubeless_function_delete get-python-secrets
}
@test "Test function: get-jvm-java" {
  verify_function get-jvm-java
  kubeless_function_delete get-jvm-java
}
@test "Test function: get-java" {
  verify_function get-java
  kubeless_function_delete get-java
}
@test "Test function: post-java" {
  verify_function post-java
  kubeless_function_delete post-java
}
@test "Test function: get-java-deps" {
  verify_function get-java-deps
}
@test "Test function: get-nodejs-distroless" {
  verify_function get-nodejs-distroless
}
@test "Test function: get-nodejs-distroless-deps" {
  verify_function get-nodejs-distroless-deps
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
@test "Test function: get-ballerina" {
  verify_function get-ballerina
  kubeless_function_delete get-ballerina
}
@test "Test function: get-ballerina-custom-port" {
  verify_function get-ballerina-custom-port
  kubeless_function_delete get-ballerina-custom-port
}
@test "Test function: get-ballerina-data" {
  verify_function get-ballerina-data
  kubeless_function_delete get-ballerina-data
}
@test "Test function: get-ballerina-conf" {
  verify_function get-ballerina-conf
  kubeless_function_delete get-ballerina-conf
}
# vim: ts=2 sw=2 si et syntax=sh
