#!/bin/bash

# Exit on error
set -e

# Generate certificates
generate_certificates() {
    echo "Generating TLS certificates..."
    mkdir -p cert
    openssl req -x509 -newkey rsa:4096 \
        -keyout cert/server.key \
        -out cert/server.crt \
        -days 365 -nodes \
        -subj "/CN=proxy-service"
}

# Setup MongoDB
setup_mongodb() {
    echo "Setting up MongoDB..."
    mongo proxy_service --eval '
        db.createCollection("customers");
        db.createCollection("proxy_configs");
        db.createCollection("metrics");
        
        db.customers.createIndex({ "api_key": 1 }, { unique: true });
        db.proxy_configs.createIndex({ "customer_id": 1 });
        db.metrics.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 604800 });
    '
}

# Setup Redis
setup_redis() {
    echo "Setting up Redis..."
    redis-cli FLUSHDB
    redis-cli CONFIG SET maxmemory-policy allkeys-lru
}

# Main setup
main() {
    echo "Starting proxy service setup..."
    
    # Check dependencies
    command -v openssl >/dev/null 2>&1 || { echo "openssl is required"; exit 1; }
    command -v mongo >/dev/null 2>&1 || { echo "mongodb is required"; exit 1; }
    command -v redis-cli >/dev/null 2>&1 || { echo "redis is required"; exit 1; }
    
    generate_certificates
    setup_mongodb
    setup_redis
    
    echo "Setup completed successfully!"
}

main