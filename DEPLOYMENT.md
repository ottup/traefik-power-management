# Deployment & Security Guide

This document provides comprehensive guidance on deploying the Traefik Power Management Plugin in production environments with security best practices.

## Table of Contents

- [Security Considerations](#security-considerations)
- [Production Deployment](#production-deployment)
- [Performance Optimization](#performance-optimization)
- [Enterprise Deployment Patterns](#enterprise-deployment-patterns)
- [Operating Modes Comparison](#operating-modes-comparison)

## Security Considerations

### 🔒 Authentication & Credentials

#### SSH Key Authentication
Always prefer SSH keys over passwords for better security:

```yaml
# Secure SSH configuration
sshKeyPath: "/home/traefik/.ssh/server_key"
# Avoid: sshPassword: "plaintext_password"
```

**Key Management Best Practices:**
- Ensure SSH keys have correct permissions (600) and ownership
- Use dedicated keys for power management operations
- Regularly rotate SSH keys and IPMI passwords

```bash
# Set correct SSH key permissions
chmod 600 /path/to/ssh/key
chmod 700 ~/.ssh
```

#### Credential Storage
Use environment variables or secrets management for sensitive data:

```yaml
ipmiPassword: "${IPMI_PASSWORD}"  # Use env var instead of plain text
sshPassword: "${SSH_PASSWORD}"   # If password auth is required
```

### 🌐 Network Security

#### Network Segmentation
- Deploy power management within trusted network segments
- Restrict access to BMC/SSH interfaces from Traefik containers
- Consider VPN requirements for remote power management

#### Firewall Configuration
```bash
# Example firewall rules for WOL and SSH
iptables -A OUTPUT -p udp --dport 9 -j ACCEPT     # WOL packets
iptables -A OUTPUT -p tcp --dport 22 -j ACCEPT    # SSH access
iptables -A OUTPUT -p udp --dport 623 -j ACCEPT   # IPMI over LAN
```

#### Access Control
Limit dashboard access through Traefik's authentication middleware:

```yaml
http:
  routers:
    protected-service:
      middlewares:
        - auth-middleware    # Add authentication first
        - power-middleware   # Then power management
```

### 🛡️ Dashboard Security

#### Confirmation Controls
Always enable power-off confirmation in production:

```yaml
confirmPowerOff: true  # Require user confirmation
```

#### Button Visibility
Hide sensitive buttons in high-security environments:

```yaml
hideRedirectButton: true      # Remove direct access bypass
showPowerOffButton: false     # Disable power-off in critical systems
```

#### Complete Security Example
```yaml
middlewares:
  secure-power-control:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        
        # Security settings
        enableControlPage: true
        confirmPowerOff: true
        hideRedirectButton: true
        showPowerOffButton: true  # But require confirmation
        
        # Custom power-off script configuration
        powerOffCommand: "/usr/local/bin/secure-shutdown.sh"
        
        # Minimal auto-redirect
        autoRedirect: false
        debug: false  # Disable debug in production
```

## Production Deployment

### 🚀 Environment Preparation

#### Custom Script Dependencies
Since power-off functionality uses external scripts, ensure your shutdown automation is set up:

```dockerfile
# Example: Add script monitoring to Traefik container
COPY scripts/monitor-poweroff.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/monitor-poweroff.sh

# Install tools for your specific shutdown method
RUN apt-get update && apt-get install -y \
    openssh-client \  # For SSH scripts
    ipmitool \        # For IPMI scripts  
    curl \            # For webhook scripts
    && apt-get clean && rm -rf /var/lib/apt/lists/*
```

#### Container Networking
Verify network access to target systems:

```yaml
# Docker Compose example
services:
  traefik:
    image: traefik:latest
    container_name: traefik
    networks:
      - management_network  # Network with access to target systems
    volumes:
      - ./plugins-local:/plugins-local:ro
      - /host/ssh/keys:/home/traefik/.ssh:ro  # Mount SSH keys securely
```

#### SSH Key Management
Mount SSH keys securely in containers:

```yaml
# Docker Compose volume mount
volumes:
  - /host/ssh/keys:/home/traefik/.ssh:ro

# Kubernetes secret mount
volumeMounts:
  - name: ssh-keys
    mountPath: /home/traefik/.ssh
    readOnly: true
volumes:
  - name: ssh-keys
    secret:
      secretName: power-management-keys
      defaultMode: 0600
```

### ⚙️ Configuration Management

#### Environment Variables
Use environment variables for sensitive configuration:

```yaml
# docker-compose.yml
services:
  traefik:
    environment:
      - IPMI_PASSWORD=your-secure-password
      - SSH_PASSWORD=your-ssh-password
    # ... rest of config
```

#### Configuration Validation
Test your custom shutdown scripts before production deployment:

```bash
# Test your custom shutdown script directly
/usr/local/bin/shutdown-script.sh

# Example: Test SSH-based script
ssh -i /path/to/key user@target "sudo shutdown -h now"

# Example: Test IPMI-based script
ipmitool -I lanplus -H target-bmc -U admin -P password chassis power off

# Example: Test webhook-based script  
curl -X POST https://automation-api.internal/shutdown

# Note: Plugin only logs commands - external execution required
```

#### Health Check Reliability
Ensure health endpoints are stable and fast:

```yaml
# Optimize health check settings
healthCheckInterval: "10"  # 10 seconds for production
timeout: "45"             # Allow time for slow systems
retryAttempts: "5"        # More retries for reliability
```

### 📊 Monitoring & Alerting

#### Power Event Logging
Monitor wake and shutdown operations:

```yaml
# Enable debug logging for monitoring
debug: true  # In development/staging

# Use structured logging in production
debug: false
```

#### Failed Operations Alerting
Set up alerts for repeated power management failures:

```bash
# Example log monitoring (using promtail/loki)
grep "WOL Plugin.*failed" /var/log/traefik.log | tail -10
```

#### Security Event Monitoring
- Log authentication failures and unauthorized access attempts
- Monitor power operations outside business hours
- Track repeated failed wake attempts

## Performance Optimization

### ⚡ Health Check Tuning

#### Cache Intervals
Balance responsiveness vs. system load:

```yaml
healthCheckInterval: "10"  # 10 seconds for most use cases
```

**Guidelines:**
- **High-traffic services**: 15-30 seconds
- **Low-traffic services**: 5-10 seconds  
- **Development**: 5 seconds
- **Production**: 10-15 seconds

#### Connection Pooling
Enabled automatically in v3.0.0 for better performance:
- HTTP keep-alive connections for health checks
- Reduced connection overhead
- Better resource utilization

#### Timeout Configuration
Set appropriate timeouts for your environment:

```yaml
timeout: "45"        # Increase for slow-booting systems
retryAttempts: "5"   # More retries for unreliable networks
retryInterval: "5"   # Balance between responsiveness and system load
```

### 🌐 Network Optimization

#### Broadcast vs Unicast
Use broadcast for container environments:

```yaml
# Container-optimized (recommended)
middlewares:
  wol-container:
    plugin:
      traefik-wol:
        healthCheck: "http://target:3000/health"
        macAddress: "00:11:22:33:44:55"
        # No ipAddress specified - uses broadcast automatically

# Direct network access (traditional)
middlewares:
  wol-direct:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        ipAddress: "192.168.1.100"  # Direct unicast
```

#### Network Interface Selection
Specify interfaces in complex network setups:

```yaml
networkInterface: "eth0"              # Use specific interface
broadcastAddress: "192.168.1.255"    # Custom broadcast address
```

### 💾 Resource Management

#### Memory Usage
Plugin optimizations:
- Connection pooling for health checks
- Optimized health check caching
- Minimal memory allocation for magic packets

#### CPU Usage  
- Non-blocking operations prevent CPU blocking
- Background processing for power operations
- Efficient mutex handling for concurrent operations

#### Concurrent Operations
Plugin handles multiple simultaneous operations safely:
- Thread-safe operation tracking
- Proper mutex handling for shared state
- Race condition prevention

## Enterprise Deployment Patterns

### 🏢 High Availability Setup

```yaml
# Primary server power management
middlewares:
  primary-power:
    plugin:
      traefik-wol:
        healthCheck: "http://primary.internal:8080/health"
        macAddress: "AA:BB:CC:DD:EE:FF"
        powerOffCommand: "/usr/local/bin/primary-ipmi-shutdown.sh"
        confirmPowerOff: true
        enableControlPage: true

# Backup server power management
middlewares:
  backup-power:
    plugin:
      traefik-wol:
        healthCheck: "http://backup.internal:8080/health"
        macAddress: "11:22:33:44:55:66"
        powerOffCommand: "/usr/local/bin/backup-ssh-shutdown.sh"
```

### 🔄 Environment-Specific Configurations

#### Development Environment
```yaml
# Development (relaxed security, verbose logging)
middlewares:
  dev-power:
    plugin:
      traefik-wol:
        healthCheck: "http://dev-server:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        confirmPowerOff: false     # Skip confirmation for faster development
        autoRedirect: true         # Auto-redirect enabled
        redirectDelay: "1"         # Fast redirect
        debug: true               # Verbose logging
        healthCheckInterval: "5"   # Frequent health checks
```

#### Staging Environment
```yaml
# Staging (production-like with some convenience features)
middlewares:
  staging-power:
    plugin:
      traefik-wol:
        healthCheck: "http://staging-server:3000/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        confirmPowerOff: true      # Require confirmation
        autoRedirect: false        # Manual control
        debug: true               # Keep debug for testing
        healthCheckInterval: "10"
        
        # Custom script configuration
        powerOffCommand: "/usr/local/bin/staging-shutdown.sh"
```

#### Production Environment
```yaml
# Production (strict security, minimal logging)
middlewares:
  prod-power:
    plugin:
      traefik-wol:
        healthCheck: "https://prod-server.internal/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: true
        confirmPowerOff: true      # Always require confirmation
        autoRedirect: false        # No auto-redirect in production
        hideRedirectButton: true   # No bypass option
        debug: false              # Minimal logging for security
        healthCheckInterval: "15"  # Reduced frequency
        
        # Custom script for enterprise hardware
        powerOffCommand: "/usr/local/bin/prod-ipmi-shutdown.sh"
```

### 🔒 Security Tiers

#### Tier 1: Development/Testing
- No confirmation dialogs
- Auto-redirect enabled
- Debug logging enabled
- Password authentication allowed

#### Tier 2: Staging/Internal
- Confirmation required for power-off
- Manual redirect control
- Debug logging for troubleshooting
- SSH key preferred

#### Tier 3: Production/Critical
- All confirmations required
- No bypass options
- Minimal logging
- SSH keys mandatory
- Network segmentation required

## Operating Modes Comparison

| Feature | Auto-Wake Mode (default) | Power Control Dashboard |
|---------|-------------------------|------------------------|
| Service Down Detection | Automatically sends WOL packets | Shows interactive control interface |
| User Interaction | None required | Manual "Turn On Service" button |
| Power-off Capability | Not available | "Power Off" button with confirmation |
| Progress Feedback | Server logs only | Real-time web interface with progress bars |
| Direct Access | Not available | "Go to Service Anyway" option (configurable) |
| Auto-redirect | N/A | Optional with configurable delay |
| Mobile Access | N/A | Fully responsive mobile interface |
| Security Controls | None | Confirmation dialogs, conditional button visibility |
| Production Suitability | High (fire-and-forget) | High (with proper security settings) |
| User Control | None | Complete manual control |

### Choosing the Right Mode

**Use Auto-Wake Mode when:**
- Service downtime is rare and brief
- No user interaction is desired
- Maximum automation is preferred
- Security requirements are minimal

**Use Power Control Dashboard when:**
- Users need manual control over power state
- Power-off capability is required
- Real-time feedback is important
- Mobile access is needed
- Security controls are required

### Hybrid Deployment
You can deploy both modes for different services:

```yaml
# Critical service - auto-wake only
middlewares:
  critical-auto:
    plugin:
      traefik-wol:
        healthCheck: "http://critical-service/health"
        macAddress: "00:11:22:33:44:55"
        enableControlPage: false  # Pure auto-wake
        
# User service - full dashboard
middlewares:
  user-dashboard:
    plugin:
      traefik-wol:
        healthCheck: "http://user-service/health" 
        macAddress: "AA:BB:CC:DD:EE:FF"
        enableControlPage: true   # Full dashboard
        confirmPowerOff: true
```