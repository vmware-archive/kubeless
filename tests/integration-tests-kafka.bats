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
@test "Wait for kafka" {
  deploy_kafka
  wait_for_kubeless_kafka_server_ready
}
@test "Test function: pubsub-python" {
  deploy_function pubsub-python
  verify_function pubsub-python
  kubeless_function_delete pubsub-python
}
@test "Test function: pubsub-python34" {
  deploy_function pubsub-python34
  verify_function pubsub-python34
  kubeless_function_delete pubsub-python34
}
@test "Test 1:n association between Kafka trigger and functions" {
  deploy_function kafka-python-func1-topic-s3-python
  deploy_function kafka-python-func2-topic-s3-python
  deploy_kafka_trigger s3-python-kafka-trigger
  verify_function kafka-python-func1-topic-s3-python
  verify_function kafka-python-func2-topic-s3-python
  kubeless_function_delete kafka-python-func1-topic-s3-python
  kubeless_function_delete kafka-python-func2-topic-s3-python
}
@test "Test function: pubsub-nodejs" {
  deploy_function pubsub-nodejs
  verify_function pubsub-nodejs
  test_kubeless_function_update pubsub-nodejs
  kubeless_function_delete pubsub-nodejs
}
@test "Test function: pubsub-ruby" {
  deploy_function pubsub-ruby
  verify_function pubsub-ruby
  kubeless_function_delete pubsub-ruby
}
@test "Test function: pubsub-go" {
  deploy_function pubsub-go
  verify_function pubsub-go
  kubeless_function_delete pubsub-go
}
@test "Test topic list" {
  wait_for_kubeless_kafka_server_ready
  for topic in topic1 topic2; do
    kubeless topic create $topic
    _wait_for_kubeless_kafka_topic_ready $topic
  done

  kubeless topic list >$BATS_TMPDIR/kubeless-topic-list
  grep -qxF topic1 $BATS_TMPDIR/kubeless-topic-list
  grep -qxF topic2 $BATS_TMPDIR/kubeless-topic-list
}
@test "Test topic deletion" {
  test_topic_deletion
}
@test "Verify Kafka after restart (if context=='minikube')" {
    local topic=$RANDOM
    kubeless topic create $topic
    sts_restart
    kubeless topic list | grep $topic
}
# vim: ts=2 sw=2 si et syntax=sh
