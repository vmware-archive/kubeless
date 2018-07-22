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

@test "Ensure build step" {
  kubectl get -n kubeless configMap kubeless-config -o yaml | grep enable-build-step | grep true
  kubectl get -n kubeless configMap kubeless-config -o yaml | grep function-registry-tls-verify | grep false
  kubectl get secret kubeless-registry-credentials ||
    kubectl create secret docker-registry kubeless-registry-credentials \
      --docker-server=http://$(minikube ip):5000/v2 \
      --docker-username="user" \
      --docker-password="password" \
      --docker-email="email"
}

@test "Deploy a function using the build system" {
  deploy_function get-python
  wait_for_job get-python
  curl http://$(minikube ip):5000/v2/_catalog
  # Speed up pod start when the image is ready
  restart_function get-python
  verify_function get-python
  kubectl logs -n kubeless -l kubeless=controller -c kubeless-function-controller | grep "Started function build job"
  kubectl get deployment -o yaml get-python | grep image | grep $(minikube ip):5000
}

@test "Deploy a Golang function using the build system" {
  deploy_function get-go-deps
  wait_for_job get-go-deps
  # Speed up pod start when the image is ready
  restart_function get-go-deps
  verify_function get-go-deps
  kubectl get deployment -o yaml get-go-deps | grep image | grep $(minikube ip):5000
}

@test "Test no-errors" {
  if kubectl logs -n kubeless -l kubeless=controller | grep "level=error"; then
    echo "Found errors in the controller logs"
    false
  fi
}
