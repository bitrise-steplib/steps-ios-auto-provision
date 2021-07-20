require 'spaceship'

def result_string(ex)
  ex.preferred_error_info&.join(' ') || ex.to_s
end

def run_and_handle_portal_function
  yield
rescue Spaceship::Client::UnexpectedResponse => ex
  raise result_string(ex)
end
