#!/usr/bin/env ruby


modName = ENV['MOD_NAME']
funcHandler = ENV['FUNC_HANDLER']


modRootPath = ENV['MOD_ROOT_PATH'] ? ENV['MOD_ROOT_PATH'] : '/kubeless/'
modPath = [ modRootPath, modName, '.rb' ]  
puts('Loading', modPath.join())

begin 
  require 'sinatra'
rescue (e)
  puts('No valid module found for the name: lambda, Failed to import module')
  exit 1
end

set :server, 'webrick' 
set :port, 8080

get '/' do 
	return "#{ request.env }"
end 

post '/' do 
  	msg = eval File.read(modPath.join())
	return msg
end 

get '/healthz' do 
	return 'OK'
end
