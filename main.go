// Package main implements a Wake-on-LAN middleware plugin for Traefik.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Config holds the plugin configuration.
type Config struct {
	HealthCheck   string `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	MacAddress    string `json:"macAddress,omitempty" yaml:"macAddress,omitempty"`
	IPAddress     string `json:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	Port          string `json:"port,omitempty" yaml:"port,omitempty"`
	Timeout       string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RetryAttempts string `json:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty"`
	RetryInterval string `json:"retryInterval,omitempty" yaml:"retryInterval,omitempty"`
	Debug         bool   `json:"debug,omitempty" yaml:"debug,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Port:          "9",
		Timeout:       "30",
		RetryAttempts: "3",
		RetryInterval: "5",
		Debug:         false,
	}
}

// WOLPlugin is the main plugin struct.
type WOLPlugin struct {
	next          http.Handler
	name          string
	healthCheck   string
	macAddress    string
	ipAddress     string
	port          int
	timeout       time.Duration
	retryAttempts int
	retryInterval time.Duration
	debug         bool
}

// New creates a new WOL plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.HealthCheck == "" {
		return nil, fmt.Errorf("healthCheck URL is required")
	}
	if config.MacAddress == "" {
		return nil, fmt.Errorf("macAddress is required")
	}
	if config.IPAddress == "" {
		return nil, fmt.Errorf("ipAddress is required")
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	timeout, err := strconv.Atoi(config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %v", err)
	}

	retryAttempts, err := strconv.Atoi(config.RetryAttempts)
	if err != nil {
		return nil, fmt.Errorf("invalid retryAttempts: %v", err)
	}

	retryInterval, err := strconv.Atoi(config.RetryInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid retryInterval: %v", err)
	}

	return &WOLPlugin{
		next:          next,
		name:          name,
		healthCheck:   config.HealthCheck,
		macAddress:    config.MacAddress,
		ipAddress:     config.IPAddress,
		port:          port,
		timeout:       time.Duration(timeout) * time.Second,
		retryAttempts: retryAttempts,
		retryInterval: time.Duration(retryInterval) * time.Second,
		debug:         config.Debug,
	}, nil
}

// ServeHTTP implements the http.Handler interface.
func (w *WOLPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Checking health for %s\n", w.name, w.healthCheck)
	}

	if !w.isHealthy() {
		fmt.Printf("WOL Plugin [%s]: Service unhealthy, attempting to wake %s\n", w.name, w.macAddress)
		
		success := false
		for attempt := 1; attempt <= w.retryAttempts; attempt++ {
			if w.debug {
				fmt.Printf("WOL Plugin [%s]: Wake attempt %d/%d\n", w.name, attempt, w.retryAttempts)
			}

			if err := w.sendWOLPacket(); err != nil {
				fmt.Printf("WOL Plugin [%s]: Failed to send WOL packet (attempt %d): %v\n", w.name, attempt, err)
				if attempt < w.retryAttempts {
					time.Sleep(w.retryInterval)
					continue
				}
				http.Error(rw, "Failed to wake up service after all attempts", http.StatusServiceUnavailable)
				return
			}

			if w.waitForService() {
				success = true
				break
			}

			if attempt < w.retryAttempts {
				fmt.Printf("WOL Plugin [%s]: Service not responding, retrying in %v\n", w.name, w.retryInterval)
				time.Sleep(w.retryInterval)
			}
		}

		if !success {
			fmt.Printf("WOL Plugin [%s]: Service did not come online after %d attempts\n", w.name, w.retryAttempts)
			http.Error(rw, "Service did not respond after wake up attempts", http.StatusServiceUnavailable)
			return
		}

		fmt.Printf("WOL Plugin [%s]: Service is now online\n", w.name)
	} else if w.debug {
		fmt.Printf("WOL Plugin [%s]: Service is already healthy\n", w.name)
	}

	w.next.ServeHTTP(rw, req)
}

func (w *WOLPlugin) isHealthy() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(w.healthCheck)
	if err != nil {
		if w.debug {
			fmt.Printf("WOL Plugin [%s]: Health check failed: %v\n", w.name, err)
		}
		return false
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Health check status: %d (healthy: %v)\n", w.name, resp.StatusCode, healthy)
	}
	return healthy
}

func (w *WOLPlugin) sendWOLPacket() error {
	macBytes, err := w.parseMACAddress(w.macAddress)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}

	packet := w.createMagicPacket(macBytes)

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", w.ipAddress, w.port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Magic packet sent to %s (%s:%d)\n", w.name, w.macAddress, w.ipAddress, w.port)
	}
	return nil
}

func (w *WOLPlugin) parseMACAddress(macStr string) ([]byte, error) {
	macStr = strings.ReplaceAll(macStr, ":", "")
	macStr = strings.ReplaceAll(macStr, "-", "")
	macStr = strings.ReplaceAll(macStr, ".", "")
	macStr = strings.ToLower(macStr)

	if len(macStr) != 12 {
		return nil, fmt.Errorf("MAC address must be 12 hex characters")
	}

	macBytes := make([]byte, 6)
	for i := 0; i < 6; i++ {
		b, err := strconv.ParseUint(macStr[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex in MAC address: %v", err)
		}
		macBytes[i] = byte(b)
	}

	return macBytes, nil
}

func (w *WOLPlugin) createMagicPacket(macBytes []byte) []byte {
	packet := make([]byte, 102)

	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}

	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], macBytes)
	}

	return packet
}

func (w *WOLPlugin) waitForService() bool {
	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Waiting for service to come online (timeout: %v)\n", w.name, w.timeout)
	}
	
	start := time.Now()
	for time.Since(start) < w.timeout {
		if w.isHealthy() {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}