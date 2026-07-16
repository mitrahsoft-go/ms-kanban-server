# Database configs
export DB_HOST=localhost
export DB_PORT="5432"
export DB_USERNAME= # Your Username 
export DB_PASSWORD= # Your Password 
export DB_NAME=kanban
export DB_AUTOMIGRATE=false

# HTTP configs
export HTTP_PORT=6369

#Logger configs
export LOGGER_TYPE=  # Production or Development
export LOGGER_LEVEL=debug #Info, Debug, Error, Warn, Fatal, Panic

# Redis configs
export REDIS_HOST=redis 
export REDIS_PORT=6379
export REDIS_PASSWORD= # Your Password 

# JWT Key
export JWT_SECRET_KEY = # Your Key 
export JWT_EXPIRY= # Set the duration (in Seconds)
export REFRESH_TOKEN_EXPIRY= # Set the duration (in Seconds)