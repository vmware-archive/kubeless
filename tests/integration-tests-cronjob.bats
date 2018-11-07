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

@test "Create Cronjob Trigger" {
    deploy_function get-python
    verify_function get-python
    create_cronjob_trigger get-python '* * * * *'
    verify_cronjob_trigger get-python '* * * * *' '"GET / HTTP/1.1" 200 11 "" "curl/7.56.1"'
    update_cronjob_trigger get-python '*/60 * * * *'
    verify_cronjob_trigger get-python '*/60 * * * *' '"GET / HTTP/1.1" 200 11 "" "curl/7.56.1"'
    delete_cronjob_trigger get-python
    verify_clean_object cronjobtrigger get-python
}

@test "Test no-errors" {
  if kubectl logs -n kubeless -l kubeless=controller | grep "level=error"; then
    echo "Found errors in the controller logs"
    false
  fi
}
