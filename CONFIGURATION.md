# Configuration Reference

This document provides comprehensive configuration examples and reference for the Traefik Power Management Plugin.

## Table of Contents

- [Basic Configuration](#basic-configuration)
- [Complete Configuration Reference](#complete-configuration-reference)
- [Power-Off Method Examples](#power-off-method-examples)
- [Alternative Configuration Formats](#alternative-configuration-formats)
- [MAC Address Formats](#mac-address-formats)
- [Container and Network Configuration](#container-and-network-configuration)
- [Interactive Power Control Dashboard](#interactive-power-control-dashboard)
- [Usage Examples](#usage-examples)

## Basic Configuration

```yaml
middlewares:
  power-middleware:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
```

## Complete Configuration Reference

```yaml
middlewares:
  power-management:
    plugin:
      traefik-wol:
        # === REQUIRED SETTINGS ===
        healthCheck: "http://192.168.1.100:3000/health"  # Health check endpoint
        macAddress: "00:11:22:33:44:55"                   # Target device MAC address
        
        # === WAKE-ON-LAN SETTINGS ===
        ipAddress: "192.168.1.100"                        # Target IP (optional, uses broadcast if not set)
        broadcastAddress: "192.168.1.255"                 # Custom broadcast address
        networkInterface: "eth0"                          # Specific network interface
        port: "9"                                         # WOL UDP port (default: 9)
        timeout: "30"                                     # Wake timeout in seconds (default: 30)
        retryAttempts: "3"                                # Number of wake retry attempts (default: 3)
        retryInterval: "5"                                # Delay between retries in seconds (default: 5)
        healthCheckInterval: "10"                         # Health check cache interval (default: 10)
        
        # === CONTROL PAGE SETTINGS ===
        enableControlPage: true                           # Enable web dashboard (default: false)
        controlPageTitle: "Server Power Control"         # Page title (default: "Service Control")
        serviceDescription: "Home Media Server"          # Service name shown on page (default: "Service")
        
        # === AUTO-REDIRECT SETTINGS ===
        autoRedirect: false                               # Auto-redirect when service is online (default: false)
        redirectDelay: "5"                                # Redirect delay in seconds (default: 3)
        
        # === DASHBOARD UI SETTINGS ===
        showPowerOffButton: true                          # Show power-off button (default: true)
        confirmPowerOff: true                             # Require confirmation for power-off (default: true)
        hideRedirectButton: false                         # Hide "Go to Service Anyway" button (default: false)
        
        # === POWER-OFF SETTINGS ===
        powerOffMethod: "ssh"                             # Power-off method: ssh|ipmi|custom (default: ssh)
        powerOffCommand: "sudo shutdown -h now"          # Command to execute (default: "sudo shutdown -h now")
        
        # === SSH CONFIGURATION (for powerOffMethod: "ssh") ===
        sshHost: "192.168.1.100"                          # SSH target host
        sshUser: "admin"                                  # SSH username
        sshKeyPath: "/path/to/ssh/key"                    # SSH private key path (preferred)
        sshPassword: "password"                           # SSH password (use key instead if possible)
        sshPort: "22"                                     # SSH port (default: 22)
        
        # === IPMI CONFIGURATION (for powerOffMethod: "ipmi") ===
        ipmiHost: "192.168.1.100"                         # IPMI/BMC host address
        ipmiUser: "ADMIN"                                 # IPMI username
        ipmiPassword: "password"                          # IPMI password
        
        # === DEBUG SETTINGS ===
        debug: true                                       # Enable detailed logging (default: false)
```

## Power-Off Method Examples

### SSH-Based Shutdown (Recommended)
```yaml
middlewares:
  ssh-power-control:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        
        # SSH power-off configuration
        powerOffMethod: "ssh"
        powerOffCommand: "sudo shutdown -h now"
        sshHost: "192.168.1.100"
        sshUser: "admin"
        sshKeyPath: "/home/traefik/.ssh/id_rsa"
        sshPort: "22"
```

### IPMI-Based Power Control (Enterprise)
```yaml
middlewares:
  ipmi-power-control:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        
        # IPMI power-off configuration
        powerOffMethod: "ipmi"
        ipmiHost: "192.168.1.101"  # BMC/iDRAC address
        ipmiUser: "ADMIN"
        ipmiPassword: "admin123"
```

### Custom Command Power Control
```yaml
middlewares:
  custom-power-control:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        
        # Custom command power-off
        powerOffMethod: "custom"
        powerOffCommand: "/usr/local/bin/my-shutdown-script.sh --server=192.168.1.100"
```

## Alternative Configuration Formats

### TOML Configuration
```toml
[http.middlewares.power-middleware.plugin.traefik-wol]
  # Required settings
  healthCheck = "http://192.168.1.100:3000/health"
  macAddress = "00:11:22:33:44:55"
  
  # Control page settings
  enableControlPage = true
  controlPageTitle = "Server Power Control"
  serviceDescription = "Home Media Server"
  
  # Power-off settings
  showPowerOffButton = true
  confirmPowerOff = true
  powerOffMethod = "ssh"
  powerOffCommand = "sudo shutdown -h now"
  
  # SSH configuration
  sshHost = "192.168.1.100"
  sshUser = "admin"
  sshKeyPath = "/path/to/ssh/key"
  sshPort = "22"
  
  # Auto-redirect settings
  autoRedirect = false
  redirectDelay = "5"
  
  # Wake-on-LAN settings
  ipAddress = "192.168.1.100"
  broadcastAddress = "192.168.1.255"
  networkInterface = "eth0"
  port = "9"
  timeout = "30"
  retryAttempts = "3"
  retryInterval = "5"
  healthCheckInterval = "10"
  debug = true
```

### JSON Configuration
```json
{
  "http": {
    "middlewares": {
      "power-middleware": {
        "plugin": {
          "traefik-wol": {
            "healthCheck": "http://192.168.1.100:3000/health",
            "macAddress": "00:11:22:33:44:55",
            "enableControlPage": true,
            "controlPageTitle": "Server Power Control",
            "serviceDescription": "Home Media Server",
            "showPowerOffButton": true,
            "confirmPowerOff": true,
            "powerOffMethod": "ssh",
            "powerOffCommand": "sudo shutdown -h now",
            "sshHost": "192.168.1.100",
            "sshUser": "admin",
            "sshKeyPath": "/path/to/ssh/key",
            "sshPort": "22",
            "autoRedirect": false,
            "redirectDelay": "5",
            "ipAddress": "192.168.1.100",
            "broadcastAddress": "192.168.1.255",
            "networkInterface": "eth0",
            "port": "9",
            "timeout": "30",
            "retryAttempts": "3",
            "retryInterval": "5",
            "healthCheckInterval": "10",
            "debug": true
          }
        }
      }
    }
  }
}
```

## MAC Address Formats

The plugin accepts various MAC address formats:

```yaml
# Colon-separated (most common)
macAddress: "00:11:22:33:44:55"

# Dash-separated
macAddress: "00-11-22-33-44-55"

# Dot-separated
macAddress: "00.11.22.33.44.55"

# No separators
macAddress: "001122334455"
```

## Container and Network Configuration

The plugin is optimized for containerized environments (Docker, LXC, etc.) and includes enhanced networking features:

### Broadcast Packet Support
- **Automatic Broadcast Discovery**: The plugin automatically detects available network interfaces and calculates broadcast addresses
- **Container Compatibility**: Uses broadcast packets that can traverse container network boundaries
- **Multi-Interface Support**: Sends WOL packets on all available network interfaces for maximum reliability

### Configuration Options

```yaml
# Basic configuration (recommended for most cases)
middlewares:
  wol-middleware:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        # ipAddress is now optional - broadcast will be used automatically

# Advanced configuration for specific networking needs
middlewares:
  wol-advanced:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        ipAddress: "192.168.1.100"              # Optional: Try unicast first
        broadcastAddress: "192.168.1.255"       # Optional: Custom broadcast address
        networkInterface: "eth0"                # Optional: Use specific interface only
        healthCheckInterval: "15"               # Optional: Cache health checks for 15s
```

### Deployment Scenarios

**Docker/Container Deployment:**
```yaml
# Minimal configuration - plugin auto-detects broadcast addresses
middlewares:
  wol-container:
    plugin:
      traefik-wol:
        healthCheck: "http://target-service:3000/health"
        macAddress: "00:11:22:33:44:55"
        healthCheckInterval: "10"  # Reduce health check frequency
```

**LXC/VM Deployment:**
```yaml
# Use specific network interface for better control
middlewares:
  wol-lxc:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        networkInterface: "lxcbr0"  # LXC bridge interface
        broadcastAddress: "192.168.1.255"
```

**Host Network Deployment:**
```yaml
# Traditional deployment with direct network access
middlewares:
  wol-host:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        ipAddress: "192.168.1.100"     # Direct unicast preferred
        healthCheckInterval: "5"       # More frequent checks on stable network
```

## Interactive Power Control Dashboard

The plugin includes a comprehensive web-based dashboard that provides complete power management control over your services.

### Dashboard Configuration

Enable the power control dashboard by setting `enableControlPage: true` and configure power management:

```yaml
middlewares:
  power-dashboard:
    plugin:
      traefik-wol:
        # Basic settings
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        
        # Enable dashboard
        enableControlPage: true
        controlPageTitle: "Home Server Control"
        serviceDescription: "Media Server"
        
        # Dashboard behavior
        autoRedirect: false          # Don't auto-redirect (let user control)
        redirectDelay: "5"           # If auto-redirect enabled, wait 5 seconds
        showPowerOffButton: true     # Show power-off button
        confirmPowerOff: true        # Require confirmation for shutdown
        hideRedirectButton: false    # Show "Go to Service Anyway" button
        
        # Power-off configuration (SSH example)
        powerOffMethod: "ssh"
        powerOffCommand: "sudo shutdown -h now"
        sshHost: "192.168.1.100"
        sshUser: "admin"
        sshKeyPath: "/home/traefik/.ssh/id_rsa"
```

### API Endpoints

When the control page is enabled, the plugin creates REST API endpoints:

- **`/_wol/wake`** (POST): Triggers wake-on-LAN sequence with progress tracking
- **`/_wol/poweroff`** (POST): Initiates secure power-off sequence  
- **`/_wol/status`** (GET): Returns JSON with current status, progress, and operation state
- **`/_wol/redirect`** (GET): Redirects to the original requested URL

## Usage Examples

### Basic Power Management Setup

Apply the power management middleware to your routes for complete lifecycle control:

```yaml
http:
  routers:
    media-server:
      rule: "Host(`media.example.com`)"
      service: media-server
      middlewares:
        - power-management

  services:
    media-server:
      loadBalancer:
        servers:
          - url: "http://192.168.1.100:3000"

  middlewares:
    power-management:
      plugin:
        traefik-wol:
          # Basic required settings
          healthCheck: "http://192.168.1.100:3000/health"
          macAddress: "00:11:22:33:44:55"
          
          # Enable power control dashboard
          enableControlPage: true
          controlPageTitle: "Media Server Control"
          serviceDescription: "Home Media Server"
          
          # Configure power-off via SSH
          showPowerOffButton: true
          confirmPowerOff: true
          powerOffMethod: "ssh"
          powerOffCommand: "sudo shutdown -h now"
          sshHost: "192.168.1.100"
          sshUser: "admin"
          sshKeyPath: "/home/traefik/.ssh/media_server_key"
          
          # Optional: disable auto-redirect for manual control
          autoRedirect: false
          debug: true
```

### Advanced Power Management Scenarios

#### Enterprise Server with IPMI
```yaml
middlewares:
  enterprise-power:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.50:8080/health"
        macAddress: "AA:BB:CC:DD:EE:FF"
        
        enableControlPage: true
        controlPageTitle: "Enterprise Server Control"
        serviceDescription: "Production Database Server"
        
        # IPMI-based power control
        powerOffMethod: "ipmi"
        ipmiHost: "192.168.1.51"  # BMC/iDRAC address
        ipmiUser: "ADMIN"
        ipmiPassword: "${IPMI_PASSWORD}"  # Use environment variable
        
        # Security settings
        confirmPowerOff: true
        hideRedirectButton: true  # Hide direct access for security
```

#### Home Lab with Custom Scripts
```yaml
middlewares:
  homelab-power:
    plugin:
      traefik-wol:
        healthCheck: "http://homelab.local:3000/health"
        macAddress: "12:34:56:78:9A:BC"
        
        enableControlPage: true
        controlPageTitle: "Home Lab Control"
        serviceDescription: "Development Environment"
        
        # Custom power management script
        powerOffMethod: "custom"
        powerOffCommand: "/home/user/scripts/graceful-shutdown.sh --target=homelab"
        
        # User-friendly settings
        autoRedirect: true
        redirectDelay: "3"
        confirmPowerOff: false  # Skip confirmation for home use
```