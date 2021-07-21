require 'spaceship'

def preferred_error_message(exception)
  exception.preferred_error_info&.join(' ') || exception.to_s
end

def run_or_raise_preferred_error_message
  yield
rescue Spaceship::Client::UnexpectedResponse => e
  raise preferred_error_message(e)
end
