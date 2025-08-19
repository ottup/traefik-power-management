# Traefik Wake-on-LAN Plugin

A robust Wake-on-LAN middleware plugin for Traefik that automatically wakes up sleeping services when they become unavailable.

## Features

- **Health Check Monitoring**: Smart health check caching with configurable intervals to reduce redundant checks
- **Automatic Wake-up**: Sends WOL magic packets when services are down
- **Container-Optimized**: Enhanced broadcast packet delivery for containers, LXC, and Docker environments
- **Multi-Interface Support**: Automatic network interface discovery and broadcast address calculation
- **Configurable Retry Logic**: Customizable retry attempts and intervals
- **Flexible MAC Address Support**: Accepts various MAC address formats (colon, dash, dot separated)
- **Smart Logging**: State-change-only logging to reduce log spam when services are healthy
- **Debug Logging**: Detailed logging for troubleshooting
- **Enhanced Error Handling**: Robust error handling with informative messages
- **Production Ready**: Designed for reliable operation in production environments

## Requirements

- **Traefik**: v2.3+ (plugin support required)
- **Network**: Target devices must support Wake-on-LAN
- **Permissions**: UDP socket access on port 9 (or configured port)

## Installation

### Method 1: Official Plugin Catalog (Recommended)

**Note**: This method will be available once the plugin is approved by Traefik.

Add the plugin to your Traefik static configuration:

#### YAML Configuration
```yaml
experimental:
  plugins:
    traefik-wol:
      moduleName: "github.com/ottup/traefik-wol"
      version: "v1.0.1"
```

#### TOML Configuration
```toml
[experimental.plugins.traefik-wol]
  moduleName = "github.com/ottup/traefik-wol"
  version = "v1.0.1"
```

#### CLI Flags
```bash
--experimental.plugins.traefik-wol.moduleName=github.com/ottup/traefik-wol
--experimental.plugins.traefik-wol.version=v1.0.1
```

### Method 2: Local Development Plugin

For development or private deployment:

1. **Create Plugin Directory Structure**
   ```bash
   mkdir -p ./plugins-local/src/github.com/ottup/traefik-wol
   ```

2. **Download Plugin Source**

   Choose one of the following methods:

   #### Option A: Git Clone (Recommended)
   ```bash
   git clone https://github.com/ottup/traefik-wol.git ./plugins-local/src/github.com/ottup/traefik-wol
   ```

   #### Option B: Direct File Download (curl)
   ```bash
   cd ./plugins-local/src/github.com/ottup/traefik-wol
   curl -L https://raw.githubusercontent.com/ottup/traefik-wol/v1.0.3/main.go -o main.go
   curl -L https://raw.githubusercontent.com/ottup/traefik-wol/v1.0.3/go.mod -o go.mod
   curl -L https://raw.githubusercontent.com/ottup/traefik-wol/v1.0.3/.traefik.yml -o .traefik.yml
   ```

   #### Option C: Direct File Download (wget)
   ```bash
   cd ./plugins-local/src/github.com/ottup/traefik-wol
   wget https://raw.githubusercontent.com/ottup/traefik-wol/v1.0.3/main.go
   wget https://raw.githubusercontent.com/ottup/traefik-wol/v1.0.3/go.mod
   wget https://raw.githubusercontent.com/ottup/traefik-wol/v1.0.3/.traefik.yml
   ```

   #### Option D: GitHub Archive Download (zip)
   ```bash
   curl -L https://github.com/ottup/traefik-wol/archive/refs/tags/v1.0.3.zip -o traefik-wol.zip
   unzip traefik-wol.zip
   mv traefik-wol-1.0.3/* ./plugins-local/src/github.com/ottup/traefik-wol/
   rm -rf traefik-wol.zip traefik-wol-1.0.3
   ```

   #### Option E: GitHub Archive Download (tarball)
   ```bash
   curl -L https://github.com/ottup/traefik-wol/archive/refs/tags/v1.0.3.tar.gz | tar -xz
   mv traefik-wol-1.0.3/* ./plugins-local/src/github.com/ottup/traefik-wol/
   rm -rf traefik-wol-1.0.3
   ```

   #### Option F: Manual Download
   1. Visit: https://github.com/ottup/traefik-wol/releases/tag/v1.0.3
   2. Download `Source code (zip)` or `Source code (tar.gz)`
   3. Extract the archive
   4. Copy `main.go`, `go.mod`, and `.traefik.yml` to `./plugins-local/src/github.com/ottup/traefik-wol/`

   **Quick Reference - Choose Your Method:**
   - **Git available**: Use Option A (git clone)
   - **No git, have curl**: Use Option B (direct download)
   - **No git, have wget**: Use Option C (direct download)
   - **Want single archive**: Use Option D (zip) or E (tarball)
   - **Browser only**: Use Option F (manual download)

   **Required Files**: All methods must include `main.go`, `go.mod`, and `.traefik.yml`

3. **Configure Static Configuration**

   #### YAML Configuration
   ```yaml
   experimental:
     localPlugins:
       traefik-wol:
         moduleName: "github.com/ottup/traefik-wol"
   ```

   #### TOML Configuration
   ```toml
   [experimental.localPlugins.traefik-wol]
     moduleName = "github.com/ottup/traefik-wol"
   ```

4. **Restart Traefik**
   
   Plugins are loaded at startup, so restart is required after configuration changes.

## Configuration

### Basic Configuration

```yaml
middlewares:
  wol-middleware:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"
        macAddress: "00:11:22:33:44:55"
        ipAddress: "192.168.1.100"
```

### Complete Configuration with All Options

```yaml
middlewares:
  wol-middleware:
    plugin:
      traefik-wol:
        healthCheck: "http://192.168.1.100:3000/health"  # Required: Health check endpoint
        macAddress: "00:11:22:33:44:55"                   # Required: Target MAC address
        ipAddress: "192.168.1.100"                        # Optional: Target IP address (used for unicast, fallback to broadcast)
        broadcastAddress: "192.168.1.255"                 # Optional: Custom broadcast address
        networkInterface: "eth0"                          # Optional: Specific network interface to use
        port: "9"                                         # Optional: WOL port (default: 9)
        timeout: "30"                                     # Optional: Wake timeout in seconds (default: 30)
        retryAttempts: "3"                                # Optional: Number of retry attempts (default: 3)
        retryInterval: "5"                                # Optional: Delay between retries in seconds (default: 5)
        healthCheckInterval: "10"                         # Optional: Health check cache interval in seconds (default: 10)
        debug: true                                       # Optional: Enable debug logging (default: false)
```

### Alternative Configuration Formats

#### TOML Configuration
```toml
[http.middlewares.wol-middleware.plugin.traefik-wol]
  healthCheck = "http://192.168.1.100:3000/health"
  macAddress = "00:11:22:33:44:55"
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

#### JSON Configuration
```json
{
  "http": {
    "middlewares": {
      "wol-middleware": {
        "plugin": {
          "traefik-wol": {
            "healthCheck": "http://192.168.1.100:3000/health",
            "macAddress": "00:11:22:33:44:55",
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

### MAC Address Formats

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

### Container and Network Configuration

The plugin is optimized for containerized environments (Docker, LXC, etc.) and includes enhanced networking features:

#### Broadcast Packet Support
- **Automatic Broadcast Discovery**: The plugin automatically detects available network interfaces and calculates broadcast addresses
- **Container Compatibility**: Uses broadcast packets that can traverse container network boundaries
- **Multi-Interface Support**: Sends WOL packets on all available network interfaces for maximum reliability

#### Configuration Options

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

#### Deployment Scenarios

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

## Usage

### Step 1: Apply Middleware to Route

Once the plugin is installed and configured, apply it to your routes:

```yaml
http:
  routers:
    my-service:
      rule: "Host(`myservice.example.com`)"
      service: my-service
      middlewares:
        - wol-middleware

  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://192.168.1.100:3000"

  middlewares:
    wol-middleware:
      plugin:
        traefik-wol:
          healthCheck: "http://192.168.1.100:3000/health"
          macAddress: "00:11:22:33:44:55"
          ipAddress: "192.168.1.100"
          debug: true
```

### Step 2: Test the Plugin

1. **Ensure Target Device Supports WOL**:
   ```bash
   # Check if WOL is enabled (Linux example)
   sudo ethtool eth0 | grep Wake
   ```

2. **Test Health Check Endpoint**:
   ```bash
   curl http://192.168.1.100:3000/health
   ```

3. **Monitor Traefik Logs**:
   ```bash
   # Enable debug mode in plugin configuration to see detailed logs
   docker logs -f traefik
   ```

### Step 3: Verify Wake-on-LAN Functionality

1. **Put Target Device to Sleep**
2. **Access Service Through Traefik**
3. **Check Logs for Wake Attempts**:
   ```
   WOL Plugin [wol-middleware]: Service unhealthy, attempting to wake 00:11:22:33:44:55
   WOL Plugin [wol-middleware]: Magic packet sent to 00:11:22:33:44:55 (192.168.1.100:9)
   WOL Plugin [wol-middleware]: Service is now online
   ```

## Troubleshooting

### Common Issues

#### Plugin Not Loading
- **Cause**: Incorrect module name or version
- **Solution**: Verify `moduleName` matches exactly: `github.com/ottup/traefik-wol`
- **Check**: Restart Traefik after configuration changes

#### Health Check Failures
- **Cause**: Incorrect health check URL or unreachable service
- **Solution**: Test health endpoint manually:
  ```bash
  curl -v http://192.168.1.100:3000/health
  ```
- **Check**: Ensure endpoint returns HTTP 2xx status codes

#### Wake-on-LAN Not Working
- **Cause**: WOL not enabled on target device
- **Solution**: Enable WOL in BIOS/UEFI and network adapter settings
- **Check**: Test WOL manually:
  ```bash
  # Linux/Mac
  wakeonlan 00:11:22:33:44:55
  
  # Windows
  wolcmd 00:11:22:33:44:55 192.168.1.100
  ```

#### Container/LXC WOL Issues
- **Cause**: Network isolation preventing WOL packets from reaching target
- **Solution**: Use broadcast addresses instead of unicast:
  ```yaml
  # Remove ipAddress to force broadcast mode
  middlewares:
    wol-middleware:
      plugin:
        traefik-wol:
          healthCheck: "http://192.168.1.100:3000/health"
          macAddress: "00:11:22:33:44:55"
          # ipAddress: "192.168.1.100"  # Comment out or remove
          broadcastAddress: "192.168.1.255"  # Add specific broadcast
  ```
- **Check**: Enable debug mode to see packet delivery attempts

#### Network Issues
- **Cause**: Firewall blocking UDP traffic or incorrect network configuration
- **Solution**: Ensure UDP port 9 is accessible and broadcast packets are allowed
- **Container Networks**: Verify container can send broadcast packets:
  ```bash
  # Test from container
  docker exec traefik ping -c 1 192.168.1.255
  ```
- **Check**: Test with different broadcast addresses if needed

### Debug Mode

Enable debug logging for detailed troubleshooting:

```yaml
middlewares:
  wol-middleware:
    plugin:
      traefik-wol:
        debug: true  # Enable verbose logging
```

Debug logs include:
- Health check attempts and results
- Wake-on-LAN packet transmission details
- Service wake-up monitoring progress
- Error details and retry attempts

### Log Examples

**Successful Wake Operation**:
```
WOL Plugin [wol-middleware]: Checking health for http://192.168.1.100:3000/health
WOL Plugin [wol-middleware]: Health check failed: connection refused
WOL Plugin [wol-middleware]: Service unhealthy, attempting to wake 00:11:22:33:44:55
WOL Plugin [wol-middleware]: Wake attempt 1/3
WOL Plugin [wol-middleware]: Magic packet sent to 00:11:22:33:44:55 (192.168.1.100:9)
WOL Plugin [wol-middleware]: Waiting for service to come online (timeout: 30s)
WOL Plugin [wol-middleware]: Health check status: 200 (healthy: true)
WOL Plugin [wol-middleware]: Service is now online
```

## Best Practices

### Security Considerations

- **Network Segmentation**: Deploy WOL functionality within trusted network segments
- **Access Control**: Limit access to services that use WOL middleware
- **Monitoring**: Enable logging to track wake events and potential security issues
- **Health Check Security**: Ensure health check endpoints don't expose sensitive information

### Production Deployment

- **Testing**: Thoroughly test in development environment before production deployment
- **Monitoring**: Set up monitoring for failed wake attempts and service availability
- **Backup Plans**: Have alternative access methods if WOL fails
- **Network Dependencies**: Consider network infrastructure requirements and limitations

### Performance Optimization

- **Health Check Frequency**: Balance between responsiveness and system load
- **Timeout Configuration**: Set appropriate timeouts based on device wake-up times
- **Retry Strategy**: Configure retry attempts based on network reliability

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the Repository**
2. **Create Feature Branch**: `git checkout -b feature/your-feature`
3. **Make Changes**: Follow Go coding standards and existing patterns
4. **Add Tests**: Include tests for new functionality
5. **Update Documentation**: Update README and inline comments
6. **Submit Pull Request**: Include detailed description of changes

### Development Setup

1. **Prerequisites**: Go 1.21+ installed
2. **Clone Repository**: `git clone https://github.com/ottup/traefik-wol.git`
3. **Install Dependencies**: `go mod tidy`
4. **Run Tests**: `go test ./...`
5. **Build Plugin**: `go build .`

### Code Standards

- Follow standard Go formatting (`go fmt`)
- Include comprehensive error handling
- Add meaningful comments for complex logic
- Maintain backward compatibility when possible

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: Report bugs and feature requests via [GitHub Issues](https://github.com/ottup/traefik-wol/issues)
- **Documentation**: Additional documentation available in the repository
- **Community**: Join Traefik community discussions for general plugin support

## Changelog

### v2.0.0
- **Container/LXC Optimization**: Enhanced WOL packet delivery for containerized environments
- **Broadcast Support**: Automatic network interface discovery and broadcast address calculation
- **Smart Health Check Caching**: Configurable health check intervals to reduce redundant checks
- **State-Change Logging**: Reduced log spam by only logging on health state changes
- **Multi-Interface Support**: Send WOL packets on multiple network interfaces for maximum reliability
- **New Configuration Options**:
  - `broadcastAddress`: Custom broadcast address configuration
  - `networkInterface`: Specific network interface selection
  - `healthCheckInterval`: Configurable health check cache interval
- **Breaking Changes**: 
  - `ipAddress` is now optional (will fallback to broadcast if not provided)
  - Health check behavior changed to use caching (may affect very rapid health state changes)

### v1.0.1
- Fixed package naming issue for proper plugin loading
- Enhanced error handling and logging
- Improved MAC address format validation

### v1.0.0
- Initial release
- Core WOL functionality with health checking
- Configurable retry logic and timeouts
- Support for multiple MAC address formats

## Acknowledgments

- **Traefik Team**: For the excellent reverse proxy and plugin system
- **Go Community**: For comprehensive standard library and ecosystem
- **Contributors**: Thanks to all contributors who help improve this plugin