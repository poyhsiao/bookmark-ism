#!/bin/bash

# Nginx Health Check and Monitoring Script
# This script monitors nginx health and upstream servers

set -e

# Configuration
NGINX_CONTAINER=${NGINX_CONTAINER:-"bookmark-nginx"}
CHECK_INTERVAL=${CHECK_INTERVAL:-30}
LOG_FILE=${LOG_FILE:-"logs/nginx-health.log"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}" | tee -a "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" | tee -a "$LOG_FILE"
}

# Create logs directory
mkdir -p logs

# Check nginx container status
check_nginx_container() {
    if docker ps --filter "name=$NGINX_CONTAINER" --filter "status=running" | grep -q "$NGINX_CONTAINER"; then
        log "Nginx container is running"
        return 0
    else
        error "Nginx container is not running"
        return 1
    fi
}

# Check nginx configuration
check_nginx_config() {
    if docker exec "$NGINX_CONTAINER" nginx -t > /dev/null 2>&1; then
        log "Nginx configuration is valid"
        return 0
    else
        error "Nginx configuration is invalid"
        docker exec "$NGINX_CONTAINER" nginx -t 2>&1 | tee -a "$LOG_FILE"
        return 1
    fi
}

# Check nginx health endpoint
check_health_endpoint() {
    local url="http://localhost/health"

    if curl -f -s "$url" > /dev/null 2>&1; then
        log "Health endpoint is responding"
        return 0
    else
        error "Health endpoint is not responding"
        return 1
    fi
}

# Check upstream servers
check_upstream_servers() {
    log "Checking upstream server status..."

    # Check API backend
    if curl -f -s "http://localhost/api/v1/health" > /dev/null 2>&1; then
        log "API backend is healthy"
    else
        warn "API backend is not responding"
    fi

    # Check Supabase Auth
    if curl -f -s "http://localhost/auth/health" > /dev/null 2>&1; then
        log "Supabase Auth is healthy"
    else
        warn "Supabase Auth is not responding"
    fi

    # Check Supabase REST
    if curl -f -s "http://localhost/rest/" > /dev/null 2>&1; then
        log "Supabase REST is healthy"
    else
        warn "Supabase REST is not responding"
    fi
}

# Check SSL certificates
check_ssl_certificates() {
    if [[ -f "nginx/ssl/cert.pem" ]]; then
        local expiry_date=$(openssl x509 -enddate -noout -in nginx/ssl/cert.pem | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s)
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))

        if [[ $days_until_expiry -gt 30 ]]; then
            log "SSL certificate is valid for $days_until_expiry more days"
        elif [[ $days_until_expiry -gt 7 ]]; then
            warn "SSL certificate expires in $days_until_expiry days"
        else
            error "SSL certificate expires in $days_until_expiry days - renewal required!"
        fi
    else
        warn "SSL certificate not found"
    fi
}

# Check nginx metrics
check_nginx_metrics() {
    log "Checking nginx metrics..."

    # Get nginx status if stub_status is enabled
    if docker exec "$NGINX_CONTAINER" curl -s "http://localhost/nginx_status" > /dev/null 2>&1; then
        local status=$(docker exec "$NGINX_CONTAINER" curl -s "http://localhost/nginx_status")
        log "Nginx status: $status"
    fi

    # Check error log for recent errors
    local error_count=$(docker exec "$NGINX_CONTAINER" tail -n 100 /var/log/nginx/error.log 2>/dev/null | grep "$(date '+%Y/%m/%d')" | wc -l || echo "0")
    if [[ $error_count -gt 0 ]]; then
        warn "Found $error_count errors in today's nginx error log"
    else
        log "No errors found in today's nginx error log"
    fi
}

# Check rate limiting
check_rate_limiting() {
    log "Testing rate limiting..."

    # Test API rate limiting
    local api_responses=0
    for i in {1..25}; do
        if curl -f -s "http://localhost/api/v1/health" > /dev/null 2>&1; then
            ((api_responses++))
        fi
    done

    if [[ $api_responses -lt 25 ]]; then
        log "Rate limiting is working - $api_responses/25 requests succeeded"
    else
        warn "Rate limiting may not be working properly - all 25 requests succeeded"
    fi
}

# Generate health report
generate_health_report() {
    local report_file="logs/nginx-health-report-$(date +%Y%m%d-%H%M%S).json"

    cat > "$report_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "nginx_container_running": $(check_nginx_container && echo "true" || echo "false"),
    "nginx_config_valid": $(check_nginx_config && echo "true" || echo "false"),
    "health_endpoint_responding": $(check_health_endpoint && echo "true" || echo "false"),
    "ssl_certificate_valid": $(check_ssl_certificates && echo "true" || echo "false"),
    "upstream_servers": {
        "api_backend": $(curl -f -s "http://localhost/api/v1/health" > /dev/null 2>&1 && echo "true" || echo "false"),
        "supabase_auth": $(curl -f -s "http://localhost/auth/health" > /dev/null 2>&1 && echo "true" || echo "false"),
        "supabase_rest": $(curl -f -s "http://localhost/rest/" > /dev/null 2>&1 && echo "true" || echo "false")
    }
}
EOF

    log "Health report generated: $report_file"
}

# Main health check function
run_health_check() {
    log "Starting nginx health check..."

    local overall_status=0

    # Run all checks
    check_nginx_container || overall_status=1
    check_nginx_config || overall_status=1
    check_health_endpoint || overall_status=1
    check_upstream_servers
    check_ssl_certificates
    check_nginx_metrics
    check_rate_limiting

    # Generate report
    generate_health_report

    if [[ $overall_status -eq 0 ]]; then
        log "All critical health checks passed"
    else
        error "Some critical health checks failed"
    fi

    return $overall_status
}

# Continuous monitoring mode
monitor_mode() {
    log "Starting continuous monitoring mode (interval: ${CHECK_INTERVAL}s)"

    while true; do
        run_health_check
        sleep "$CHECK_INTERVAL"
    done
}

# Main execution
main() {
    case "${1:-check}" in
        "check")
            run_health_check
            ;;
        "monitor")
            monitor_mode
            ;;
        "report")
            generate_health_report
            ;;
        *)
            echo "Usage: $0 [check|monitor|report]"
            echo "  check   - Run health check once (default)"
            echo "  monitor - Run continuous monitoring"
            echo "  report  - Generate health report only"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"