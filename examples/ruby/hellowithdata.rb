def handler(event, context)
  puts event
  JSON.generate(event[:data])
end
