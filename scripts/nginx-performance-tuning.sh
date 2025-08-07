#!/bin/bash

# Nginx Performance Tuning Script
# This script optimizes nginx configuration for production workloads

set -e

# Configuration
NGINX_CONTAINER=${NGINX_CONTAINER:-"bookmark-nginx"}
WORKER_PROCESSES=${WORKER_PROCESSES:-"auto"}
WORKER_CONNECTIONS=${WORKER_CONNECTIONS:-"2048"}
KEEPALIVE_TIMEOUT=${KEEPALIVE_TIMEOUT:-"65"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Detect system resources
detect_system_resources() {
    log "Detecting system resources..."

    # Get CPU cores
    local cpu_cores=$(nproc)
    log "CPU cores: $cpu_cores"

    # Get memory
    local memory_gb=$(free -g | awk '/^Mem:/{print $2}')
    log "Memory: ${memory_gb}GB"

    # Calculate optimal worker processes
    if [[ "$WORKER_PROCESSES" == "auto" ]]; then
        WORKER_PROCESSES=$cpu_cores
        log "Setting worker_processes to: $WORKER_PROCESSES"
    fi

    # Calculate optimal worker connections
    local max_connections=$((WORKER_PROCESSES * WORKER_CONNECTIONS))
    log "Maximum connections: $max_connections"

    # Check system limits
    local ulimit_n=$(ulimit -n)
    if [[ $max_connections -gt $ulimit_n ]]; then
        warn "Maximum connections ($max_connections) exceeds ulimit -n ($ulimit_n)"
        warn "Consider increasing system limits"
    fi
}

# Generate optimized nginx configuration
generate_optimized_config() {
    log "Generating optimized nginx configuration..."

    cat > nginx/nginx.optimized.conf << EOF
user nginx;
worker_processes $WORKER_PROCESSES;
worker_rlimit_nofile 65535;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections $WORKER_CONNECTIONS;
    use epoll;
    multi_accept on;
    accept_mutex off;
}

http {
    include /etc/nginx/mime.types;
    include /etc/nginx/conf.d/*.conf;
    default_type application/octet-stream;

    # Logging format with performance metrics
    log_format main '\$remote_addr - \$remote_user [\$time_local] "\$request" '
                    '\$status \$body_bytes_sent "\$http_referer" '
                    '"\$http_user_agent" "\$http_x_forwarded_for" '
                    'rt=\$request_time uct="\$upstream_connect_time" '
                    'uht="\$upstream_header_time" urt="\$upstream_response_time" '
                    'cs=\$upstream_cache_status';

    access_log /var/log/nginx/access.log main buffer=64k flush=5s;

    # Performance optimizations
    sendfile on;
    sendfile_max_chunk 1m;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout $KEEPALIVE_TIMEOUT;
    keepalive_requests 1000;
    types_hash_max_size 2048;
    server_names_hash_bucket_size 128;
    server_names_hash_max_size 1024;

    # Buffer sizes
    client_body_buffer_size 128k;
    client_max_body_size 100m;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 4k;
    output_buffers 1 32k;
    postpone_output 1460;

    # Timeouts
    client_body_timeout 12;
    client_header_timeout 12;
    send_timeout 10;

    # Hide nginx version
    server_tokens off;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml
        application/x-font-ttf
        application/vnd.ms-fontobject
        font/opentype;

    # Open file cache
    open_file_cache max=10000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;
    open_file_cache_errors on;

    # Include server configurations
    include /etc/nginx/sites-enabled/*;
}
EOF

    log "Optimized configuration generated: nginx/nginx.optimized.conf"
}

# Create performance monitoring configuration
create_monitoring_config() {
    log "Creating performance monitoring configuration..."

    cat > nginx/conf.d/monitoring.conf << 'EOF'
# Performance monitoring endpoints

# Nginx status endpoint (requires --with-http_stub_status_module)
server {
    listen 127.0.0.1:8080;
    server_name localhost;

    location /nginx_status {
        stub_status on;
        access_log off;
        allow 127.0.0.1;
        allow ::1;
        deny all;
    }

    location /nginx_metrics {
        access_log off;
        allow 127.0.0.1;
        allow ::1;
        deny all;

        content_by_lua_block {
            local status = ngx.location.capture("/nginx_status")
            if status.status == 200 then
                local body = status.body
                -- Parse nginx status and convert to Prometheus format
                local active = body:match("Active connections: (%d+)")
                local accepts, handled, requests = body:match("(%d+) (%d+) (%d+)")
                local reading, writing, waiting = body:match("Reading: (%d+) Writing: (%d+) Waiting: (%d+)")

                ngx.header.content_type = "text/plain"
                ngx.say("# HELP nginx_connections_active Active connections")
                ngx.say("# TYPE nginx_connections_active gauge")
                ngx.say("nginx_connections_active " .. (active or 0))

                ngx.say("# HELP nginx_connections_reading Reading connections")
                ngx.say("# TYPE nginx_connections_reading gauge")
                ngx.say("nginx_connections_reading " .. (reading or 0))

                ngx.say("# HELP nginx_connections_writing Writing connections")
                ngx.say("# TYPE nginx_connections_writing gauge")
                ngx.say("nginx_connections_writing " .. (writing or 0))

                ngx.say("# HELP nginx_connections_waiting Waiting connections")
                ngx.say("# TYPE nginx_connections_waiting gauge")
                ngx.say("nginx_connections_waiting " .. (waiting or 0))
            else
                ngx.status = 500
                ngx.say("Error getting nginx status")
            end
        }
    }
}
EOF

    log "Monitoring configuration created: nginx/conf.d/monitoring.conf"
}

# Benchmark nginx performance
benchmark_performance() {
    log "Running performance benchmark..."

    # Check if ab (Apache Bench) is available
    if ! command -v ab &> /dev/null; then
        warn "Apache Bench (ab) not found. Installing..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install httpd
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            if command -v apt-get &> /dev/null; then
                sudo apt-get update && sudo apt-get install -y apache2-utils
            elif command -v yum &> /dev/null; then
                sudo yum install -y httpd-tools
            fi
        fi
    fi

    if command -v ab &> /dev/null; then
        log "Running benchmark with 1000 requests, 10 concurrent connections..."
        ab -n 1000 -c 10 -g benchmark.tsv http://localhost/health > benchmark_results.txt 2>&1

        # Extract key metrics
        local requests_per_second=$(grep "Requests per second" benchmark_results.txt | awk '{print $4}')
        local time_per_request=$(grep "Time per request" benchmark_results.txt | head -1 | awk '{print $4}')
        local failed_requests=$(grep "Failed requests" benchmark_results.txt | awk '{print $3}')

        log "Benchmark results:"
        log "  Requests per second: $requests_per_second"
        log "  Time per request: ${time_per_request}ms"
        log "  Failed requests: $failed_requests"

        # Save results
        echo "Benchmark completed at $(date)" >> logs/performance-history.log
        echo "Requests per second: $requests_per_second" >> logs/performance-history.log
        echo "Time per request: ${time_per_request}ms" >> logs/performance-history.log
        echo "Failed requests: $failed_requests" >> logs/performance-history.log
        echo "---" >> logs/performance-history.log
    else
        warn "Could not install Apache Bench. Skipping benchmark."
    fi
}

# Optimize system settings
optimize_system_settings() {
    log "Checking system settings for optimization..."

    # Check current limits
    log "Current system limits:"
    log "  ulimit -n (open files): $(ulimit -n)"
    log "  ulimit -u (processes): $(ulimit -u)"

    # Suggest optimizations
    log "Recommended system optimizations:"
    log "  1. Increase open file limit: ulimit -n 65535"
    log "  2. Optimize kernel parameters in /etc/sysctl.conf:"
    log "     net.core.somaxconn = 65535"
    log "     net.core.netdev_max_backlog = 5000"
    log "     net.ipv4.tcp_max_syn_backlog = 65535"
    log "     net.ipv4.tcp_fin_timeout = 10"
    log "     net.ipv4.tcp_keepalive_time = 600"
    log "     net.ipv4.tcp_keepalive_intvl = 60"
    log "     net.ipv4.tcp_keepalive_probes = 10"

    # Create sysctl configuration
    cat > system-optimizations.conf << 'EOF'
# System optimizations for nginx
# Add these to /etc/sysctl.conf and run: sysctl -p

# Network optimizations
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 10
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_intvl = 60
net.ipv4.tcp_keepalive_probes = 10
net.ipv4.tcp_tw_reuse = 1

# Memory optimizations
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5

# File system optimizations
fs.file-max = 2097152
EOF

    log "System optimization suggestions saved to: system-optimizations.conf"
}

# Generate performance report
generate_performance_report() {
    log "Generating performance report..."

    local report_file="logs/nginx-performance-report-$(date +%Y%m%d-%H%M%S).json"
    mkdir -p logs

    cat > "$report_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "system_info": {
        "cpu_cores": $(nproc),
        "memory_gb": $(free -g | awk '/^Mem:/{print $2}'),
        "ulimit_n": $(ulimit -n),
        "ulimit_u": $(ulimit -u)
    },
    "nginx_config": {
        "worker_processes": "$WORKER_PROCESSES",
        "worker_connections": "$WORKER_CONNECTIONS",
        "keepalive_timeout": "$KEEPALIVE_TIMEOUT",
        "max_connections": $((WORKER_PROCESSES * WORKER_CONNECTIONS))
    },
    "optimizations_applied": [
        "worker_processes optimized",
        "worker_connections optimized",
        "gzip compression enabled",
        "open_file_cache enabled",
        "sendfile enabled",
        "tcp_nopush enabled",
        "tcp_nodelay enabled"
    ]
}
EOF

    log "Performance report generated: $report_file"
}

# Main execution
main() {
    log "Starting nginx performance tuning..."

    # Create logs directory
    mkdir -p logs

    case "${1:-optimize}" in
        "optimize")
            detect_system_resources
            generate_optimized_config
            create_monitoring_config
            optimize_system_settings
            generate_performance_report
            log "Performance tuning completed!"
            log "To apply optimizations:"
            log "  1. Review nginx/nginx.optimized.conf"
            log "  2. Replace current configuration if satisfied"
            log "  3. Apply system optimizations from system-optimizations.conf"
            log "  4. Restart nginx: docker-compose restart nginx"
            ;;
        "benchmark")
            benchmark_performance
            ;;
        "monitor")
            create_monitoring_config
            log "Monitoring configuration created"
            log "Access nginx status at: http://localhost:8080/nginx_status"
            ;;
        "report")
            generate_performance_report
            ;;
        *)
            echo "Usage: $0 [optimize|benchmark|monitor|report]"
            echo "  optimize  - Generate optimized configuration (default)"
            echo "  benchmark - Run performance benchmark"
            echo "  monitor   - Setup monitoring endpoints"
            echo "  report    - Generate performance report"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"