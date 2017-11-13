def log_info(message)
  puts
  puts "\e[34m#{message}\e[0m"
end

def log_details(message)
  puts message.to_s
end

def log_done(message)
  puts "\e[32m#{message}\e[0m"
end

def log_warning(message)
  puts "\e[33m#{message}\e[0m"
end

def log_error(message)
  puts "\e[31m#{message}\e[0m"
end

def log_debug(message)
  return unless DEBUG_LOG
  puts message.to_s
end
