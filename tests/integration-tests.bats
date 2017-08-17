#!/usr/bin/env bats

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
}
@test "Test function: get-nodejs" {
  test_kubeless_function get-nodejs
}
@test "Test function: post-python" {
  test_kubeless_function post-python
}
@test "Test function: post-nodejs" {
  test_kubeless_function post-nodejs
}
@test "Test function: get-python-metadata" {
  test_kubeless_function get-python-metadata
}
@test "Test function: pubsub" {
  test_kubeless_function pubsub
}
# vim: ts=2 sw=2 si et syntax=sh
