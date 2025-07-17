// Copyright 2024 The KitOps Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"time"
)

// OptimizedTransport creates a Docker CLI-inspired HTTP transport with optimizations
// for container registry operations
func OptimizedTransport(tlsConfig *tls.Config) *http.Transport {
	// Base transport with Docker CLI-inspired settings
	transport := &http.Transport{
		// Connection pooling optimizations
		MaxIdleConns:        100,              // Docker CLI uses 100
		MaxIdleConnsPerHost: 10,               // Docker CLI uses 10
		MaxConnsPerHost:     0,                // No limit, let connection pooling handle it
		IdleConnTimeout:     90 * time.Second, // Docker CLI uses 90s

		// TCP optimizations
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // Connection timeout
			KeepAlive: 30 * time.Second, // TCP keep-alive
			DualStack: true,             // Enable IPv4 and IPv6
		}).DialContext,

		// TLS optimizations
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: 10 * time.Second, // TLS handshake timeout

		// Response handling optimizations
		ResponseHeaderTimeout: 30 * time.Second, // Response header timeout
		ExpectContinueTimeout: 1 * time.Second,  // Expect: 100-continue timeout

		// Compression handling - Docker CLI enables this
		DisableCompression: false, // Enable gzip compression

		// HTTP/2 will be configured separately
		ForceAttemptHTTP2: true, // Force HTTP/2 attempt
	}

	// Apply adaptive connection limits based on system resources
	maxIdle, maxIdlePerHost, maxPerHost := AdaptiveConnectionLimits()
	transport.MaxIdleConns = maxIdle
	transport.MaxIdleConnsPerHost = maxIdlePerHost
	transport.MaxConnsPerHost = maxPerHost

	return transport
}

// GetOptimizedDialer returns a network dialer optimized for registry operations
func GetOptimizedDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
}

// GetOptimizedHTTPClient returns an HTTP client optimized for registry operations
func GetOptimizedHTTPClient(tlsConfig *tls.Config) *http.Client {
	transport := OptimizedTransport(tlsConfig)

	return &http.Client{
		Transport: transport,
		// Timeout for the entire request (Docker CLI uses no timeout, but we set a reasonable one)
		Timeout: 5 * time.Minute,
		// Don't follow redirects automatically for registry operations
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

// AdaptiveConnectionLimits adjusts connection limits based on system resources
// Enhanced for high-end GPU machines with 100+ CPU cores and high bandwidth networks
func AdaptiveConnectionLimits() (maxIdleConns, maxIdleConnsPerHost, maxConnsPerHost int) {
	cpus := runtime.NumCPU()

	// Base limits similar to Docker CLI
	maxIdleConns = 100
	maxIdleConnsPerHost = 10
	maxConnsPerHost = 0 // Unlimited

	// Enhanced scaling for high-end GPU machines with 100+ CPU cores
	if cpus > 4 {
		// More aggressive scaling for high-end systems
		if cpus >= 100 {
			// For extreme configurations (100+ CPUs), scale much more aggressively
			maxIdleConns = 1000 + (cpus-100)*5       // Start at 1000 for 100+ CPU systems
			maxIdleConnsPerHost = 100 + (cpus-100)*2 // Start at 100 per host for 100+ CPU systems
		} else if cpus >= 32 {
			// For high-end systems (32-99 CPUs), scale moderately
			maxIdleConns = 500 + (cpus-32)*10      // Scale up for high-end systems
			maxIdleConnsPerHost = 50 + (cpus-32)*2 // Scale up per-host connections
		} else {
			// For mid-range systems (5-31 CPUs), use original scaling
			maxIdleConns = 100 + (cpus-4)*10      // Original scaling
			maxIdleConnsPerHost = 10 + (cpus-4)*2 // Original scaling
		}
	}

	// Enhanced caps for high-end GPU machines and high bandwidth networks (2Gbps+)
	// Remove restrictive caps to allow proper utilization of extreme hardware
	if maxIdleConns > 2000 {
		maxIdleConns = 2000 // Cap at 2000 for extreme configurations
	}
	if maxIdleConnsPerHost > 200 {
		maxIdleConnsPerHost = 200 // Cap at 200 per host for extreme configurations
	}

	return maxIdleConns, maxIdleConnsPerHost, maxConnsPerHost
}
