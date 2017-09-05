def handler(request)
  payload = request.body.read
  puts JSON.parse(payload)
  return payload
end
