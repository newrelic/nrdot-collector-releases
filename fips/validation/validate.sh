#!/bin/bash
# Copyright New Relic, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'

print_status() {
    local status="$1"
    local message="$2"
    
    case "$status" in
        "info")
            echo -e "${BLUE}ℹ️ ${message}${NC}"
            ;;
        "success")
            echo -e "${GREEN}✅ ${message}${NC}"
            ;;
        "warning")
            echo -e "${YELLOW}⚠️ ${message}${NC}"
            ;;
        "error")
            echo -e "${RED}❌ ${message}${NC}"
            ;;
    esac
}

is_cipher_fips_compliant() {
    local cipher="$1"
    
    # TLS 1.3 FIPS-approved ciphers (only specific ones, not all)
    # NIST SP 800-52 Rev. 2 and FIPS 140-2 IG specify these TLS 1.3 ciphers:
    if [[ "$cipher" =~ ^TLS_AES_(128|256)_GCM_SHA(256|384)$ ]] || 
       [[ "$cipher" =~ ^TLS_AES_(128|256)_CCM(_8)?_SHA256$ ]]; then
        return 0  # FIPS compliant TLS 1.3
    fi
    
    # TLS 1.2 FIPS-approved patterns
    [[ "$cipher" =~ ECDHE_(ECDSA|RSA).*AES_(128|256)_GCM_SHA(256|384) ]] ||
    [[ "$cipher" =~ DHE_RSA.*AES_(128|256)_GCM_SHA(256|384) ]] ||
    [[ "$cipher" =~ TLS_RSA_WITH_AES_(128|256)_GCM_SHA(256|384) ]] ||
    [[ "$cipher" =~ ^(AES_(128|256)_GCM|RSA.*AES_(128|256)_GCM) ]]
}

# Setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP_DIR="${GITHUB_WORKSPACE:-$SCRIPT_DIR}/.tmp"
CERT_DIR="$TMP_DIR/certs"
CONFIG_FILE="$TMP_DIR/fips-config.yaml"
CONTAINER_NAME="fips-test"
DOCKER_IMAGE="$1"

mkdir -p "$CERT_DIR"

print_status "info" "FIPS Validation for: $DOCKER_IMAGE"

# Generate certificates
create_certs() {
    [[ -f "$CERT_DIR/server.crt" ]] && return 0
    
    print_status "info" "Generating FIPS validation certificates..."
    cd "$CERT_DIR"
    
    # RSA certificate
    openssl genrsa -out server.key 2048 2>/dev/null
    openssl req -new -x509 -key server.key -out server.crt -days 365 -sha256 \
        -subj "/C=US/O=Test/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,DNS:host.docker.internal,IP:127.0.0.1" 2>/dev/null
    
    print_status "success" "Certificates generated"
}

# Create config
create_config() {
    cat > "$CONFIG_FILE" << EOF
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
        tls:
          cert_file: /certs/server.crt
          key_file: /certs/server.key
  hostmetrics:
    collection_interval: 1s
    scrapers:
      memory:

exporters:
  otlphttp:
    endpoint: https://localhost:8443
    tls:
      insecure_skip_verify: true

service:
  pipelines:
    metrics:
      receivers: [otlp, hostmetrics]
      exporters: [otlphttp]

EOF
}

# Start collector
start_collector() {
    # Clean up any existing containers and processes
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true

     # Fix file permissions for Docker container access
    chmod 644 "$CONFIG_FILE"
    chmod -R 644 "$CERT_DIR"/*
    chmod 755 "$CERT_DIR"
    
    docker run -d --name "$CONTAINER_NAME" --network host \
        -v "$CERT_DIR:/certs:ro" -v "$CONFIG_FILE:/config.yaml:ro" \
        "$DOCKER_IMAGE" --config=/config.yaml >/dev/null
    
    # Wait for ready
    for i in {1..15}; do
        if curl -k -s --max-time 1 https://localhost:4318/v1/metrics >/dev/null 2>&1; then
            print_status "success" "Collector ready (${i}s)"

            for i in {1..15}; do
                if docker logs "$CONTAINER_NAME" 2>&1 | grep -q "running in FIPS mode"; then
                    print_status "success" "BoringSSL FIPS module is enabled"

                    # Quick test to see what server negotiates by default
                    print_status "info" "Testing default server cipher negotiation..."
                    if openssl s_client -connect localhost:4318 -servername localhost < /dev/null > "$TMP_DIR/default-cipher.log" 2>&1; then
                        local default_cipher=$(grep "Cipher is" "$TMP_DIR/default-cipher.log" | awk '{print $NF}')
                        local default_protocol=$(grep "Protocol" "$TMP_DIR/default-cipher.log" | awk '{print $NF}')
                        print_status "info" "Default negotiation: $default_cipher ($default_protocol)"
                    fi
                    return 0
                fi
                sleep 1
            done

            print_status "error" "BoringSSL FIPS module is NOT enabled"
            return 0
        fi
        sleep 1
    done
    
    print_status "error" "Collector failed to start"
    docker logs "$CONTAINER_NAME" | tail -5
    exit 1
}

# Test server ciphers
test_server_ciphers() {
    print_status "info" "Testing server cipher suites..."
    
    local fips_count=0 non_fips_count=0 total_count=0
    local server_summary="$TMP_DIR/server-cipher-summary.log"
    
    # Initialize summary logs
    echo "📋 Server Cipher Suite Summary" > "$server_summary"
    echo "================================" >> "$server_summary"
    echo "Timestamp: $(date)" >> "$server_summary"
    echo "" >> "$server_summary"
    echo "✅ FIPS-Compliant Ciphers:" >> "$server_summary"
    
    local fips_ciphers=()
    local non_fips_ciphers=()
    
    # Get all available ciphers from OpenSSL
    local all_ciphers=$(openssl ciphers -stdname 'ALL:!aNULL:!eNULL:!EXPORT:!LOW' | tr ':' '\n')
    local cipher_count=$(echo "$all_ciphers" | wc -l)
    
    print_status "info" "Testing $cipher_count cipher suites..."

    local test_count=0
    while IFS= read -r cipher_line; do
        [[ -z "$cipher_line" ]] && continue
        ((test_count++))
        # Show progress every 20 ciphers
        if [[ $((test_count % 20)) -eq 0 ]]; then
            echo -n "."
        fi
        
        # Parse cipher line: "RFC_NAME - OPENSSL_NAME"
        local rfc_name=$(echo "$cipher_line" | awk '{print $1}')
        local openssl_name=$(echo "$cipher_line" | awk '{print $3}')
        
        [[ -z "$rfc_name" || -z "$openssl_name" ]] && continue

        # Create individual log file for this cipher test
        local cipher_log="${TMP_DIR}/cipher-4318-${rfc_name//[^a-zA-Z0-9]/_}.log"
        
        # Determine TLS version and parameters
        local tls_flag cipher_param cipher_value
        if [[ "$rfc_name" =~ ^TLS_(AES|CHACHA20) ]]; then
            tls_flag="-tls1_3"; cipher_param="-ciphersuites"; cipher_value="$rfc_name"
        elif [[ "$rfc_name" =~ ^TLS_ ]]; then
            tls_flag="-tls1_2"; cipher_param="-ciphersuites"; cipher_value="$rfc_name"
        else
            tls_flag="-tls1_2"; cipher_param="-cipher"; cipher_value="$openssl_name"
        fi
        
        # Test cipher with STRICT enforcement - no upgrades/downgrades allowed
        local cipher_restriction=""
        if [[ "$tls_flag" == "-tls1_3" ]]; then
            # For TLS 1.3: Use only -ciphersuites (no -cipher restriction needed)
            cipher_restriction="-ciphersuites ${cipher_value}"
        elif [[ "$cipher_param" == "-ciphersuites" ]]; then
            # For TLS 1.2 with -ciphersuites: Force ONLY this cipher suite
            cipher_restriction="-ciphersuites ${cipher_value} -cipher NONE"
        else
            # For legacy ciphers: Force ONLY this cipher
            cipher_restriction="-cipher ${cipher_value} -ciphersuites NONE"
        fi
        
        if timeout 10s openssl s_client -no_ticket -no_renegotiation -servername localhost $tls_flag \
           -connect localhost:4318 $cipher_restriction </dev/null >"$cipher_log" 2>&1; then
            
            local negotiated=$(grep "Cipher is" "$cipher_log" | awk '{print $NF}')            
            # Check if server used requested cipher (handle format differences)
            local cipher_match=false
            if [[ "$negotiated" == "$cipher_value" || "$negotiated" == "$rfc_name" || "$negotiated" == "$openssl_name" ]]; then
                cipher_match=true
            fi
            
            if [[ "$cipher_match" == "true" ]]; then
                ((total_count++))
                local tls_version=$(grep "Protocol" "$cipher_log" | awk '{print $NF}' | head -1)
                
                if is_cipher_fips_compliant "$rfc_name"; then
                    ((fips_count++))
                    fips_ciphers+=("  ✅ $rfc_name ($tls_version)")
                else
                    ((non_fips_count++))
                    non_fips_ciphers+=("  ⚠️  $rfc_name ($tls_version)")
                fi
            fi
        fi
    done <<< "$all_ciphers"
    echo ""
    
    # Write FIPS ciphers to summary
    for cipher in "${fips_ciphers[@]}"; do
        echo "$cipher" >> "$server_summary"
    done
    
    # Write non-FIPS ciphers to summary
    if [[ ${#non_fips_ciphers[@]} -gt 0 ]]; then
        echo "" >> "$server_summary"
        echo "⚠️  Non-FIPS Ciphers:" >> "$server_summary"
        for cipher in "${non_fips_ciphers[@]}"; do
            echo "$cipher" >> "$server_summary"
        done
    fi
    
    # Add summary statistics  
    echo "" >> "$server_summary"
    echo "📊 Server Summary Statistics:" >> "$server_summary"
    echo "  Total working ciphers: $total_count" >> "$server_summary"
    echo "  FIPS-compliant: $fips_count" >> "$server_summary" 
    echo "  Non-FIPS: $non_fips_count" >> "$server_summary"
    
    print_status "info" "Results: $total_count working, $fips_count FIPS, $non_fips_count non-FIPS"
    
    if [[ $fips_count -gt 0 && $non_fips_count -eq 0 ]]; then
        print_status "success" "Server: FULLY FIPS COMPLIANT"
        return 0
    elif [[ $fips_count -gt 0 ]]; then
        print_status "warning" "Server: PARTIALLY FIPS COMPLIANT"
        print_status "warning" "⚠️  Server accepts non-FIPS ciphers - this may indicate configuration issues"
        return 1
    else
        print_status "error" "Server: NOT FIPS COMPLIANT"
        print_status "error" "❌ No FIPS ciphers working - server configuration or connectivity issues"
        
        # Show collector logs for server cipher failures
        echo ""
        print_status "info" "📋 Recent collector logs (server cipher failures):"
        echo "========================================"
        docker logs "$CONTAINER_NAME" 2>&1 | tail -20 | sed 's/^/    /'
        echo "========================================"
        
        return 1
    fi
}

# Test client ciphers
test_client_ciphers() {
    print_status "info" "Testing client cipher suites (outbound connections)..."
    
    local server_log="$TMP_DIR/openssl-server.log"
    local client_summary="$TMP_DIR/client-cipher-summary.log"
    local server_port=8443
    
    # Check if we already have the pre-started OpenSSL server running
    local openssl_server_log="$TMP_DIR/openssl-server.log"
    if [[ -f "$openssl_server_log" ]]; then
        print_status "info" "Using pre-started OpenSSL server for client cipher capture"
        server_log="$openssl_server_log"
    else
        print_status "error" "OpenSSL server not started"
        exit 1
    fi
    
    print_status "info" "Waiting for collector to connect and complete TLS handshake..."
    print_status "info" "Collector should connect to https://localhost:${server_port}"
    
    # Wait for client connection and ClientHello
    local client_hello_captured=false
    for i in {1..30}; do
        if [[ -f "$server_log" ]] && grep -q "ClientHello" "$server_log" 2>/dev/null; then
            client_hello_captured=true
            print_status "success" "ClientHello captured! Analyzing cipher data..."
            break
        fi
        
        # Show progress every 5 seconds
        if [[ $((i % 5)) -eq 0 ]]; then
            print_status "info" "Still waiting for ClientHello... (${i}s elapsed)"
        fi
        sleep 1
    done
    
    if [[ "$client_hello_captured" != "true" ]]; then
        print_status "warning" "No ClientHello captured within 30 seconds"
        print_status "info" "This might indicate collector connection issues"
        
        if [[ -f "$server_log" ]]; then
            print_status "info" "OpenSSL server log analysis:"
            echo "  - Log file size: $(wc -l < "$server_log") lines"
            echo "  - Contains 'ACCEPT': $(grep -c "ACCEPT" "$server_log" 2>/dev/null || echo "0")"
            echo "  - Contains 'ClientHello': $(grep -c -i "client hello" "$server_log" 2>/dev/null || echo "0")"
            
            # Show last few lines for debugging
            print_status "info" "Last 10 lines of server log:"
            tail -10 "$server_log" | sed 's/^/    /'
        fi
        return 1
    fi
    
    # Initialize client summary
    echo "📋 Client Cipher Suite Summary" > "$client_summary"
    echo "================================" >> "$client_summary"
    echo "Timestamp: $(date)" >> "$client_summary"
    echo "" >> "$client_summary"
    
    # Analyze the captured TLS handshake using -trace output
    print_status "info" "📊 Analyzing captured client cipher suites..."
    
    
    # Parse cipher suites from OpenSSL trace output
    local fips_count=0 non_fips_count=0
    local fips_client_ciphers=()
    local non_fips_client_ciphers=()
    
    # Method 1: Look for explicit cipher suite names in trace output
    local all_ciphers=$(grep -i -A 100 "cipher_suites" "$server_log" 2>/dev/null | \
                        grep -E "(TLS_|ECDHE_|DHE_)" | \
                        grep -v "extensions\|compression\|session" | \
                        sed -E 's/.*[[:space:]]+(TLS_[A-Z0-9_]+).*/\1/' | \
                        sed -E 's/.*[[:space:]]+(ECDHE_[A-Z0-9_]+).*/\1/' | \
                        sed -E 's/.*[[:space:]]+(DHE_[A-Z0-9_]+).*/\1/' | \
                        sort -u)
    
    # Process found cipher names
    while IFS= read -r cipher; do
        [[ -z "$cipher" ]] && continue
        # Clean up cipher name
        cipher=$(echo "$cipher" | tr -d ',' | awk '{print $1}')
        [[ -z "$cipher" ]] && continue
        
        if is_cipher_fips_compliant "$cipher"; then
            ((fips_count++))
            fips_client_ciphers+=("  ✅ $cipher")
        else
            ((non_fips_count++))
            non_fips_client_ciphers+=("  ⚠️  $cipher")
        fi
    done <<< "$all_ciphers"
    
    # Write FIPS client ciphers
    echo "✅ FIPS-Compliant Ciphers Offered by Client:" >> "$client_summary"
    if [[ ${#fips_client_ciphers[@]} -gt 0 ]]; then
        for cipher in "${fips_client_ciphers[@]}"; do
            echo "$cipher" >> "$client_summary"
        done
    else
        echo "  (None)" >> "$client_summary"
    fi
    
    # Write non-FIPS client ciphers
    echo "" >> "$client_summary"
    echo "⚠️  Non-FIPS Ciphers Offered by Client:" >> "$client_summary"
    if [[ ${#non_fips_client_ciphers[@]} -gt 0 ]]; then
        for cipher in "${non_fips_client_ciphers[@]}"; do
            echo "$cipher" >> "$client_summary"
        done
    else
        echo "  (None)" >> "$client_summary"
    fi
    
    # Add client summary statistics
    echo "" >> "$client_summary"
    echo "📊 Client Summary Statistics:" >> "$client_summary"
    echo "  Total ciphers offered: $((fips_count + non_fips_count))" >> "$client_summary"
    echo "  FIPS-compliant: $fips_count" >> "$client_summary"
    echo "  Non-FIPS: $non_fips_count" >> "$client_summary"
    
    local total_client_ciphers=$((fips_count + non_fips_count))
    print_status "info" "Client offered: $fips_count FIPS, $non_fips_count non-FIPS (total: $total_client_ciphers)"
    
    # Check if client offered any ciphers at all
    if [[ $total_client_ciphers -eq 0 ]]; then
        print_status "error" "Client: No cipher suites offered - connection/configuration failure"
        print_status "error" "❌ This indicates the collector is not making outbound TLS connections"
        
        # Show recent collector logs for client connection debugging
        echo ""
        print_status "info" "📋 Recent collector logs (client connection debug):"
        echo "========================================"
        docker logs "$CONTAINER_NAME" 2>&1 | grep -i -E "(connect|dial|export|tls|error)" | tail -10 | sed 's/^/    /' || \
        docker logs "$CONTAINER_NAME" 2>&1 | tail -10 | sed 's/^/    /'
        echo "========================================"
        
        return 1
    elif [[ $fips_count -gt 0 && $non_fips_count -eq 0 ]]; then
        print_status "success" "Client: FIPS-only ciphers"
        return 0
    elif [[ $fips_count -gt 0 ]]; then
        print_status "warning" "Client: Mixed FIPS/non-FIPS ciphers"
        print_status "warning" "⚠️  Client offers non-FIPS ciphers alongside FIPS ones"
        return 1
    else
        print_status "error" "Client: Only non-FIPS ciphers offered"
        return 1
    fi
}

# Function to start OpenSSL test server for client cipher capture
start_openssl_server() {
    print_status "info" "🔍 Setting up OpenSSL test server for client cipher capture..."
    
    local server_port=8443
    local server_log="$TMP_DIR/openssl-server.log"
    
    # Use the existing server certificate
    cd "$CERT_DIR"
    
    # Kill any existing server on this port
    pkill -f "openssl s_server.*${server_port}" 2>/dev/null || true
    sleep 1
    
    # Start OpenSSL server with detailed tracing
    print_status "info" "Starting OpenSSL server on port ${server_port}..."
    openssl s_server \
        -accept 0.0.0.0:${server_port} \
        -cert server.crt \
        -key server.key \
        -trace \
        -state \
        -www \
        > "$server_log" 2>&1 &
    
    local server_pid=$!
    
    # Store server PID for cleanup
    echo "${server_pid}" > "$TMP_DIR/openssl-server.pid"
    
    print_status "info" "OpenSSL server started with PID ${server_pid}"
    
    # Wait for server to be ready
    local server_ready=false
    for i in {1..10}; do
        if timeout 1s bash -c "echo > /dev/tcp/localhost/${server_port}" 2>/dev/null; then
            server_ready=true
            print_status "success" "OpenSSL server ready on port ${server_port}"
            break
        fi
        sleep 1
    done
    
    if [[ "$server_ready" != "true" ]]; then
        print_status "error" "OpenSSL server failed to start"
        kill ${server_pid} 2>/dev/null || true
        return 1
    fi
    
    print_status "success" "OpenSSL test server ready to capture client connections"
    return 0
}

# Cleanup
cleanup() {
    print_status "info" "🧹 Cleaning up test resources..."
    
    # Stop and remove Docker container
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true
    
    # Kill OpenSSL test servers
    if [[ -f "$TMP_DIR/openssl-server.pid" ]]; then
        local server_pid=$(cat "$TMP_DIR/openssl-server.pid" 2>/dev/null)
        if [[ -n "$server_pid" ]]; then
            kill "$server_pid" 2>/dev/null || true
        fi
        rm -f "$TMP_DIR/openssl-server.pid"
    fi
    
    # Kill any OpenSSL servers on our test port
    pkill -f "openssl s_server.*8443" 2>/dev/null || true
    
}

# Main execution
main() {
    create_certs
    create_config
    
    # Start OpenSSL server BEFORE starting the collector
    print_status "info" "Setting up client cipher capture before starting collector..."
    if ! start_openssl_server; then
        print_status "warning" "OpenSSL server setup failed - client cipher testing will be skipped"
        print_status "info" "Continuing with server-side testing only"
    fi
    
    start_collector
    
    local server_ok=0 client_ok=0
    
    if test_server_ciphers; then
        server_ok=1
    fi
    
    # Give collector time to start making outbound connections
    print_status "info" "Allowing time for collector to establish outbound connections..."
    sleep 5
    
    if test_client_ciphers; then
        client_ok=1
    fi
    
    echo ""
    print_status "info" "Final FIPS Validation Results"
    echo "================================"
    
    if [[ $server_ok -eq 1 ]]; then
        print_status "success" "Server TLS: FIPS Compliant"
    else
        print_status "error" "Server TLS: Not FIPS Compliant"
    fi
    
    if [[ $client_ok -eq 1 ]]; then
        print_status "success" "Client TLS: FIPS Compliant"  
    else
        print_status "error" "Client TLS: Not FIPS Compliant"
    fi
    
    # Display summary files
    echo ""
    print_status "info" "📋 Detailed Cipher Summaries:"
    
    if [[ -f "$TMP_DIR/server-cipher-summary.log" ]]; then
        echo ""
        print_status "info" "Server Cipher Summary:"
        echo "=========================="
        cat "$TMP_DIR/server-cipher-summary.log"
        echo ""
    fi
    
    if [[ -f "$TMP_DIR/client-cipher-summary.log" ]]; then
        echo ""
        print_status "info" "Client Cipher Summary:"
        echo "=========================="
        cat "$TMP_DIR/client-cipher-summary.log"
        echo ""
    fi
    
    # Create combined summary
    echo ""
    if [[ -f "$TMP_DIR/server-cipher-summary.log" || -f "$TMP_DIR/client-cipher-summary.log" ]]; then
        local combined_summary="$TMP_DIR/fips-validation-complete-summary.log"
        echo "🎯 Complete FIPS Validation Summary" > "$combined_summary"
        echo "====================================" >> "$combined_summary"
        echo "Docker Image: $DOCKER_IMAGE" >> "$combined_summary"
        echo "Validation Date: $(date)" >> "$combined_summary"
        echo "Server Compliance: $(if [[ $server_ok -eq 1 ]]; then echo "✅ FIPS Compliant"; else echo "❌ Not FIPS Compliant"; fi)" >> "$combined_summary"
        echo "Client Compliance: $(if [[ $client_ok -eq 1 ]]; then echo "✅ FIPS Compliant"; else echo "⚠️  Validation Inconclusive"; fi)" >> "$combined_summary"
        echo "" >> "$combined_summary"
        
        if [[ -f "$TMP_DIR/server-cipher-summary.log" ]]; then
            cat "$TMP_DIR/server-cipher-summary.log" >> "$combined_summary"
            echo "" >> "$combined_summary"
        fi
        
        if [[ -f "$TMP_DIR/client-cipher-summary.log" ]]; then
            cat "$TMP_DIR/client-cipher-summary.log" >> "$combined_summary"
        fi

    fi
    
    if [[ $server_ok -eq 1 && $client_ok -eq 1 ]]; then
        print_status "success" "🎉 FIPS validation passed!"
        print_status "success" "✅ Both server and client TLS are FIPS compliant"
        exit 0
    else
        print_status "error" "❌ FIPS validation failed!"
        
        # Show which component failed
        if [[ $server_ok -ne 1 ]]; then
            print_status "error" "❌ Server TLS validation failed"
        fi
        if [[ $client_ok -ne 1 ]]; then
            print_status "error" "❌ Client TLS validation failed"
        fi
        
        # Dump collector logs for debugging
        echo ""
        print_status "info" "📋 Collector logs for debugging:"
        echo "========================================"
        docker logs "$CONTAINER_NAME" 2>&1 | tail -50
        echo "========================================"
        
        # Save full logs to file
        local debug_log="$TMP_DIR/collector-debug.log"
        docker logs "$CONTAINER_NAME" 2>&1 > "$debug_log"
        print_status "info" "Full collector logs saved to: $debug_log"
        
        exit 1
    fi
}

trap cleanup EXIT

main "$@"
