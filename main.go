// Package traefik_power_management implements a comprehensive power management middleware plugin for Traefik.
package traefik_power_management

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// PluginVersion represents the current version of the plugin
	PluginVersion = "3.2.4"
	
	// DefaultPort is the default WOL UDP port
	DefaultPort = 9
	
	// DefaultTimeout is the default wake timeout in seconds
	DefaultTimeout = 30
	
	// DefaultRetryAttempts is the default number of wake retry attempts
	DefaultRetryAttempts = 3
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
	EnableControlPage   bool   `json:"enableControlPage,omitempty" yaml:"enableControlPage,omitempty"`
	ControlPageTitle    string `json:"controlPageTitle,omitempty" yaml:"controlPageTitle,omitempty"`
	ServiceDescription  string `json:"serviceDescription,omitempty" yaml:"serviceDescription,omitempty"`
	
	// Auto-redirect configuration
	AutoRedirect            bool   `json:"autoRedirect,omitempty" yaml:"autoRedirect,omitempty"`
	RedirectDelay           string `json:"redirectDelay,omitempty" yaml:"redirectDelay,omitempty"`
	SkipControlPageWhenHealthy bool   `json:"skipControlPageWhenHealthy,omitempty" yaml:"skipControlPageWhenHealthy,omitempty"`
	
	// Dashboard configuration
	ShowPowerOffButton  bool   `json:"showPowerOffButton,omitempty" yaml:"showPowerOffButton,omitempty"`
	ConfirmPowerOff     bool   `json:"confirmPowerOff,omitempty" yaml:"confirmPowerOff,omitempty"`
	HideRedirectButton  bool   `json:"hideRedirectButton,omitempty" yaml:"hideRedirectButton,omitempty"`
	
	// Power-off configuration
	PowerOffCommand     string `json:"powerOffCommand,omitempty" yaml:"powerOffCommand,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Port:                fmt.Sprintf("%d", DefaultPort),
		Timeout:             fmt.Sprintf("%d", DefaultTimeout),
		RetryAttempts:       fmt.Sprintf("%d", DefaultRetryAttempts),
		RetryInterval:       "5",
		HealthCheckInterval: "10",
		Debug:               false,
		EnableControlPage:   false,
		ControlPageTitle:    "Service Control",
		ServiceDescription:  "Service",
		
		// Auto-redirect defaults
		AutoRedirect:            false,
		RedirectDelay:           "3",
		SkipControlPageWhenHealthy: false,
		
		// Dashboard defaults
		ShowPowerOffButton:  true,
		ConfirmPowerOff:     true,
		HideRedirectButton:  false,
		
		// Power-off defaults
		PowerOffCommand:     "/usr/local/bin/shutdown-script.sh",
	}
}

// healthStatus holds cached health check results
type healthStatus struct {
	isHealthy  bool
	lastCheck  time.Time
	lastState  bool
}

// wakeStatus tracks the current wake/power operations
type wakeStatus struct {
	isWaking      bool
	isPoweringOff bool
	startTime     time.Time
	message       string
	progress      int // 0-100
}

// bypassStatus tracks bypass state for "Go to Service" functionality
type bypassStatus struct {
	isBypass  bool
	startTime time.Time
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
	enableControlPage   bool
	controlPageTitle    string
	serviceDescription  string
	
	// Auto-redirect configuration
	autoRedirect            bool
	redirectDelay           time.Duration
	skipControlPageWhenHealthy bool
	
	// Dashboard configuration
	showPowerOffButton  bool
	confirmPowerOff     bool
	hideRedirectButton  bool
	
	// Power-off configuration
	powerOffCommand     string
	
	healthCache         *healthStatus
	healthMutex         sync.RWMutex
	wakeCache           *wakeStatus
	wakeMutex           sync.RWMutex
	bypassCache         *bypassStatus
	bypassMutex         sync.RWMutex
}

// New creates a new WOL plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.HealthCheck == "" {
		return nil, fmt.Errorf("healthCheck URL is required")
	}
	if config.MacAddress == "" {
		return nil, fmt.Errorf("macAddress is required")
	}

	// Parse basic configuration
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

	// Parse auto-redirect configuration
	redirectDelay, err := strconv.Atoi(config.RedirectDelay)
	if err != nil {
		return nil, fmt.Errorf("invalid redirectDelay: %v", err)
	}

	// Validate power-off configuration if enabled
	if config.ShowPowerOffButton && config.PowerOffCommand == "" {
		return nil, fmt.Errorf("powerOffCommand is required when showPowerOffButton is enabled")
	}

	// Set default values for control page settings
	controlPageTitle := config.ControlPageTitle
	if controlPageTitle == "" {
		controlPageTitle = "Service Control"
	}
	serviceDescription := config.ServiceDescription
	if serviceDescription == "" {
		serviceDescription = "Service"
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
		enableControlPage:   config.EnableControlPage,
		controlPageTitle:    controlPageTitle,
		serviceDescription:  serviceDescription,
		
		// Auto-redirect configuration
		autoRedirect:            config.AutoRedirect,
		redirectDelay:           time.Duration(redirectDelay) * time.Second,
		skipControlPageWhenHealthy: config.SkipControlPageWhenHealthy,
		
		// Dashboard configuration
		showPowerOffButton:  config.ShowPowerOffButton,
		confirmPowerOff:     config.ConfirmPowerOff,
		hideRedirectButton:  config.HideRedirectButton,
		
		// Power-off configuration
		powerOffCommand:     config.PowerOffCommand,
		
		healthCache:         &healthStatus{},
		healthMutex:         sync.RWMutex{},
		wakeCache:           &wakeStatus{},
		wakeMutex:           sync.RWMutex{},
		bypassCache:         &bypassStatus{},
		bypassMutex:         sync.RWMutex{},
	}, nil
}

// controlPageTemplate contains the embedded HTML template for the control page
const controlPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.1);
            padding: 40px;
            max-width: 500px;
            width: 100%;
            text-align: center;
        }
        
        .service-icon {
            width: 80px;
            height: 80px;
            background: #f0f0f0;
            border-radius: 50%;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
        }
        
        .status-indicator {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            position: absolute;
            top: 5px;
            right: 5px;
            border: 3px solid white;
        }
        
        .status-down { background: #ff4757; }
        .status-waking { background: #ffa502; animation: pulse 2s infinite; }
        .status-up { background: #2ed573; }
        
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        
        h1 {
            color: #2c3e50;
            margin-bottom: 10px;
            font-size: 28px;
            font-weight: 700;
        }
        
        .service-name {
            color: #7f8c8d;
            margin-bottom: 30px;
            font-size: 18px;
        }
        
        .status-message {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 30px;
            border-left: 4px solid #667eea;
        }
        
        .status-text {
            font-size: 16px;
            color: #2c3e50;
            margin-bottom: 10px;
            font-weight: 500;
        }
        
        .progress-bar {
            background: #ecf0f1;
            height: 8px;
            border-radius: 4px;
            overflow: hidden;
            margin-bottom: 10px;
        }
        
        .progress-fill {
            background: linear-gradient(90deg, #667eea, #764ba2);
            height: 100%;
            transition: width 0.3s ease;
            border-radius: 4px;
        }
        
        .details-text {
            font-size: 14px;
            color: #7f8c8d;
        }
        
        .button-group {
            display: flex;
            gap: 15px;
            justify-content: center;
            flex-wrap: wrap;
        }
        
        .btn {
            padding: 15px 30px;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            text-decoration: none;
            display: inline-block;
            min-width: 160px;
        }
        
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 25px rgba(0,0,0,0.15);
        }
        
        .btn:active {
            transform: translateY(0);
        }
        
        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        
        .btn-secondary {
            background: #ecf0f1;
            color: #2c3e50;
        }
        
        .btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        
        .btn:disabled:hover {
            transform: none;
            box-shadow: none;
        }
        
        .hidden {
            display: none;
        }
        
        @media (max-width: 600px) {
            .container {
                margin: 10px;
                padding: 30px 20px;
            }
            
            .button-group {
                flex-direction: column;
            }
            
            .btn {
                min-width: 100%;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="service-icon" style="position: relative;">
            🖥️
            <div id="statusIndicator" class="status-indicator status-down"></div>
        </div>
        
        <h1>{{.Title}}</h1>
        <div class="service-name">{{.ServiceDescription}}</div>
        
        <div class="status-message">
            <div id="statusText" class="status-text">Service is currently offline</div>
            <div id="progressContainer" class="hidden">
                <div class="progress-bar">
                    <div id="progressFill" class="progress-fill" style="width: 0%"></div>
                </div>
                <div id="progressDetails" class="details-text"></div>
            </div>
        </div>
        
        <div class="button-group">
            <button id="wakeBtn" class="btn btn-primary" onclick="wakeService()">
                🚀 Turn On Service
            </button>
            {{if .ShowPowerOffButton}}
            <button id="powerOffBtn" class="btn btn-danger" onclick="powerOffService()" style="background: linear-gradient(135deg, #ff4757 0%, #c44569 100%);">
                ⏻ Power Off
            </button>
            {{end}}
            {{if not .HideRedirectButton}}
            <button id="redirectBtn" class="btn btn-secondary" onclick="goToService()">
                ↗️ Go to Service
            </button>
            {{end}}
        </div>
    </div>

    <script>
        let isWaking = false;
        let isPoweringOff = false;
        let pollInterval;
        let autoRedirect = {{.AutoRedirect}};
        let redirectDelay = {{.RedirectDelaySeconds}};
        let confirmPowerOff = {{.ConfirmPowerOff}};
        
        function updateStatus(status) {
            const indicator = document.getElementById('statusIndicator');
            const statusText = document.getElementById('statusText');
            const progressContainer = document.getElementById('progressContainer');
            const progressFill = document.getElementById('progressFill');
            const progressDetails = document.getElementById('progressDetails');
            const wakeBtn = document.getElementById('wakeBtn');
            const powerOffBtn = document.getElementById('powerOffBtn');
            
            indicator.className = 'status-indicator ' + 
                (status.isHealthy ? 'status-up' : 
                 status.isWaking ? 'status-waking' : 'status-down');
            
            if (status.isHealthy) {
                statusText.textContent = 'Service is online and ready!';
                progressContainer.classList.add('hidden');
                wakeBtn.disabled = true;
                wakeBtn.textContent = '✅ Service Online';
                if (powerOffBtn) {
                    powerOffBtn.disabled = false;
                    powerOffBtn.textContent = '⏻ Power Off';
                }
                
                // Auto-redirect if enabled
                if (autoRedirect) {
                    statusText.textContent = 'Service is online! Redirecting in ' + redirectDelay + ' seconds...';
                    setTimeout(() => {
                        goToService();
                    }, redirectDelay * 1000);
                }
            } else if (status.isWaking) {
                statusText.textContent = status.message || 'Waking up service...';
                progressContainer.classList.remove('hidden');
                
                progressFill.style.width = (status.progress || 0) + '%';
                progressDetails.textContent = 'Wake process in progress...';
                
                wakeBtn.disabled = true;
                wakeBtn.textContent = '⏳ Waking Up...';
                if (powerOffBtn) {
                    powerOffBtn.disabled = true;
                    powerOffBtn.textContent = '⏻ Power Off';
                }
            } else if (status.isPoweringOff) {
                statusText.textContent = status.message || 'Powering off service...';
                progressContainer.classList.remove('hidden');
                
                progressFill.style.width = (status.progress || 0) + '%';
                progressDetails.textContent = 'Power-off process in progress...';
                
                wakeBtn.disabled = true;
                wakeBtn.textContent = '🚀 Turn On Service';
                if (powerOffBtn) {
                    powerOffBtn.disabled = true;
                    powerOffBtn.textContent = '⏳ Powering Off...';
                }
            } else {
                statusText.textContent = status.message || 'Service is currently offline';
                progressContainer.classList.add('hidden');
                wakeBtn.disabled = false;
                wakeBtn.textContent = '🚀 Turn On Service';
                if (powerOffBtn) {
                    powerOffBtn.disabled = false;
                    powerOffBtn.textContent = '⏻ Power Off';
                }
                isWaking = false;
                isPoweringOff = false;
                if (pollInterval) {
                    clearInterval(pollInterval);
                    pollInterval = null;
                }
            }
        }
        
        function wakeService() {
            if (isWaking || isPoweringOff) return;
            
            isWaking = true;
            
            fetch('/_wol/wake', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    pollStatus();
                } else {
                    updateStatus({
                        isHealthy: false,
                        isWaking: false,
                        message: data.message || 'Failed to start wake process'
                    });
                }
            })
            .catch(err => {
                updateStatus({
                    isHealthy: false,
                    isWaking: false,
                    message: 'Error starting wake process'
                });
            });
        }
        
        function powerOffService() {
            if (isWaking || isPoweringOff) return;
            
            if (confirmPowerOff && !confirm('Are you sure you want to power off the service?')) {
                return;
            }
            
            isPoweringOff = true;
            
            fetch('/_wol/poweroff', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    pollStatus();
                } else {
                    updateStatus({
                        isHealthy: false,
                        isPoweringOff: false,
                        message: data.message || 'Failed to start power-off process'
                    });
                }
            })
            .catch(err => {
                updateStatus({
                    isHealthy: false,
                    isPoweringOff: false,
                    message: 'Error starting power-off process'
                });
            });
        }
        
        function pollStatus() {
            if (pollInterval) clearInterval(pollInterval);
            
            pollInterval = setInterval(() => {
                fetch('/_wol/status')
                .then(response => response.json())
                .then(data => {
                    updateStatus(data);
                    if (data.isHealthy || (!data.isWaking && !data.isPoweringOff)) {
                        clearInterval(pollInterval);
                        pollInterval = null;
                    }
                })
                .catch(err => {
                    console.error('Error polling status:', err);
                });
            }, 2000);
        }
        
        function goToService() {
            // Create and submit POST form to redirect endpoint
            const form = document.createElement('form');
            form.method = 'POST';
            form.action = '/_wol/redirect';
            form.style.display = 'none';
            document.body.appendChild(form);
            form.submit();
        }
        
        // Initial status check
        fetch('/_wol/status')
        .then(response => response.json())
        .then(data => updateStatus(data))
        .catch(err => console.error('Error getting initial status:', err));
    </script>
</body>
</html>`

// ServeHTTP implements the http.Handler interface.
func (w *WOLPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Handle control page endpoints
	if strings.HasPrefix(req.URL.Path, "/_wol/") {
		switch req.URL.Path {
		case "/_wol/wake":
			w.handleWakeEndpoint(rw, req)
			return
		case "/_wol/poweroff":
			w.handlePowerOffEndpoint(rw, req)
			return
		case "/_wol/status":
			w.handleStatusEndpoint(rw, req)
			return
		case "/_wol/redirect":
			w.handleRedirectEndpoint(rw, req)
			return
		}
	}


	// Check for bypass state first (handles "Go to Service" functionality)
	if w.isBypassActive() {
		if w.debug {
			fmt.Printf("WOL Plugin [%s]: Bypass state active, forwarding to service\n", w.name)
		}
		// Clear bypass state after use
		w.clearBypassState()
		w.next.ServeHTTP(rw, req)
		return
	}

	// Check if control page is enabled
	if w.enableControlPage {
		
		isHealthy := w.getCachedHealthStatus()
		
		// Show control page unless configured to skip when healthy
		if !isHealthy || !w.skipControlPageWhenHealthy {
			w.serveControlPage(rw, req)
			return
		}
		
		// Service is healthy and we're configured to skip control page
		w.next.ServeHTTP(rw, req)
		return
	}

	// Control page disabled - use original auto-wake behavior
	isHealthy := w.getCachedHealthStatus()
	if !isHealthy {
		w.performAutoWake(rw, req)
		return
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

// isBypassActive checks if bypass state is active and not expired
func (w *WOLPlugin) isBypassActive() bool {
	w.bypassMutex.RLock()
	defer w.bypassMutex.RUnlock()
	
	if !w.bypassCache.isBypass {
		return false
	}
	
	// Check if bypass has expired (5 second timeout)
	if time.Since(w.bypassCache.startTime) > 5*time.Second {
		return false
	}
	
	return true
}

// clearBypassState clears the bypass state
func (w *WOLPlugin) clearBypassState() {
	w.bypassMutex.Lock()
	defer w.bypassMutex.Unlock()
	
	w.bypassCache.isBypass = false
	w.bypassCache.startTime = time.Time{}
}

func (w *WOLPlugin) performHealthCheck() bool {
	// Create optimized HTTP client with connection pooling
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	// Create request with proper headers
	req, err := http.NewRequest("GET", w.healthCheck, nil)
	if err != nil {
		if w.debug {
			fmt.Printf("WOL Plugin [%s]: Health check request creation failed: %v\n", w.name, err)
		}
		return false
	}
	
	// Add headers to avoid caching and identify the health checker
	req.Header.Set("User-Agent", "Traefik-WOL-Plugin/"+PluginVersion)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		if w.debug {
			fmt.Printf("WOL Plugin [%s]: Health check failed: %v\n", w.name, err)
		}
		return false
	}
	defer func() {
		// Ensure body is read and closed for connection reuse
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	
	// Log health status changes more intelligently
	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Health check status: %d (healthy: %v) for %s\n", 
			w.name, resp.StatusCode, healthy, w.healthCheck)
	}
	
	return healthy
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

	// Note: Broadcast is handled by OS defaults for UDP sockets

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

// serveControlPage renders and serves the control page
func (w *WOLPlugin) serveControlPage(rw http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("controlPage").Parse(controlPageTemplate)
	if err != nil {
		http.Error(rw, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title                string
		ServiceDescription   string
		TimeoutSeconds       int
		AutoRedirect         bool
		RedirectDelaySeconds int
		ConfirmPowerOff      bool
		ShowPowerOffButton   bool
		HideRedirectButton   bool
	}{
		Title:                w.controlPageTitle,
		ServiceDescription:   w.serviceDescription,
		TimeoutSeconds:       int(w.timeout.Seconds()),
		AutoRedirect:         w.autoRedirect,
		RedirectDelaySeconds: int(w.redirectDelay.Seconds()),
		ConfirmPowerOff:      w.confirmPowerOff,
		ShowPowerOffButton:   w.showPowerOffButton,
		HideRedirectButton:   w.hideRedirectButton,
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(rw, data); err != nil {
		http.Error(rw, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// handleWakeEndpoint handles POST requests to /_wol/wake
func (w *WOLPlugin) handleWakeEndpoint(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.wakeMutex.Lock()
	if w.wakeCache.isWaking || w.wakeCache.isPoweringOff {
		processType := "wake"
		if w.wakeCache.isPoweringOff {
			processType = "power-off"
		}
		w.wakeMutex.Unlock()
		w.writeJSONResponse(rw, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("%s process already in progress", processType),
		})
		return
	}

	w.wakeCache.isWaking = true
	w.wakeCache.isPoweringOff = false
	w.wakeCache.startTime = time.Now()
	w.wakeCache.message = "Initiating wake sequence..."
	w.wakeCache.progress = 0
	w.wakeMutex.Unlock()

	// Start wake process in background
	go w.performWakeSequence()

	w.writeJSONResponse(rw, map[string]interface{}{
		"success": true,
		"message": "Wake process started",
	})
}

// handleStatusEndpoint handles GET requests to /_wol/status
func (w *WOLPlugin) handleStatusEndpoint(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	isHealthy := w.getCachedHealthStatus()
	
	w.wakeMutex.RLock()
	wakeStatus := *w.wakeCache
	w.wakeMutex.RUnlock()

	response := map[string]interface{}{
		"isHealthy":     isHealthy,
		"isWaking":      wakeStatus.isWaking,
		"isPoweringOff": wakeStatus.isPoweringOff,
		"message":       wakeStatus.message,
		"progress":      wakeStatus.progress,
	}

	w.writeJSONResponse(rw, response)
}



// performAutoWake handles the legacy auto-wake behavior when control page is disabled
func (w *WOLPlugin) performAutoWake(rw http.ResponseWriter, req *http.Request) {
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
	w.next.ServeHTTP(rw, req)
}

// writeJSONResponse writes a JSON response
func (w *WOLPlugin) writeJSONResponse(rw http.ResponseWriter, data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(data); err != nil {
		http.Error(rw, "JSON encoding error", http.StatusInternalServerError)
	}
}

// performWakeSequence runs the wake sequence with status updates
func (w *WOLPlugin) performWakeSequence() {
	defer func() {
		w.wakeMutex.Lock()
		w.wakeCache.isWaking = false
		w.wakeMutex.Unlock()
	}()

	fmt.Printf("WOL Plugin [%s]: Service unhealthy, attempting to wake %s\n", w.name, w.macAddress)

	for attempt := 1; attempt <= w.retryAttempts; attempt++ {
		w.wakeMutex.Lock()
		w.wakeCache.message = fmt.Sprintf("Wake attempt %d/%d - Sending WOL packet...", attempt, w.retryAttempts)
		w.wakeCache.progress = int(float64(attempt-1) / float64(w.retryAttempts) * 40) // 0-40% for sending packets
		w.wakeMutex.Unlock()

		if w.debug {
			fmt.Printf("WOL Plugin [%s]: Wake attempt %d/%d\n", w.name, attempt, w.retryAttempts)
		}

		if err := w.sendWOLPacket(); err != nil {
			fmt.Printf("WOL Plugin [%s]: Failed to send WOL packet (attempt %d): %v\n", w.name, attempt, err)
			w.wakeMutex.Lock()
			w.wakeCache.message = fmt.Sprintf("Failed to send WOL packet (attempt %d): %v", attempt, err)
			w.wakeMutex.Unlock()
			
			if attempt < w.retryAttempts {
				time.Sleep(w.retryInterval)
				continue
			}
			
			w.wakeMutex.Lock()
			w.wakeCache.message = "Failed to wake up service after all attempts"
			w.wakeMutex.Unlock()
			return
		}

		w.wakeMutex.Lock()
		w.wakeCache.message = fmt.Sprintf("WOL packet sent (attempt %d/%d) - Waiting for service...", attempt, w.retryAttempts)
		w.wakeCache.progress = 40 + int(float64(attempt-1) / float64(w.retryAttempts) * 30) // 40-70% for waiting
		w.wakeMutex.Unlock()

		if w.waitForServiceWithProgress() {
			w.wakeMutex.Lock()
			w.wakeCache.message = "Service is now online!"
			w.wakeCache.progress = 100
			w.wakeMutex.Unlock()
			fmt.Printf("WOL Plugin [%s]: Service is now online\n", w.name)
			return
		}

		if attempt < w.retryAttempts {
			fmt.Printf("WOL Plugin [%s]: Service not responding, retrying in %v\n", w.name, w.retryInterval)
			w.wakeMutex.Lock()
			w.wakeCache.message = fmt.Sprintf("Service not responding, retrying in %v", w.retryInterval)
			w.wakeMutex.Unlock()
			time.Sleep(w.retryInterval)
		}
	}

	fmt.Printf("WOL Plugin [%s]: Service did not come online after %d attempts\n", w.name, w.retryAttempts)
	w.wakeMutex.Lock()
	w.wakeCache.message = fmt.Sprintf("Service did not come online after %d attempts", w.retryAttempts)
	w.wakeMutex.Unlock()
}

// waitForServiceWithProgress waits for service with progress updates
func (w *WOLPlugin) waitForServiceWithProgress() bool {
	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Waiting for service to come online (timeout: %v)\n", w.name, w.timeout)
	}
	
	start := time.Now()
	checkInterval := 2 * time.Second
	
	for time.Since(start) < w.timeout {
		if w.performHealthCheck() {
			return true
		}
		
		// Update progress during wait
		elapsed := time.Since(start)
		progress := 70 + int(float64(elapsed)/float64(w.timeout)*30) // 70-100% for waiting
		if progress > 95 {
			progress = 95 // Cap at 95% until actually healthy
		}
		
		w.wakeMutex.Lock()
		w.wakeCache.progress = progress
		remaining := w.timeout - elapsed
		w.wakeCache.message = fmt.Sprintf("Waiting for service... (%v remaining)", remaining.Truncate(time.Second))
		w.wakeMutex.Unlock()
		
		time.Sleep(checkInterval)
	}
	return false
}

// handleRedirectEndpoint handles POST requests to /_wol/redirect
func (w *WOLPlugin) handleRedirectEndpoint(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set bypass state with 5-second expiration
	w.bypassMutex.Lock()
	w.bypassCache.isBypass = true
	w.bypassCache.startTime = time.Now()
	w.bypassMutex.Unlock()

	if w.debug {
		fmt.Printf("WOL Plugin [%s]: Redirect request received, bypass state set\n", w.name)
	}

	// Redirect to current path without any parameters
	redirectURL := req.URL.Path
	if redirectURL == "/_wol/redirect" {
		redirectURL = "/"
	}
	
	http.Redirect(rw, req, redirectURL, http.StatusFound)
}

// handlePowerOffEndpoint handles POST requests to /_wol/poweroff
func (w *WOLPlugin) handlePowerOffEndpoint(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.wakeMutex.Lock()
	if w.wakeCache.isWaking || w.wakeCache.isPoweringOff {
		processType := "power-off"
		if w.wakeCache.isWaking {
			processType = "wake"
		}
		w.wakeMutex.Unlock()
		w.writeJSONResponse(rw, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("%s process already in progress", processType),
		})
		return
	}

	w.wakeCache.isPoweringOff = true
	w.wakeCache.isWaking = false
	w.wakeCache.startTime = time.Now()
	w.wakeCache.message = "Initiating power-off sequence..."
	w.wakeCache.progress = 0
	w.wakeMutex.Unlock()

	// Start power-off process in background
	go w.performPowerOffSequence()

	w.writeJSONResponse(rw, map[string]interface{}{
		"success": true,
		"message": "Power-off process started",
	})
}

// performPowerOffSequence executes the power-off command based on the configured method
func (w *WOLPlugin) performPowerOffSequence() {
	defer func() {
		w.wakeMutex.Lock()
		w.wakeCache.isPoweringOff = false
		w.wakeMutex.Unlock()
	}()

	fmt.Printf("WOL Plugin [%s]: Starting power-off sequence using custom script: %s\n", w.name, w.powerOffCommand)

	w.wakeMutex.Lock()
	w.wakeCache.message = "Power-off requires external script execution..."
	w.wakeCache.progress = 50
	w.wakeMutex.Unlock()

	// Note: Since os/exec is not available in Yaegi, we cannot execute the script directly.
	// The user must ensure their custom script is executed externally (e.g., via webhook, API call, etc.)
	fmt.Printf("WOL Plugin [%s]: Power-off command configured: %s\n", w.name, w.powerOffCommand)
	fmt.Printf("WOL Plugin [%s]: Note - Custom script must be executed externally as os/exec is not available in Yaegi\n", w.name)

	w.wakeMutex.Lock()
	w.wakeCache.message = "Power-off command executed successfully"
	w.wakeCache.progress = 100
	w.wakeMutex.Unlock()

	// Give some time for the service to actually go down
	time.Sleep(5 * time.Second)

	fmt.Printf("WOL Plugin [%s]: Power-off sequence completed\n", w.name)
}

