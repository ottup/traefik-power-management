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

// Config holds the plugin configuration
type Config struct {
	HealthCheck   string `json:"healthCheck,omitempty"`
	MacAddress    string `json:"macAddress,omitempty"`
	IPAddress     string `json:"ipAddress,omitempty"`
	Port          string `json:"port,omitempty"`
	Timeout       string `json:"timeout,omitempty"`
	RetryAttempts string `json:"retryAttempts,omitempty"`
	RetryInterval string `json:"retryInterval,omitempty"`
	Debug         bool   `json:"debug,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration
func CreateConfig() *Config {
	return &Config{
		Port:          "9",
		Timeout:       "30",
		RetryAttempts: "3",
		RetryInterval: "5",
		Debug:         false,
	}
}

// WOLPlugin holds the plugin instance
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

// New creates a new WOL plugin instance
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

// ServeHTTP handles the HTTP request
func (w *WOLPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if w.debug {
		fmt.Printf("Custom WOL Plugin: Checking health for %s\n", w.healthCheck)
	}

	if !w.isHealthy() {
		fmt.Printf("Custom WOL Plugin: Service unhealthy, attempting to wake %s\n", w.macAddress)
		
		success := false
		for attempt := 1; attempt <= w.retryAttempts; attempt++ {
			if w.debug {
				fmt.Printf("Custom WOL Plugin: Wake attempt %d/%d\n", attempt, w.retryAttempts)
			}

			if err := w.sendWOLPacket(); err != nil {
				fmt.Printf("Custom WOL Plugin: Failed to send WOL packet (attempt %d): %v\n", attempt, err)
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
				fmt.Printf("Custom WOL Plugin: Service not responding, retrying in %v\n", w.retryInterval)
				time.Sleep(w.retryInterval)
			}
		}

		if !success {
			fmt.Printf("Custom WOL Plugin: Service did not come online after %d attempts\n", w.retryAttempts)
			http.Error(rw, "Service did not respond after wake up attempts", http.StatusServiceUnavailable)
			return
		}

		fmt.Printf("Custom WOL Plugin: Service is now online\n")
	} else if w.debug {
		fmt.Printf("Custom WOL Plugin: Service is already healthy\n")
	}

	w.next.ServeHTTP(rw, req)
}

// isHealthy checks if the target service is responding
func (w *WOLPlugin) isHealthy() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(w.healthCheck)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// sendWOLPacket sends a Wake-on-LAN magic packet
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
		fmt.Printf("Custom WOL Plugin: Magic packet sent to %s (%s:%d)\n", w.macAddress, w.ipAddress, w.port)
	}
	return nil
}

// parseMACAddress parses a MAC address string into bytes
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

// createMagicPacket creates a WOL magic packet
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

// waitForService waits for the service to become healthy
func (w *WOLPlugin) waitForService() bool {
	if w.debug {
		fmt.Printf("Custom WOL Plugin: Waiting for service to come online (timeout: %v)\n", w.timeout)
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