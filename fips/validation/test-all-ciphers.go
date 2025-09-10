// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-all-ciphers.go <endpoint>")
		fmt.Println("Example: go run test-all-ciphers.go localhost:4318")
		os.Exit(1)
	}

	endpoint := os.Args[1]
	
	fmt.Printf("🔍 Testing All Cipher Suites Against %s
", endpoint)
	fmt.Println("=========================================")
	fmt.Println("(Certificate validation disabled for cipher testing)")
	fmt.Println()

	// Test basic connectivity first
	fmt.Println("🔗 Testing basic TLS connectivity...")
	defaultInfo, err := getServerDefaults(endpoint)
	if err != nil {
		fmt.Printf("❌ Cannot establish TLS connection: %v
", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Connection successful
")
	fmt.Printf("   Default cipher: %s
", defaultInfo.Cipher)
	fmt.Printf("   TLS version: %s
", defaultInfo.Version)
	fmt.Printf("   Certificate type: %s
", defaultInfo.CertType)
	fmt.Println()

	// Get all cipher suites and test them
	allCiphers := tls.CipherSuites()
	fmt.Printf("📋 Testing %d cipher suites (no certificate validation)...
", len(allCiphers))
	fmt.Println()

	var results []CipherResult
	
	for i, suite := range allCiphers {
		fmt.Printf("Testing %d/%d: %s... ", i+1, len(allCiphers), suite.Name)
		
		supported, tlsVersion := testCipherNoValidation(endpoint, suite.ID)
		result := CipherResult{
			Name:      suite.Name,
			ID:        suite.ID,
			Supported: supported,
			TLSVersion: tlsVersion,
			IsFIPS:    isFIPSCipher(suite.Name),
		}
		results = append(results, result)
		
		if supported {
			fmt.Printf("✅ WORKS (TLS %s)
", tlsVersion)
		} else {
			fmt.Println("❌ FAILS")
		}
	}

	// Analyze results
	analyzeResults(results)
}

type ServerInfo struct {
	Cipher   string
	Version  string
	CertType string
}

type CipherResult struct {
	Name       string
	ID         uint16
	Supported  bool
	TLSVersion string
	IsFIPS     bool
}

// Get server defaults without certificate validation
func getServerDefaults(endpoint string) (*ServerInfo, error) {
	config := &tls.Config{
		InsecureSkipVerify: true, // Skip ALL certificate validation
	}
	
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", endpoint, config)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	
	state := conn.ConnectionState()
	
	// Get certificate type from peer certificates
	certType := "Unknown"
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		switch cert.PublicKeyAlgorithm.String() {
		case "RSA":
			certType = "RSA"
		case "ECDSA":
			certType = "ECDSA" 
		case "Ed25519":
			certType = "Ed25519"
		default:
			certType = cert.PublicKeyAlgorithm.String()
		}
	}
	
	return &ServerInfo{
		Cipher:   getCipherName(state.CipherSuite),
		Version:  getTLSVersion(state.Version),
		CertType: certType,
	}, nil
}

// Test cipher without any certificate validation
func testCipherNoValidation(endpoint string, cipherID uint16) (bool, string) {
	// Try TLS 1.2 and 1.3
	versions := []struct {
		min, max uint16
		name     string
	}{
		{tls.VersionTLS12, tls.VersionTLS13, "1.2+"},
		{tls.VersionTLS13, tls.VersionTLS13, "1.3"},
		{tls.VersionTLS12, tls.VersionTLS12, "1.2"},
	}
	
	for _, v := range versions {
		config := &tls.Config{
			CipherSuites:       []uint16{cipherID},
			InsecureSkipVerify: true, // Skip ALL certificate validation
			MinVersion:         v.min,
			MaxVersion:         v.max,
		}

		dialer := &net.Dialer{Timeout: 2 * time.Second}
		conn, err := tls.DialWithDialer(dialer, "tcp", endpoint, config)
		if err != nil {
			continue // Try next version
		}
		defer conn.Close()

		// Verify the cipher was actually negotiated
		state := conn.ConnectionState()
		if state.CipherSuite == cipherID {
			return true, getTLSVersion(state.Version)
		}
	}
	
	return false, ""
}

// Analyze and display results
func analyzeResults(results []CipherResult) {
	var supported []CipherResult
	var fipsSupported []CipherResult
	var nonFipsSupported []CipherResult
	
	for _, result := range results {
		if result.Supported {
			supported = append(supported, result)
			if result.IsFIPS {
				fipsSupported = append(fipsSupported, result)
			} else {
				nonFipsSupported = append(nonFipsSupported, result)
			}
		}
	}
	
	fmt.Println()
	fmt.Println("🎯 Results Summary:")
	fmt.Println("===================")
	fmt.Printf("Total supported ciphers: %d
", len(supported))
	fmt.Printf("FIPS-approved ciphers: %d
", len(fipsSupported))
	fmt.Printf("Non-FIPS ciphers: %d
", len(nonFipsSupported))
	fmt.Println()

	if len(fipsSupported) > 0 {
		fmt.Println("✅ FIPS-Approved Ciphers:")
		fmt.Println("=========================")
		sort.Slice(fipsSupported, func(i, j int) bool {
			return fipsSupported[i].Name < fipsSupported[j].Name
		})
		
		for _, cipher := range fipsSupported {
			opensslName := convertToOpenSSL(cipher.Name)
			fmt.Printf("  - %s → %s (TLS %s)
", cipher.Name, opensslName, cipher.TLSVersion)
		}
		fmt.Println()

		fmt.Println("📋 OpenSSL format for validation:")
		fmt.Println("=================================")
		fmt.Println("fips_patterns=(")
		for _, cipher := range fipsSupported {
			opensslName := convertToOpenSSL(cipher.Name)
			fmt.Printf("    \"%s\"
", opensslName)
		}
		fmt.Println(")")
		fmt.Println()
	}

	if len(nonFipsSupported) > 0 {
		fmt.Println("⚠️  Non-FIPS Ciphers (Security Concern):")
		fmt.Println("========================================")
		for _, cipher := range nonFipsSupported {
			fmt.Printf("  - %s (TLS %s)
", cipher.Name, cipher.TLSVersion)
		}
		fmt.Println()
	}

	// FIPS compliance assessment
	if len(fipsSupported) > 0 {
		if len(nonFipsSupported) == 0 {
			fmt.Println("🎉 FULLY FIPS COMPLIANT")
			fmt.Println("  ✅ Only FIPS-approved ciphers are supported")
		} else {
			fmt.Println("⚠️  PARTIALLY FIPS COMPLIANT")
			fmt.Printf("  ✅ %d FIPS ciphers supported
", len(fipsSupported))
			fmt.Printf("  ⚠️  %d non-FIPS ciphers also supported
", len(nonFipsSupported))
		}
	} else {
		fmt.Println("❌ NOT FIPS COMPLIANT")
		fmt.Println("  ❌ No FIPS-approved ciphers are supported")
	}
}

// Check if cipher is FIPS-approved
func isFIPSCipher(name string) bool {
	// TLS 1.3 FIPS ciphers
	fips13 := []string{
		"TLS_AES_128_GCM_SHA256",
		"TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256",
	}
	
	for _, cipher := range fips13 {
		if name == cipher {
			return true
		}
	}
	
	// TLS 1.2 FIPS ciphers (GCM mode)
	if strings.Contains(name, "_GCM_") {
		if strings.Contains(name, "ECDHE_RSA") ||
		   strings.Contains(name, "ECDHE_ECDSA") ||
		   strings.Contains(name, "DHE_RSA") ||
		   strings.Contains(name, "RSA_WITH_AES") {
			return true
		}
	}
	
	return false
}

// Helper functions
func getCipherName(cipherID uint16) string {
	allSuites := append(tls.CipherSuites(), tls.InsecureCipherSuites()...)
	for _, suite := range allSuites {
		if suite.ID == cipherID {
			return suite.Name
		}
	}
	return fmt.Sprintf("Unknown(0x%04x)", cipherID)
}

func getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return fmt.Sprintf("Unknown(0x%04x)", version)
	}
}

func convertToOpenSSL(goName string) string {
	mappings := map[string]string{
		"TLS_AES_128_GCM_SHA256":                    "TLS_AES_128_GCM_SHA256",
		"TLS_AES_256_GCM_SHA384":                    "TLS_AES_256_GCM_SHA384",
		"TLS_CHACHA20_POLY1305_SHA256":              "TLS_CHACHA20_POLY1305_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":     "ECDHE-RSA-AES128-GCM-SHA256",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":     "ECDHE-RSA-AES256-GCM-SHA384",
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":   "ECDHE-ECDSA-AES128-GCM-SHA256",
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":   "ECDHE-ECDSA-AES256-GCM-SHA384",
		"TLS_DHE_RSA_WITH_AES_128_GCM_SHA256":       "DHE-RSA-AES128-GCM-SHA256",
		"TLS_DHE_RSA_WITH_AES_256_GCM_SHA384":       "DHE-RSA-AES256-GCM-SHA384",
		"TLS_RSA_WITH_AES_128_GCM_SHA256":           "AES128-GCM-SHA256",
		"TLS_RSA_WITH_AES_256_GCM_SHA384":           "AES256-GCM-SHA384",
	}
	
	if mapped, exists := mappings[goName]; exists {
		return mapped
	}
	
	result := strings.ReplaceAll(goName, "TLS_", "")
	result = strings.ReplaceAll(result, "_WITH_", "-")
	result = strings.ReplaceAll(result, "_", "-")
	return result
}
