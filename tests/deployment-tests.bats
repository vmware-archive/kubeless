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
@test "Verify k8s RBAC mode" {
  verify_rbac_mode
}
@test "Test simple function failure without RBAC rules" {
  test_must_fail_without_rbac_roles
}
@test "Redeploy with proper RBAC rules" {
  redeploy_with_rbac_roles
}
