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

@test "Wait for Ingress" {
  wait_for_ingress
}

@test "Create HTTP Trigger" {
    deploy_function get-python
    verify_function get-python
    create_http_trigger get-python "test.domain"
    verify_http_trigger get-python $(minikube ip) "hello.*world" "test.domain"
    update_http_trigger get-python "test.domain-updated"
    verify_http_trigger get-python $(minikube ip) "hello.*world" "test.domain-updated"
    delete_http_trigger get-python
    verify_clean_object httptrigger ing-get-python
    verify_clean_object ingress ing-get-python
}

@test "Create HTTP Trigger with a path" {
    deploy_function get-python
    verify_function get-python
    create_http_trigger get-python "test.domain" "get-python"
    verify_http_trigger get-python $(minikube ip) "hello.*world" "test.domain" "get-python"
    delete_http_trigger get-python
    verify_clean_object httptrigger ing-get-python
    verify_clean_object ingress ing-get-python
}

@test "Create HTTP Trigger with TLS private key and certificate" {
    deploy_function get-python
    verify_function get-python
    create_tls_secret_from_key_cert foo-secret
    create_http_trigger_with_tls_secret get-python "foo.bar.com" "get-python" "foo-secret"
    verify_https_trigger get-python $(minikube ip) "hello.*world" "foo.bar.com" "get-python"
    delete_http_trigger get-python
    verify_clean_object httptrigger ing-get-python
    verify_clean_object ingress ing-get-python
}

@test "Create HTTP Trigger with basic auth" {
    deploy_function get-python
    verify_function get-python
    create_basic_auth_secret "basic-auth"
    create_http_trigger get-python "test.domain"  "get-python" "basic-auth" "nginx"
    verify_http_trigger_basic_auth get-python $(minikube ip) "hello.*world" "test.domain" "get-python" "foo:bar"
    delete_http_trigger get-python
    verify_clean_object httptrigger ing-get-python
    verify_clean_object ingress ing-get-python
}

@test "Test no-errors" {
  if kubectl logs -n kubeless -l kubeless=controller | grep "level=error"; then
    echo "Found errors in the controller logs"
    false
  fi
}
