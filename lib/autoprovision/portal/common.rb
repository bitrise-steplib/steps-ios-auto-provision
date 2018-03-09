require 'spaceship'

def result_string(exception)
  result = exception.preferred_error_info
  return nil unless result
  result.join(' ')
end

def run_and_handle_portal_function
  yield
rescue Spaceship::Client::UnexpectedResponse => ex
  message = result_string(ex)
  raise ex unless message
  raise message
end
