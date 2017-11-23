#!/usr/bin/env ruby
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
#
require 'sinatra'
require 'kafka'
require 'timeout'

# Don't buffer stdout
$stdout.sync = true
MOD_NAME = ENV['MOD_NAME']
FUNC_HANDLER = ENV['FUNC_HANDLER']
MOD_ROOT_PATH = ENV.fetch('MOD_ROOT_PATH', '/kubeless/')
MOD_PATH = "#{File.join(MOD_ROOT_PATH, MOD_NAME)}.rb"
FUNC_TIMEOUT = ENV.fetch('FUNC_TIMEOUT', '180').to_i

TOPIC_NAME = ENV['TOPIC_NAME']
KAFKA_SVC = ENV.fetch('KUBELESS_KAFKA_SVC', 'kafka')
KAFKA_NAMESPACE = ENV.fetch('KUBELESS_KAFKA_NAMESPACE', 'kubeless')
KAFKA_HOST = "#{KAFKA_SVC}.#{KAFKA_NAMESPACE}:9092";

begin
  puts "Loading #{MOD_PATH}"
  mod = Module.new
  mod.module_eval(File.read(MOD_PATH))
  # export the function handler
  mod.module_eval("module_function :#{FUNC_HANDLER}")
rescue
  puts "No valid function found for the name: #{MOD_NAME}.#{FUNC_HANDLER}, failed to import module"
  raise
end

kafka = Kafka.new(seed_brokers: [KAFKA_HOST])
consumer = kafka.consumer(group_id: "ruby#{MOD_NAME}#{FUNC_HANDLER}")
consumer.subscribe(TOPIC_NAME)
trap("TERM") { consumer.stop }

consumer.each_message do |message|
  begin
    status = Timeout::timeout(FUNC_TIMEOUT) {
      mod.send(FUNC_HANDLER.to_sym, message.value)
    }
  rescue => e
    puts "ERROR: " + e.to_s
  end
end

set :server, 'webrick'
set :port, 8080

get '/healthz' do
  return 'OK'
end
