#!/bin/bash

# Exit on error
set -e

# Setup Prometheus
setup_prometheus() {
    echo "Setting up Prometheus..."
    
    # Create necessary directories
    mkdir -p monitoring/prometheus
    
    # Download and extract Prometheus
    wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
    tar xvfz prometheus-*.tar.gz
    
    # Copy configuration
    cp deployment/monitoring/prometheus/prometheus.yml monitoring/prometheus/
    
    echo "Prometheus setup completed!"
}

# Setup Grafana
setup_grafana() {
    echo "Setting up Grafana..."
    
    # Create necessary directories
    mkdir -p monitoring/grafana/dashboards
    
    # Copy dashboards
    cp deployment/monitoring/grafana/dashboards/*.json monitoring/grafana/dashboards/
    
    # Set permissions
    chmod -R 777 monitoring/grafana
    
    echo "Grafana setup completed!"
}

# Main monitoring setup
main() {
    echo "Starting monitoring setup..."
    
    # Check if running in Kubernetes
    if [ -f /var/run/secrets/kubernetes.io/serviceaccount/token ]; then
        echo "Running in Kubernetes environment"
        kubectl apply -f deployment/monitoring/
    else
        echo "Running in standalone environment"
        setup_prometheus
        setup_grafana
    fi
    
    echo "Monitoring setup completed successfully!"
}

main