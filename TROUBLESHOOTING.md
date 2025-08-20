# Troubleshooting Guide

This document provides comprehensive troubleshooting guidance for the Traefik Power Management Plugin.

## Table of Contents

- [Common Issues](#common-issues)
- [Power Management Issues](#power-management-issues)
- [Dashboard/Control Page Issues](#dashboardcontrol-page-issues)
- [Network Issues](#network-issues)
- [Debug Mode](#debug-mode)
- [Log Examples](#log-examples)
- [Testing Steps](#testing-steps)

## Common Issues

### Plugin Not Loading

**Symptoms:**
- Traefik starts but middleware is not available
- Error: "plugin not found" or "middleware not found"
- Plugin not listed in Traefik dashboard

**Causes & Solutions:**

#### Incorrect Module Name
- **Cause**: Wrong `moduleName` in configuration
- **Solution**: Verify `moduleName` matches exactly: `github.com/ottup/traefik-power-management`

```yaml
# Correct configuration
experimental:
  plugins:
    traefik-power-management:
      moduleName: "github.com/ottup/traefik-power-management"  # Must match exactly
      version: "v3.0.0"
```

#### Plugin Not Reloaded
- **Cause**: Traefik not restarted after configuration changes
- **Solution**: Restart Traefik after adding/updating plugins

```bash
# Docker Compose
docker-compose restart traefik

# Docker
docker restart traefik

# Systemd
sudo systemctl restart traefik
```

#### Version Issues
- **Cause**: Invalid or non-existent version specified
- **Solution**: Use a valid released version tag

```bash
# Check available versions
curl -s https://api.github.com/repos/ottup/traefik-power-management/tags | jq '.[].name'
```

### Health Check Failures

**Symptoms:**
- Plugin shows service as always down
- Wake attempts trigger even when service is running
- Health check timeouts in logs

**Causes & Solutions:**

#### Incorrect Health Check URL
- **Solution**: Test health endpoint manually

```bash
# Test the exact URL from your configuration
curl -v http://192.168.1.100:3000/health

# Check response code and content
curl -I http://192.168.1.100:3000/health
```

#### Endpoint Returns Non-2xx Status
- **Cause**: Health endpoint returns 404, 500, or other error codes
- **Solution**: Ensure endpoint returns HTTP 2xx status codes

```bash
# Check actual response code
curl -w "%{http_code}\n" -o /dev/null -s http://192.168.1.100:3000/health
```

#### Network Connectivity Issues
- **Solution**: Test from Traefik container perspective

```bash
# Test from within Traefik container
docker exec traefik curl http://192.168.1.100:3000/health

# Check container networking
docker exec traefik ping 192.168.1.100
```

#### Health Check Timeout
- **Solution**: Increase health check timeout

```yaml
middlewares:
  power-middleware:
    plugin:
      traefik-power-management:
        healthCheck: "http://192.168.1.100:3000/health"
        timeout: "60"  # Increase from default 30 seconds
```

### Wake-on-LAN Not Working

**Symptoms:**
- Magic packets sent but device doesn't wake
- No network activity on target device
- WOL works with other tools but not with plugin

**Causes & Solutions:**

#### WOL Not Enabled on Target Device

**BIOS/UEFI Settings:**
```
1. Enter BIOS/UEFI setup during boot
2. Navigate to Power Management or Advanced settings
3. Enable "Wake on LAN" or "Wake on PCI-E"
4. Save settings and reboot
```

**Network Adapter Settings (Windows):**
```
1. Device Manager → Network Adapters
2. Right-click network adapter → Properties
3. Power Management tab:
   - ✓ Allow this device to wake the computer
   - ✓ Only allow a magic packet to wake the computer
4. Advanced tab:
   - Wake on Magic Packet: Enabled
   - Wake on Pattern Match: Enabled (optional)
```

**Network Adapter Settings (Linux):**
```bash
# Check current WOL status
sudo ethtool eth0 | grep Wake

# Enable WOL if disabled
sudo ethtool -s eth0 wol g

# Make permanent (Ubuntu/Debian)
echo 'ethtool -s eth0 wol g' >> /etc/rc.local

# Or create systemd service for persistence
sudo systemctl enable wakeonlan
```

#### Test WOL Manually
```bash
# Linux/Mac - install wakeonlan
sudo apt-get install wakeonlan  # Ubuntu/Debian
brew install wakeonlan          # macOS

# Test manual wake
wakeonlan 00:11:22:33:44:55

# Windows - use wolcmd or similar tool
wolcmd 00:11:22:33:44:55 192.168.1.100
```

#### MAC Address Issues
- **Cause**: Wrong or improperly formatted MAC address
- **Solution**: Verify MAC address format and value

```bash
# Find correct MAC address
# Linux
ip link show
cat /sys/class/net/eth0/address

# Windows
ipconfig /all
getmac

# macOS  
ifconfig
networksetup -getmacaddress en0
```

#### Network Firewall Blocking
- **Cause**: Firewall blocking UDP port 9
- **Solution**: Allow WOL traffic

```bash
# Linux iptables
sudo iptables -A INPUT -p udp --dport 9 -j ACCEPT
sudo iptables -A OUTPUT -p udp --dport 9 -j ACCEPT

# Windows Firewall
netsh advfirewall firewall add rule name="Wake-on-LAN" dir=in action=allow protocol=UDP localport=9
```

## Power Management Issues

### SSH Power-Off Failures

**Symptoms:**
- Power-off button doesn't work
- SSH authentication errors
- Permission denied errors

**Causes & Solutions:**

#### SSH Authentication Issues

**Test SSH Connection:**
```bash
# Test with the exact parameters from your configuration
ssh -i /path/to/key -p 22 user@192.168.1.100 "sudo shutdown -h now"

# Test basic connectivity
ssh -i /path/to/key user@192.168.1.100 "echo 'Connection successful'"

# Debug SSH connection
ssh -vvv -i /path/to/key user@192.168.1.100
```

**SSH Key Issues:**
```bash
# Check key permissions
ls -la /path/to/ssh/key
# Should show: -rw------- (600)

# Fix permissions
chmod 600 /path/to/ssh/key
chmod 700 ~/.ssh

# Verify key format
ssh-keygen -l -f /path/to/ssh/key
```

**SSH Client Missing:**
```dockerfile
# Add to Traefik Dockerfile
RUN apt-get update && apt-get install -y openssh-client

# For password authentication
RUN apt-get install -y sshpass
```

#### Permission Issues

**Sudo Configuration:**
```bash
# Allow user to shutdown without password prompt
# Add to /etc/sudoers on target system
username ALL=(ALL) NOPASSWD: /sbin/shutdown, /sbin/poweroff, /sbin/halt

# Test sudo permissions
ssh user@target "sudo -n shutdown --help"
```

**Alternative Commands:**
```yaml
# Try different shutdown commands
powerOffCommand: "sudo poweroff"           # systemd systems
powerOffCommand: "sudo halt -p"            # older systems  
powerOffCommand: "shutdown -h now"         # if sudo not needed
powerOffCommand: "systemctl poweroff"      # direct systemctl
```

### IPMI Power-Off Failures

**Symptoms:**
- IPMI commands fail
- Authentication errors to BMC
- Network timeouts to IPMI interface

**Causes & Solutions:**

#### IPMI Tool Missing
```bash
# Install ipmitool
apt-get install ipmitool        # Ubuntu/Debian
yum install ipmitool           # CentOS/RHEL
dnf install ipmitool           # Fedora

# Test installation
which ipmitool
ipmitool -V
```

#### Test IPMI Connection
```bash
# Test basic connectivity
ipmitool -I lanplus -H 192.168.1.101 -U ADMIN -P password chassis power status

# Test power control
ipmitool -I lanplus -H 192.168.1.101 -U ADMIN -P password chassis power cycle

# Debug connection issues
ipmitool -I lanplus -H 192.168.1.101 -U ADMIN -P password -v chassis power status
```

#### Common IPMI Issues

**Wrong Interface Type:**
```yaml
# Plugin uses lanplus interface by default
# If your BMC requires different interface, create custom command:
powerOffMethod: "custom"
powerOffCommand: "ipmitool -I lan -H 192.168.1.101 -U ADMIN -P password chassis power off"
```

**Network Access:**
```bash
# Test BMC network connectivity
ping 192.168.1.101
telnet 192.168.1.101 623  # IPMI port

# Check from Traefik container
docker exec traefik ping 192.168.1.101
```

**Authentication:**
- Verify IPMI username and password
- Check if account is enabled in BMC
- Ensure account has power control privileges

### Custom Command Failures

**Symptoms:**
- Custom commands don't execute
- "Command not found" errors
- Permission denied errors

**Causes & Solutions:**

#### Command Not Found
```bash
# Test command availability
which your-command
/path/to/your/command --help

# Check from Traefik container context
docker exec traefik which your-command
docker exec traefik /path/to/your/command --test
```

#### Permission Issues
```bash
# Check command permissions
ls -la /path/to/your/command

# Make executable if needed
chmod +x /path/to/your/command

# Test execution
/path/to/your/command --test
```

#### Script Dependencies
```bash
# Check script dependencies
# For shell scripts, verify shebang line
head -1 /path/to/script.sh  # Should show #!/bin/bash or similar

# Check for required utilities
ldd /path/to/binary        # for compiled binaries
bash -n /path/to/script.sh # syntax check for bash scripts
```

## Dashboard/Control Page Issues

### Template Errors

**Symptoms:**
- Control page doesn't load
- HTML rendering errors
- JavaScript errors in browser console

**Causes & Solutions:**

#### Template Rendering Errors
- **Solution**: Enable debug logging and check Traefik logs

```yaml
middlewares:
  power-middleware:
    plugin:
      traefik-power-management:
        debug: true  # Enable verbose logging
```

#### JavaScript Errors
- **Solution**: Check browser developer tools

```
1. Open browser Developer Tools (F12)
2. Go to Console tab
3. Refresh the page
4. Look for JavaScript errors
5. Check Network tab for failed requests
```

### API Endpoint Issues

**Symptoms:**
- Buttons don't respond
- Status updates don't work
- 404 errors on `/_wol/*` endpoints

**Causes & Solutions:**

#### API Endpoints Not Available
- **Cause**: `enableControlPage` not set to `true`
- **Solution**: Enable control page

```yaml
middlewares:
  power-middleware:
    plugin:
      traefik-power-management:
        enableControlPage: true  # Required for dashboard
```

#### Network Request Failures
```bash
# Test API endpoints directly
curl -X POST http://your-service/_wol/wake
curl http://your-service/_wol/status
```

## Network Issues

### Container/LXC WOL Issues

**Symptoms:**
- WOL works from host but not from container
- Magic packets not reaching target
- Network isolation preventing WOL delivery

**Causes & Solutions:**

#### Network Isolation
- **Solution**: Use broadcast mode instead of unicast

```yaml
# Remove ipAddress to force broadcast mode
middlewares:
  wol-container:
    plugin:
      traefik-power-management:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        # ipAddress: "192.168.1.100"  # Comment out or remove
        broadcastAddress: "192.168.1.255"  # Add specific broadcast
```

#### Container Network Configuration
```bash
# Test broadcast capability from container
docker exec traefik ping -c 1 192.168.1.255

# Check container network settings
docker exec traefik ip route
docker exec traefik ip addr show
```

#### Docker Network Settings
```yaml
# Docker Compose - use host network for WOL
services:
  traefik:
    network_mode: host  # Direct host network access

# Or configure bridge network properly
networks:
  default:
    driver: bridge
    driver_opts:
      com.docker.network.enable_ipv6: "false"
```

### Firewall Issues

**Symptoms:**
- WOL packets blocked by firewall
- Network timeouts
- Works locally but not remotely

**Causes & Solutions:**

#### UDP Port 9 Blocked
```bash
# Test UDP port access
nmap -sU -p 9 192.168.1.100

# Allow WOL traffic
# Linux iptables
iptables -A INPUT -p udp --dport 9 -j ACCEPT
iptables -A OUTPUT -p udp --dport 9 -j ACCEPT

# UFW (Ubuntu)
ufw allow 9/udp

# Windows
netsh advfirewall firewall add rule name="WOL" dir=in action=allow protocol=UDP localport=9
```

#### Broadcast Traffic Blocked
```bash
# Some networks block broadcast traffic
# Test with specific broadcast address
middlewares:
  wol-middleware:
    plugin:
      traefik-power-management:
        broadcastAddress: "192.168.1.255"  # Try directed broadcast
```

## Debug Mode

Enable comprehensive debug logging for detailed troubleshooting:

```yaml
middlewares:
  wol-middleware:
    plugin:
      traefik-power-management:
        debug: true  # Enable verbose logging
```

Debug logs include:
- Health check attempts and results with full HTTP details
- Wake-on-LAN packet transmission details and network interface information
- Service wake-up monitoring progress with timestamps
- Power-off operation details and command execution
- Error details and retry attempts with full context
- Network discovery and interface detection
- Template rendering and API endpoint activity

## Log Examples

### Successful Wake Operation
```
WOL Plugin [wol-middleware]: Checking health for http://192.168.1.100:3000/health
WOL Plugin [wol-middleware]: Health check failed: dial tcp 192.168.1.100:3000: connect: connection refused
WOL Plugin [wol-middleware]: Service unhealthy, attempting to wake 00:11:22:33:44:55
WOL Plugin [wol-middleware]: Wake attempt 1/3
WOL Plugin [wol-middleware]: Sending magic packet to 00:11:22:33:44:55 via broadcast (192.168.1.255:9)
WOL Plugin [wol-middleware]: Magic packet sent successfully
WOL Plugin [wol-middleware]: Waiting for service to come online (timeout: 30s)
WOL Plugin [wol-middleware]: Health check status: 200 (healthy: true)
WOL Plugin [wol-middleware]: Service is now online after 15.2 seconds
```

### Failed Wake Operation
```
WOL Plugin [wol-middleware]: Checking health for http://192.168.1.100:3000/health
WOL Plugin [wol-middleware]: Health check failed: dial tcp 192.168.1.100:3000: i/o timeout
WOL Plugin [wol-middleware]: Service unhealthy, attempting to wake 00:11:22:33:44:55
WOL Plugin [wol-middleware]: Wake attempt 1/3
WOL Plugin [wol-middleware]: Sending magic packet to 00:11:22:33:44:55 via broadcast (192.168.1.255:9)
WOL Plugin [wol-middleware]: Magic packet sent successfully
WOL Plugin [wol-middleware]: Waiting for service to come online (timeout: 30s)
WOL Plugin [wol-middleware]: Wake attempt 2/3
WOL Plugin [wol-middleware]: Wake attempt 3/3
WOL Plugin [wol-middleware]: All wake attempts failed, service remains offline
WOL Plugin [wol-middleware]: Returning 503 Service Unavailable
```

### Successful Power-Off Operation
```
WOL Plugin [wol-middleware]: Power-off requested via SSH
WOL Plugin [wol-middleware]: Executing SSH command: ssh -i /path/to/key user@192.168.1.100 "sudo shutdown -h now"
WOL Plugin [wol-middleware]: SSH connection established successfully
WOL Plugin [wol-middleware]: Power-off command executed successfully
WOL Plugin [wol-middleware]: Service shutting down...
```

### Configuration Errors
```
WOL Plugin [wol-middleware]: Invalid MAC address format: "invalid-mac"
WOL Plugin [wol-middleware]: Configuration error: healthCheck URL is required
WOL Plugin [wol-middleware]: SSH configuration incomplete: missing sshHost or sshUser
WOL Plugin [wol-middleware]: IPMI configuration incomplete: missing ipmiHost
```

## Testing Steps

### Step 1: Verify Plugin Loading
```bash
# Check Traefik logs for plugin loading
docker logs traefik | grep -i "traefik-power-management"

# Should see: Plugin traefik-power-management loaded successfully
```

### Step 2: Test Health Check
```bash
# Test the exact health endpoint from your configuration
curl -v http://192.168.1.100:3000/health

# Expected: HTTP 200 response
```

### Step 3: Test Target Device WOL Support
```bash
# Check if WOL is enabled (Linux target)
ssh user@target "sudo ethtool eth0 | grep Wake"
# Should show: Wake-on: g

# Manual WOL test
wakeonlan 00:11:22:33:44:55
```

### Step 4: Test Power-Off Method
```bash
# SSH method
ssh -i /path/to/key user@target "sudo shutdown -h now"

# IPMI method  
ipmitool -I lanplus -H target-bmc -U admin -P password chassis power status

# Custom method
/path/to/custom/command --test
```

### Step 5: Test Through Traefik
```bash
# Put target device to sleep/shutdown
# Access service through Traefik
curl http://your-traefik-service/

# Check Traefik logs for plugin activity
docker logs -f traefik | grep "WOL Plugin"
```

### Step 6: Test Dashboard (if enabled)
```bash
# Access control page directly
curl http://your-traefik-service/

# Test API endpoints
curl -X POST http://your-traefik-service/_wol/wake
curl http://your-traefik-service/_wol/status
```

If issues persist after following this guide, please create a [GitHub issue](https://github.com/ottup/traefik-power-management/issues) with:
1. Complete configuration (sanitized)
2. Relevant log output with debug enabled
3. Network topology description
4. Target device details and WOL configuration