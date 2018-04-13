def printable_response(response)
  [
    'response',
    "status: #{response.code}",
    "body: #{response.body}"
  ].join("\n")
end
