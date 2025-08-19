// Package traefik_wol implements a Wake-on-LAN middleware plugin for Traefik.
package traefik_wol

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds the plugin configuration.
type Config struct {
	HealthCheck         string `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	MacAddress          string `json:"macAddress,omitempty" yaml:"macAddress,omitempty"`
	IPAddress           string `json:"ipAddress,omitempty" yaml:"ipAddress,omitempty"`
	BroadcastAddress    string `json:"broadcastAddress,omitempty" yaml:"broadcastAddress,omitempty"`
	NetworkInterface    string `json:"networkInterface,omitempty" yaml:"networkInterface,omitempty"`
	Port                string `json:"port,omitempty" yaml:"port,omitempty"`
	Timeout             string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RetryAttempts       string `json:"retryAttempts,omitempty" yaml:"retryAttempts,omitempty"`
	RetryInterval       string `json:"retryInterval,omitempty" yaml:"retryInterval,omitempty"`
	HealthCheckInterval string `json:"healthCheckInterval,omitempty" yaml:"healthCheckInterval,omitempty"`
	Debug               bool   `json:"debug,omitempty" yaml:"debug,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Port:                "9",
		Timeout:             "30",
		RetryAttempts:       "3",
		RetryInterval:       "5",
		HealthCheckInterval: "10",
		Debug:               false,
	}
}

// healthStatus holds cached health check results
type healthStatus struct {
	isHealthy  bool
	lastCheck  time.Time
	lastState  bool
}

// WOLPlugin is the main plugin struct.
type WOLPlugin struct {
	next                http.Handler
	name                string
	healthCheck         string
	macAddress          string
	ipAddress           string
	broadcastAddress    string
	networkInterface    string
	port                int
	timeout             time.Duration
	retryAttempts       int
	retryInterval       time.Duration
	healthCheckInterval time.Duration
	debug               bool
	healthCache         *healthStatus
	healthMutex         sync.RWMutex
}

// New creates a new WOL plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.HealthCheck == "" {
		return nil, fmt.Errorf("healthCheck URL is required")
	}
	if config.MacAddress == "" {
		return nil, fmt.Errorf("macAddress is required")
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

	healthCheckInterval, err := strconv.Atoi(config.HealthCheckInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid healthCheckInterval: %v", err)
	}

	return &WOLPlugin{
		next:                next,
		name:                name,
		healthCheck:         config.HealthCheck,
		macAddress:          config.MacAddress,
		ipAddress:           config.IPAddress,
		broadcastAddress:    config.BroadcastAddress,
		networkInterface:    config.NetworkInterface,
		port:                port,
		timeout:             time.Duration(timeout) * time.Second,
		retryAttempts:       retryAttempts,
		retryInterval:       time.Duration(retryInterval) * time.Second,
		healthCheckInterval: time.Duration(healthCheckInterval) * time.Second,
		debug:               config.Debug,
		healthCache:         &healthStatus{},
		healthMutex:         sync.RWMutex{},
	}, nil
}

// ServeHTTP implements the http.Handler interface.
func (w *WOLPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	isHealthy := w.getCachedHealthStatus()

	if !isHealthy {
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
	}

	w.next.ServeHTTP(rw, req)
}

// getCachedHealthStatus returns cached health status or performs new check if cache expired
func (w *WOLPlugin) getCachedHealthStatus() bool {
	w.healthMutex.RLock()
	cache := w.healthCache
	now := time.Now()
	
	// Check if cache is valid
	if now.Sub(cache.lastCheck) < w.healthCheckInterval {
		w.healthMutex.RUnlock()
		return cache.isHealthy
	}
	w.healthMutex.RUnlock()

	// Cache expired, perform new health check
	w.healthMutex.Lock()
	defer w.healthMutex.Unlock()

	// Double-check pattern - another goroutine might have updated while waiting for lock
	if now.Sub(w.healthCache.lastCheck) < w.healthCheckInterval {
		return w.healthCache.isHealthy
	}

	newHealth := w.performHealthCheck()
	
	// Log only on state changes or debug mode
	if w.healthCache.lastState != newHealth || w.debug {
		if w.debug || w.healthCache.lastCheck.IsZero() {
			fmt.Printf("WOL Plugin [%s]: Health status changed to %v for %s\n", w.name, newHealth, w.healthCheck)
		}
		w.healthCache.lastState = newHealth
	}
	
	w.healthCache.isHealthy = newHealth
	w.healthCache.lastCheck = now
	
	return newHealth
}

func (w *WOLPlugin) performHealthCheck() bool {
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

// isHealthy performs a direct health check without caching (used by waitForService)
func (w *WOLPlugin) isHealthy() bool {
	return w.performHealthCheck()
}

// getNetworkInterfaces returns available network interfaces for WOL packet sending
func (w *WOLPlugin) getNetworkInterfaces() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	var validInterfaces []net.Interface
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		// If specific interface is configured, only use that one
		if w.networkInterface != "" && iface.Name != w.networkInterface {
			continue
		}
		
		validInterfaces = append(validInterfaces, iface)
	}
	
	if len(validInterfaces) == 0 {
		return nil, fmt.Errorf("no valid network interfaces found")
	}
	
	return validInterfaces, nil
}

// calculateBroadcastAddress calculates broadcast address for a given network
func (w *WOLPlugin) calculateBroadcastAddress(ip net.IP, mask net.IPMask) net.IP {
	if ip == nil || mask == nil {
		return nil
	}
	
	network := ip.Mask(mask)
	broadcast := make(net.IP, len(network))
	for i := range network {
		broadcast[i] = network[i] | ^mask[i]
	}
	
	return broadcast
}

// getBroadcastAddresses returns all possible broadcast addresses for WOL
func (w *WOLPlugin) getBroadcastAddresses() []string {
	var addresses []string
	
	// Use configured broadcast address if provided
	if w.broadcastAddress != "" {
		addresses = append(addresses, w.broadcastAddress)
		return addresses
	}
	
	// Auto-discover broadcast addresses
	interfaces, err := w.getNetworkInterfaces()
	if err != nil {
		if w.debug {
			fmt.Printf("WOL Plugin [%s]: Failed to get interfaces: %v\n", w.name, err)
		}
		return addresses
	}
	
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				broadcast := w.calculateBroadcastAddress(ipNet.IP, ipNet.Mask)
				if broadcast != nil {
					addresses = append(addresses, broadcast.String())
				}
			}
		}
	}
	
	// Add common broadcast addresses as fallback
	if len(addresses) == 0 {
		addresses = append(addresses, "255.255.255.255") // Limited broadcast
	}
	
	return addresses
}

func (w *WOLPlugin) sendWOLPacket() error {
	macBytes, err := w.parseMACAddress(w.macAddress)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}

	packet := w.createMagicPacket(macBytes)
	sentSuccessfully := false
	var lastError error

	// Try unicast to specific IP first (if provided)
	if w.ipAddress != "" {
		err := w.sendToAddress(packet, w.ipAddress)
		if err == nil {
			sentSuccessfully = true
			if w.debug {
				fmt.Printf("WOL Plugin [%s]: Magic packet sent via unicast to %s (%s:%d)\n", w.name, w.macAddress, w.ipAddress, w.port)
			}
		} else {
			lastError = err
			if w.debug {
				fmt.Printf("WOL Plugin [%s]: Unicast failed: %v\n", w.name, err)
			}
		}
	}

	// Try broadcast addresses for better container/LXC compatibility
	broadcastAddresses := w.getBroadcastAddresses()
	for _, broadcastAddr := range broadcastAddresses {
		err := w.sendToAddress(packet, broadcastAddr)
		if err == nil {
			sentSuccessfully = true
			if w.debug {
				fmt.Printf("WOL Plugin [%s]: Magic packet sent via broadcast to %s (%s:%d)\n", w.name, w.macAddress, broadcastAddr, w.port)
			}
		} else {
			lastError = err
			if w.debug {
				fmt.Printf("WOL Plugin [%s]: Broadcast to %s failed: %v\n", w.name, broadcastAddr, err)
			}
		}
	}

	if !sentSuccessfully {
		return fmt.Errorf("failed to send WOL packet to any address: %v", lastError)
	}

	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Magic packet sent to %s\n", w.name, w.macAddress)
	}
	return nil
}

// sendToAddress sends WOL packet to a specific address
func (w *WOLPlugin) sendToAddress(packet []byte, targetAddr string) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", targetAddr, w.port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %v", targetAddr, err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection to %s: %v", targetAddr, err)
	}
	defer conn.Close()

	// Enable broadcast for this socket
	if udpConn, ok := conn.(*net.UDPConn); ok {
		if rawConn, err := udpConn.SyscallConn(); err == nil {
			rawConn.Control(func(fd uintptr) {
				// Enable broadcast on socket (platform-specific implementation would go here)
				// For now, we rely on the OS default behavior
			})
		}
	}

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to send packet to %s: %v", targetAddr, err)
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
		if w.performHealthCheck() {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}
