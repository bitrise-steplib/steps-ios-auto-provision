def printable_response(response)
  str = "response\n"
  str += "status: #{response.code}\n"
  str += "body: #{response.body}\n"
  str
end
