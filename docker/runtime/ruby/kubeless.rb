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
require 'timeout'

# Don't buffer stdout
$stdout.sync = true
MOD_NAME = ENV['MOD_NAME']
FUNC_HANDLER = ENV['FUNC_HANDLER']
MOD_ROOT_PATH = ENV.fetch('MOD_ROOT_PATH', '/kubeless/')
MOD_PATH = "#{File.join(MOD_ROOT_PATH, MOD_NAME)}.rb"
FUNC_TIMEOUT = ENV.fetch('FUNC_TIMEOUT', '180')
ftimeout = FUNC_TIMEOUT.to_i # We need the timeout as a number

function_context = {
    'function-name': FUNC_HANDLER,
    'timeout': FUNC_TIMEOUT,
    'runtime': ENV['RUNTIME'],
    'memory-limit': ENV['MEMORY_LIMIT'],
}

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

set :server, 'webrick'
set :port, 8090

def funcWrapper(mod, t)
    status = Timeout::timeout(t) {
      res = mod.send(FUNC_HANDLER.to_sym, @event, @context)
    }
end

before do
  contentType = request.env["CONTENT_TYPE"]
  data = ''
  if request.env["CONTENT_LENGTH"] != nil && request.env["CONTENT_LENGTH"] > "0"
    if contentType == "application/json"
      data =  JSON.parse(request.body.read)
    else
      data = request.body.read
    end
  end
  @event = {
      data: data,
      'event-id': request.env["HTTP_EVENT_ID"],
      'event-type': request.env["HTTP_EVENT_TYPE"],
      'event-time': request.env["HTTP_EVENT_TIME"],
      'event-namespace': request.env["HTTP_EVENT_NAMESPACE"],
      extensions: {
        request: request,
      }
    }
  @context = function_context
end

get '/' do
  begin
    funcWrapper(mod, ftimeout)
  rescue Timeout::Error
    status 408
  end
end

post '/' do
  begin
    funcWrapper(mod, ftimeout)
  rescue Timeout::Error
    status 408
  end
end

get '/healthz' do
  'OK'
end
