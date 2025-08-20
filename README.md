# Traefik Power Management Plugin

A comprehensive power management middleware plugin for Traefik that provides complete lifecycle control over your services - from automatic wake-up via Wake-on-LAN to secure remote shutdown via SSH, IPMI, or custom commands.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Basic Configuration](#basic-configuration)
- [Interactive Dashboard](#interactive-dashboard)
- [Usage Examples](#usage-examples)
- [Testing](#testing)
- [Documentation](#documentation)
- [Support](#support)

## Features

### 🔋 Complete Power Lifecycle Management
- **Wake-on-LAN (WOL)**: Automatic service wake-up via magic packets with multi-interface broadcasting
- **Remote Shutdown**: Power-off capabilities via custom scripts (SSH, IPMI, or other methods)
- **Power State Monitoring**: Real-time service health monitoring with smart caching
- **Bi-directional Control**: Both wake-up and shutdown operations with progress tracking

### 🎛️ Interactive Dashboard
- **Web-based Control Panel**: Beautiful, responsive interface for manual power management
- **Real-time Status Updates**: Live progress indicators during wake/shutdown operations
- **Security Controls**: Confirmation dialogs, credential protection, and secure command execution
- **Mobile-Optimized**: Touch-friendly interface that works on all devices

### 🔧 Flexible Power Management
- **Custom Scripts**: Execute any power management script or command externally
- **SSH Support**: Via custom scripts with full authentication control
- **IPMI Support**: Via custom scripts for enterprise hardware management
- **Webhook Integration**: Custom scripts can trigger webhooks, APIs, or any shutdown method

### 🌐 Network & Container Optimization
- **Container-Native**: Enhanced support for Docker, LXC, and containerized environments
- **Multi-Interface Broadcasting**: Automatic network discovery and broadcast packet delivery
- **Smart Health Caching**: Configurable health check intervals with connection pooling

## Quick Start

### 1. Install the Plugin

Add to your Traefik static configuration:

```yaml
experimental:
  plugins:
    traefik-power-management:
      moduleName: "github.com/ottup/traefik-power-management"
      version: "v3.1.0"
```

### 2. Configure the Middleware

```yaml
middlewares:
  power-control:
    plugin:
      traefik-power-management:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
```

### 3. Apply to Your Routes

```yaml
http:
  routers:
    my-service:
      rule: "Host(`service.example.com`)"
      service: my-service
      middlewares:
        - power-control

  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://192.168.1.100:3000"
```

### 4. Test the Setup

1. Put your target device to sleep
2. Access your service through Traefik
3. Watch the magic happen! ✨

## Installation

### Method 1: Official Plugin Catalog (Recommended)

> **Note**: Use v3.1.0 which removes os/exec dependency for full Yaegi compatibility.

Add the plugin to your Traefik static configuration:

```yaml
experimental:
  plugins:
    traefik-power-management:
      moduleName: "github.com/ottup/traefik-power-management"
      version: "v3.1.0"
```

### Method 2: Local Development Plugin

For development or private deployment:

```bash
# Create plugin directory
mkdir -p ./plugins-local/src/github.com/ottup/traefik-power-management

# Clone the repository
git clone https://github.com/ottup/traefik-power-management.git ./plugins-local/src/github.com/ottup/traefik-power-management
```

Configure as local plugin:

```yaml
experimental:
  localPlugins:
    traefik-power-management:
      moduleName: "github.com/ottup/traefik-power-management"
```

## Basic Configuration

### Minimal Configuration

```yaml
middlewares:
  power-middleware:
    plugin:
      traefik-power-management:
        healthCheck: "http://192.168.1.100:3000/health"  # Required
        macAddress: "00:11:22:33:44:55"                   # Required
```

### Common Configuration Options

```yaml
middlewares:
  power-management:
    plugin:
      traefik-power-management:
        # Required settings
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        
        # Enable interactive dashboard
        enableControlPage: true
        controlPageTitle: "Server Control"
        serviceDescription: "Media Server"
        
        # Wake-on-LAN settings
        timeout: "30"                    # Wake timeout in seconds
        retryAttempts: "3"               # Number of retry attempts
        healthCheckInterval: "10"        # Health check cache interval
        
        # Debug logging
        debug: true
```

### Power Management Configuration

#### Custom Script-Based Shutdown
```yaml
middlewares:
  custom-power:
    plugin:
      traefik-power-management:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        
        # Custom power-off script
        powerOffCommand: "/usr/local/bin/ssh-shutdown.sh"
```

**Example SSH Shutdown Script** (`/usr/local/bin/ssh-shutdown.sh`):
```bash
#!/bin/bash
ssh -i /home/traefik/.ssh/id_rsa -o StrictHostKeyChecking=no admin@192.168.1.100 "sudo shutdown -h now"
```

**Example IPMI Shutdown Script** (`/usr/local/bin/ipmi-shutdown.sh`):
```bash
#!/bin/bash
ipmitool -I lanplus -H 192.168.1.101 -U ADMIN -P admin123 chassis power off
```

### MAC Address Formats

The plugin accepts various MAC address formats:

```yaml
macAddress: "00:11:22:33:44:55"  # Colon-separated (most common)
macAddress: "00-11-22-33-44-55"  # Dash-separated
macAddress: "00.11.22.33.44.55"  # Dot-separated
macAddress: "001122334455"       # No separators
```

## Interactive Dashboard

Enable the interactive power control dashboard for manual power management:

```yaml
middlewares:
  dashboard-power:
    plugin:
      traefik-power-management:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        
        # Enable dashboard
        enableControlPage: true
        controlPageTitle: "Home Server Control"
        serviceDescription: "Media Server"
        
        # Dashboard behavior
        showPowerOffButton: true      # Show power-off button
        confirmPowerOff: true         # Require confirmation
        autoRedirect: false           # Manual control
        
        # Power-off configuration
        powerOffCommand: "/usr/local/bin/ssh-shutdown.sh"
```

### Dashboard Features

- **🔋 Power Controls**: Wake-up and power-off buttons with progress tracking
- **📱 Mobile-Friendly**: Responsive design that works on all devices
- **🔒 Security**: Confirmation dialogs and configurable button visibility
- **⚡ Real-time Updates**: Live progress indicators during operations

### API Endpoints

When the dashboard is enabled, these endpoints are available:

- `/_wol/wake` (POST): Trigger wake-on-LAN
- `/_wol/poweroff` (POST): Initiate power-off
- `/_wol/status` (GET): Get current status
- `/_wol/redirect` (GET): Redirect to service

## Usage Examples

### Home Media Server
```yaml
middlewares:
  media-power:
    plugin:
      traefik-power-management:
        healthCheck: "http://media-server.local:8080/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        controlPageTitle: "Media Server"
        serviceDescription: "Plex Media Server"
        autoRedirect: true
        redirectDelay: "3"
```

### Enterprise Database Server
```yaml
middlewares:
  db-power:
    plugin:
      traefik-power-management:
        healthCheck: "https://db.internal:5432/health"
        macAddress: "AA:BB:CC:DD:EE:FF"
        enableControlPage: true
        controlPageTitle: "Database Server"
        
        # IPMI via custom script for enterprise hardware
        powerOffCommand: "/usr/local/bin/db-ipmi-shutdown.sh"
        confirmPowerOff: true
        hideRedirectButton: true  # Security
```

### Development Environment
```yaml
middlewares:
  dev-power:
    plugin:
      traefik-power-management:
        healthCheck: "http://dev.local:3000/health"
        macAddress: "12:34:56:78:9A:BC"
        enableControlPage: true
        
        # Custom shutdown script
        powerOffCommand: "/scripts/dev-shutdown.sh"
        confirmPowerOff: false  # Skip confirmation for dev
        debug: true            # Verbose logging
```

## Testing

### 1. Verify Plugin Loading
```bash
# Check Traefik logs
docker logs traefik | grep -i "traefik-wol"
```

### 2. Test Health Endpoint
```bash
# Test your health check URL
curl -v http://192.168.1.100:3000/health
```

### 3. Test Wake-on-LAN
```bash
# Check WOL is enabled on target device
sudo ethtool eth0 | grep Wake

# Test manual WOL
wakeonlan 00:11:22:33:44:55
```

### 4. Test Power Management Scripts
```bash
# Test your custom shutdown script
/usr/local/bin/ssh-shutdown.sh

# Test IPMI script
/usr/local/bin/ipmi-shutdown.sh

# Test webhook script
curl -X POST http://your-power-management-api/shutdown
```

### 5. Monitor Operation
```bash
# Watch Traefik logs for plugin activity
docker logs -f traefik | grep "WOL Plugin"
```

## Documentation

For detailed information, see our comprehensive documentation:

- **[Configuration Reference](CONFIGURATION.md)** - Complete configuration options and examples
- **[Deployment Guide](DEPLOYMENT.md)** - Production deployment and security best practices
- **[Troubleshooting Guide](TROUBLESHOOTING.md)** - Common issues and solutions
- **[Changelog](CHANGELOG.md)** - Version history and breaking changes
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the project

## Requirements

- **Traefik**: v2.3+ (plugin support required)
- **Target Device**: Must support Wake-on-LAN
- **Network Access**: UDP port 9 (or configured port)
- **Optional**: Custom scripts for power-off (can use SSH, IPMI, webhooks, or any method)

## Support

- **🐛 Issues**: Report bugs via [GitHub Issues](https://github.com/ottup/traefik-power-management/issues)
- **💬 Discussions**: Ask questions in [GitHub Discussions](https://github.com/ottup/traefik-power-management/discussions)
- **📚 Documentation**: Comprehensive guides in the repository
- **🏘️ Community**: Join Traefik community for general support

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- **Traefik Team**: For the excellent reverse proxy and plugin system
- **Go Community**: For comprehensive standard library and ecosystem
- **Contributors**: Thanks to all contributors who help improve this plugin

---

**Ready to get started?** Check out our [Configuration Reference](CONFIGURATION.md) for detailed setup instructions!