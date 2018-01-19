def handler(event, context)
  puts event
  puts context
  JSON.generate(event)
end
