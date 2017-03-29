#!/usr/bin/env ruby


modName = ENV['MOD_NAME']
funcHandler = ENV['FUNC_HANDLER']


modRootPath = ENV['MOD_ROOT_PATH'] ? ENV['MOD_ROOT_PATH'] : '/kubeless/'
modPath = [ modRootPath, modName, '.rb' ]
puts('Loading', modPath.join())

begin
  require 'sinatra'
  require  modPath.join()
rescue RuntimeError => e
  puts('No valid module found for the name: lambda, Failed to import module')
  exit 1
end

set :server, 'webrick'
set :port, 8080

get '/' do
	Kubelessfunction.run(request)
end

post '/' do
	Kubelessfunction.run(request)
end

get '/healthz' do
	return 'OK'
end
