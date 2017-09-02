require 'logging'

def foo(context)
  logging = Logging.logger(STDOUT)
  logging.info "it works!"
  "hello world"
end
