#!/bin/bash

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    local status=$1
    local message=$2
    case $status in
        "info")
            echo -e "${BLUE}ℹ️  ${message}${NC}"
            ;;
        "success")
            echo -e "${GREEN}✅ ${message}${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️  ${message}${NC}"
            ;;
        "error")
            echo -e "${RED}❌ ${message}${NC}"
            ;;
    esac
}

# In GitHub Actions/Act, use a path relative to workspace root
if [[ -n "${GITHUB_WORKSPACE:-}" ]]; then
    # Running in GitHub Actions or Act
    TMP_DIR="${GITHUB_WORKSPACE}/fips/validation/.tmp"
    print_status "info" "Running in CI/Act environment"
    print_status "info" "GitHub Workspace: ${GITHUB_WORKSPACE}"
else
    # Running locally
    TMP_DIR="${SCRIPT_DIR}/.tmp"
    print_status "info" "Running locally"
fi

# Set up directories relative to script location
mkdir -p "${TMP_DIR}"

print_status "info" "Working directory: $(pwd)"
print_status "info" "Script directory: ${SCRIPT_DIR}"
print_status "info" "Temp directory: ${TMP_DIR}"

CERT_DIR="${TMP_DIR}/certs"
CONFIG_FILE="${TMP_DIR}/fips-tls-config.yaml"
CONTAINER_NAME="fips-otel-collector-test"
# Create directories if they don't exist
mkdir -p "${CERT_DIR}"

# Use the Docker image directly as passed in
DOCKER_IMAGE="$1"

print_status "info" "Using Docker image: ${DOCKER_IMAGE}"

# Function to generate FIPS-compliant self-signed certificates
create_certs() {
    # Skip if certificates already exist
    if [[ -f "${CERT_DIR}/rsa-server.crt" && -f "${CERT_DIR}/rsa-server.key" && \
          -f "${CERT_DIR}/ecdsa-server.crt" && -f "${CERT_DIR}/ecdsa-server.key" ]]; then
        print_status "info" "Using existing certificates in ${CERT_DIR}"
        return 0
    fi
    
    print_status "info" "Generating FIPS-compliant self-signed certificates (RSA + ECDSA)..."
    
    cd "${CERT_DIR}"
    
    # ===================
    # Generate RSA Certificate
    # ===================
    print_status "info" "Generating RSA 2048-bit certificate..."
    
    # Generate RSA private key
    openssl genrsa -out rsa-server.key 2048
    
    # Create RSA certificate configuration
    cat > rsa-cert.conf << 'EOF'
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C=US
ST=California
L=San Francisco
O=New Relic
OU=FIPS Testing RSA
CN=localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment, digitalSignature
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
    
    # Generate RSA certificate
    openssl req -new -x509 -key rsa-server.key -out rsa-server.crt -days 365 \
        -config rsa-cert.conf -extensions v3_req -sha256
    
    # ===================
    # Generate ECDSA Certificate
    # ===================
    print_status "info" "Generating ECDSA P-256 certificate..."
    
    # Generate ECDSA private key using P-256 curve (FIPS-approved)
    openssl ecparam -genkey -name prime256v1 -out ecdsa-server.key
    
    # Create ECDSA certificate configuration
    cat > ecdsa-cert.conf << 'EOF'
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C=US
ST=California
L=San Francisco
O=New Relic
OU=FIPS Testing ECDSA
CN=localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment, digitalSignature
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
    
    # Generate ECDSA certificate
    openssl req -new -x509 -key ecdsa-server.key -out ecdsa-server.crt -days 365 \
        -config ecdsa-cert.conf -extensions v3_req -sha256
    
    # ===================
    # Verify certificates
    # ===================
    print_status "info" "Verifying generated certificates..."
    
    if openssl x509 -in rsa-server.crt -text -noout > /dev/null 2>&1; then
        print_status "success" "RSA certificate validation: ✅ PASS"
    else
        print_status "error" "RSA certificate validation failed"
        exit 1
    fi
    
    if openssl x509 -in ecdsa-server.crt -text -noout > /dev/null 2>&1; then
        print_status "success" "ECDSA certificate validation: ✅ PASS"
    else
        print_status "error" "ECDSA certificate validation failed"
        exit 1
    fi
    
    # Show certificate information
    print_status "info" "Certificate generation summary:"
    echo "  📋 RSA Certificate:"
    echo "    - Key: RSA 2048-bit"
    echo "    - Enables: ECDHE-RSA-*, DHE-RSA-*, AES*-GCM-* ciphers"
    echo "    - Files: rsa-server.crt, rsa-server.key"
    echo ""
    echo "  📋 ECDSA Certificate:"
    echo "    - Key: ECDSA P-256"
    echo "    - Enables: ECDHE-ECDSA-* ciphers"
    echo "    - Files: ecdsa-server.crt, ecdsa-server.key"
    echo ""
    
    print_status "success" "All FIPS-compliant certificates generated successfully"
}

# Function to create FIPS TLS configuration
create_fips_tls_config() {
    print_status "info" "Creating FIPS TLS configuration for Docker..."
    
    # Create configuration using RSA certificate (default)
    cat > "${CONFIG_FILE}" << EOF
receivers:
  # OTLP HTTP with RSA certificate (primary endpoint)
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
        tls:
          cert_file: /certs/rsa-server.crt
          key_file: /certs/rsa-server.key
          
  # OTLP HTTP with ECDSA certificate (secondary endpoint for testing)
  otlp/ecdsa:
    protocols:
      http:
        endpoint: 0.0.0.0:4319
        tls:
          cert_file: /certs/ecdsa-server.crt
          key_file: /certs/ecdsa-server.key

processors:
  batch:
    send_batch_size: 1024

exporters:
  debug:
  
  # Export internal telemetry to Docker host
  otlphttp/internal:
    endpoint: https://host.docker.internal:8080/v1/metrics
    timeout: 10s
    retry_on_failure:
      enabled: true
      initial_interval: 1s
      max_interval: 30s
      max_elapsed_time: 300s

service:
  telemetry:
    logs:
      level: info
    metrics:
      # Export internal metrics to Docker host
      level: detailed
      readers:
        - periodic:
            interval: 10000
            exporter:
              otlp:
                protocol: http/protobuf
                endpoint: https://host.docker.internal:8080/v1/metrics
                timeout: 10000
                
  pipelines:
    traces:
      receivers: [otlp, otlp/ecdsa]
      processors: [batch]
      exporters: [debug]
    metrics:
      receivers: [otlp, otlp/ecdsa]
      processors: [batch]
      exporters: [debug, otlphttp/internal]
    logs:
      receivers: [otlp, otlp/ecdsa]
      processors: [batch]
      exporters: [debug]
EOF
    
    print_status "success" "Docker FIPS TLS configuration created: ${CONFIG_FILE}"
    print_status "info" "Configured endpoints:"
    echo "  🔐 Port 4318: RSA certificate (ECDHE-RSA-*, DHE-RSA-*, AES*-GCM-*)"
    echo "  🔐 Port 4319: ECDSA certificate (ECDHE-ECDSA-*)"
    print_status "info" "Internal telemetry will be exported to host.docker.internal:8080"
}

# Function to start Docker container
start_docker_collector() {
    print_status "info" "Starting FIPS collector in Docker..."
    
    # Stop existing container if running
    if docker ps -a --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_status "info" "Stopping existing container: ${CONTAINER_NAME}"
        docker stop "${CONTAINER_NAME}" >/dev/null 2>&1 || true
        docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true
    fi
    
    # Start new container
    print_status "info" "Starting container with image: ${DOCKER_IMAGE}"
    print_status "info" "Config file: ${CONFIG_FILE}"
    print_status "info" "Config file exists: $(test -f "${CONFIG_FILE}" && echo "yes" || echo "no")"
    print_status "info" "Config file size: $(wc -c < "${CONFIG_FILE}") bytes"
    
    local container_id
    container_id=$(docker run -d \
        --platform linux/amd64 \
        --name "${CONTAINER_NAME}" \
        -p 4318:4318 \
        -p 4319:4319 \
        -v "${CERT_DIR}:/certs:ro" \
        -v "${CONFIG_FILE}:/tmp/config.yaml:ro" \
        "${DOCKER_IMAGE}" \
        --config=/tmp/config.yaml)
    
    print_status "info" "Container started with ID: ${container_id:0:12}"
    
    # Wait for container to be ready
    print_status "info" "Waiting for container to start up..."
    
    local container_ready=false
    for i in {1..30}; do
        sleep 1
        
        # Check if container is still running
        if ! docker ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
            print_status "error" "Container stopped unexpectedly"
            print_status "info" "Container logs:"
            docker logs "${CONTAINER_NAME}" 2>&1 | tail -20
            exit 1
        fi
        
        # Check if port is accessible
        if timeout 2s bash -c "echo > /dev/tcp/localhost/4318" 2>/dev/null; then
            container_ready=true
            print_status "success" "Container ready and port 4318 accessible (after ${i}s)"
            break
        fi
        
        # Show progress every 5 seconds
        if [[ $((i % 5)) -eq 0 ]]; then
            print_status "info" "Still waiting for container to be ready... (${i}s elapsed)"
            docker logs "${CONTAINER_NAME}" --tail=3 2>/dev/null | grep -E "(Starting|server|HTTP|TLS|error|fail)" || true
        fi
    done
    
    if [[ "$container_ready" != "true" ]]; then
        print_status "error" "Container started but port 4318 not accessible after 30s"
        print_status "info" "Container logs:"
        docker logs "${CONTAINER_NAME}" 2>&1 | tail -20
        exit 1
    fi
    
    # Show container info
    print_status "info" "Container information:"
    echo "  Name: ${CONTAINER_NAME}"
    echo "  ID: ${container_id:0:12}"
    echo "  Port: 4318 (HTTPS)"
    echo "  Image: ${DOCKER_IMAGE}"
    
    # Check for FIPS validation in logs
    local container_logs
    container_logs=$(docker logs "${CONTAINER_NAME}" 2>&1)
    if echo "$container_logs" | grep -q "FIPS MODE"; then
        print_status "success" "FIPS mode validation found in container logs"
        echo "$container_logs" | grep "FIPS" | head -3
    else
        print_status "warning" "FIPS validation messages not found in container logs"
    fi
    
    echo ""
    print_status "success" "🎉 Docker collector started successfully!"
    echo ""
}



# Function to check server cipher suites
check_server_ciphers() {
    print_status "info" "🔐 Testing FIPS cipher suites against multiple certificate endpoints"
    echo ""
    
    # Test RSA endpoint (port 4318)
    print_status "info" "🔍 Testing RSA certificate endpoint (localhost:4318)..."
    local rsa_results
    rsa_results=$(cd "${SCRIPT_DIR}" && CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go run -tags=boringcrypto test-all-ciphers.go localhost:4318 2>/dev/null || true)
    
    # Test ECDSA endpoint (port 4319) 
    print_status "info" "🔍 Testing ECDSA certificate endpoint (localhost:4319)..."
    local ecdsa_results
    ecdsa_results=$(cd "${SCRIPT_DIR}" && CGO_ENABLED=1 GOEXPERIMENT=boringcrypto go run -tags=boringcrypto test-all-ciphers.go localhost:4319 2>/dev/null || true)
    
    # Analyze combined results
    print_status "info" "📊 Combined Cipher Support Analysis:"
    echo "=================================="
    echo ""
    
    if [[ -n "$rsa_results" ]]; then
        echo "🔐 RSA Certificate Endpoint (Port 4318):"
        echo "========================================="
        echo "$rsa_results" | grep -E "(Testing|WORKS|FAILS|supported ciphers:|FIPS-approved|Non-FIPS|COMPLIANT)" | head -20
        echo ""
        
        # Extract RSA counts
        local rsa_total=$(echo "$rsa_results" | grep "Total supported ciphers:" | sed 's/.*: //' || echo "0")
        local rsa_fips=$(echo "$rsa_results" | grep "FIPS-approved ciphers:" | sed 's/.*: //' || echo "0")
        local rsa_non_fips=$(echo "$rsa_results" | grep "Non-FIPS ciphers:" | sed 's/.*: //' || echo "0")
        
        print_status "info" "RSA Endpoint: ${rsa_fips} FIPS ciphers, ${rsa_non_fips} non-FIPS ciphers"
    else
        print_status "warning" "RSA endpoint testing failed"
        rsa_total=0; rsa_fips=0; rsa_non_fips=0
    fi
    
    if [[ -n "$ecdsa_results" ]]; then
        echo "🔐 ECDSA Certificate Endpoint (Port 4319):"
        echo "==========================================="
        echo "$ecdsa_results" | grep -E "(Testing|WORKS|FAILS|supported ciphers:|FIPS-approved|Non-FIPS|COMPLIANT)" | head -20
        echo ""
        
        # Extract ECDSA counts
        local ecdsa_total=$(echo "$ecdsa_results" | grep "Total supported ciphers:" | sed 's/.*: //' || echo "0")
        local ecdsa_fips=$(echo "$ecdsa_results" | grep "FIPS-approved ciphers:" | sed 's/.*: //' || echo "0")
        local ecdsa_non_fips=$(echo "$ecdsa_results" | grep "Non-FIPS ciphers:" | sed 's/.*: //' || echo "0")
        
        print_status "info" "ECDSA Endpoint: ${ecdsa_fips} FIPS ciphers, ${ecdsa_non_fips} non-FIPS ciphers"
    else
        print_status "warning" "ECDSA endpoint testing failed"
        ecdsa_total=0; ecdsa_fips=0; ecdsa_non_fips=0
    fi
    
    # Calculate combined results
    local combined_fips=$((rsa_fips + ecdsa_fips))
    local combined_non_fips=$((rsa_non_fips + ecdsa_non_fips))
    local combined_total=$((rsa_total + ecdsa_total))
    
    echo ""
    print_status "info" "🎯 Overall Container Cipher Capability:"
    echo "========================================"
    echo "  📊 Total unique cipher combinations: ${combined_total}"
    echo "  ✅ FIPS-approved cipher combinations: ${combined_fips}" 
    echo "  ⚠️  Non-FIPS cipher combinations: ${combined_non_fips}"
    echo ""
    
    # Determine overall compliance
    if [[ ${combined_fips} -gt 0 ]]; then
        if [[ ${combined_non_fips} -eq 0 ]]; then
            print_status "success" "🎉 FULLY FIPS COMPLIANT!"
            print_status "info" "Container supports ${combined_fips} FIPS cipher combinations across both certificate types"
            print_status "info" "No non-FIPS ciphers detected - excellent security posture"
            return 0
        else
            print_status "warning" "⚠️  PARTIALLY FIPS COMPLIANT"
            print_status "info" "Container supports ${combined_fips} FIPS ciphers, but also ${combined_non_fips} non-FIPS ciphers"
            print_status "warning" "Non-FIPS cipher support indicates the collector may not be in strict FIPS mode"
            return 1
        fi
    else
        print_status "error" "❌ NOT FIPS COMPLIANT"
        print_status "error" "No FIPS cipher suites are working on either endpoint"
        return 1
    fi
}

# Function to cleanup on script exit
cleanup() {
    if docker ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$" 2>/dev/null; then
        print_status "info" "Container ${CONTAINER_NAME} is still running"
    fi
}

# Main execution
main() {
    print_status "info" "Starting Docker FIPS collector setup..."
    echo ""
    
    create_certs
    create_fips_tls_config
    start_docker_collector
    
    # Wait a moment for TLS server to be fully ready
    print_status "info" "Waiting for TLS server to be fully ready..."
    sleep 3
    
    # Test FIPS cipher compliance
    if check_server_ciphers; then
        print_status "success" "🎉 FIPS validation completed successfully!"
        exit 0
    else
        print_status "error" "❌ FIPS validation failed!"
        
        # Show container logs for debugging
        print_status "info" "Container logs for debugging:"
        docker logs "${CONTAINER_NAME}" 2>&1 | tail -20
        
        exit 1
    fi
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"
