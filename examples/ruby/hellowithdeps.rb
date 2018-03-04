require 'logging'

def foo(event, context)
  logging = Logging.logger(STDOUT)
  logging.info "it works!"
  "hello world"
end
