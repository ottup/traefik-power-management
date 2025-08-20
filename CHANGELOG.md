# Changelog

All notable changes to the Traefik Power Management Plugin are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v3.0.0] - Complete Power Management Overhaul 🚀

**BREAKING CHANGES**: This is a complete rewrite focused on comprehensive power management. No backward compatibility with previous versions.

### 🔋 Power Management Revolution

#### Added
- **Bi-directional Power Control**: Full lifecycle management with both wake-up and shutdown capabilities
- **SSH Power Control**: Secure remote shutdown via SSH with key or password authentication
- **IPMI Integration**: Enterprise-grade power control via BMC/iDRAC interfaces
- **Custom Command Support**: Flexible power management with user-defined scripts and commands
- **Power State Tracking**: Real-time monitoring of both wake and shutdown operations

### 🎛️ Enhanced Interactive Dashboard

#### Added
- **Power-Off Button**: Secure shutdown control with optional confirmation dialogs
- **Conditional UI Elements**: Configurable button visibility and user interface customization
- **Auto-Redirect Control**: Optional automatic redirection with customizable delays
- **Security Controls**: Confirmation dialogs, credential protection, and secure command execution

#### Changed
- **Progress Tracking**: Removed confusing time remaining display, enhanced progress indicators
- **Mobile Optimization**: Improved responsive design for better mobile experience

### ⚙️ Comprehensive Configuration System

#### Added
- **25+ Configuration Fields**: Extensive customization options for every aspect of power management
- **Dashboard Customization**: `showPowerOffButton`, `confirmPowerOff`, `hideRedirectButton`, `autoRedirect`, `redirectDelay`
- **SSH Configuration**: `sshHost`, `sshUser`, `sshKeyPath`, `sshPassword`, `sshPort`
- **IPMI Configuration**: `ipmiHost`, `ipmiUser`, `ipmiPassword`
- **Power Management**: `powerOffMethod`, `powerOffCommand`

#### Improved
- **Enhanced Validation**: Comprehensive configuration validation with detailed error messages

### 🌐 Network & Performance Improvements  

#### Added
- **HTTP Connection Pooling**: Optimized health checks with keep-alive connections and proper headers
- **Race Condition Prevention**: Enhanced mutex handling for concurrent operations
- **Memory Optimization**: Reduced allocations and improved resource management

#### Improved
- **Error Recovery**: Better error handling with detailed feedback and retry mechanisms

### 🔗 API Enhancement

#### Added
- **New Endpoint**: `/_wol/poweroff` (POST) for secure power-off operations
- **Enhanced Status API**: Updated `/_wol/status` with power-off state tracking (`isPoweringOff`)
- **Improved Progress Tracking**: Real-time progress updates for both wake and shutdown operations

### 🔒 Security & Enterprise Features

#### Added
- **Authentication Methods**: Support for SSH keys, passwords, and IPMI credentials
- **Secure Command Execution**: Safe handling of remote commands with proper validation
- **Credential Protection**: Secure storage and handling of authentication credentials
- **Confirmation Controls**: User-configurable confirmation dialogs for destructive operations

### 📋 Breaking Changes

- **No Backward Compatibility**: Complete rewrite requires new configuration format
- **New Default Behavior**: Control page enabled by default with power management features
- **Configuration Changes**: All configuration field names and structure updated
- **Template Changes**: Complete UI overhaul removes time remaining and adds power controls
- **API Changes**: Status endpoint now includes `isPoweringOff` field

### Migration Guide

To upgrade from v2.x to v3.0.0:

1. **Update Configuration Format**: All configuration fields have been updated
2. **Enable Control Page**: Set `enableControlPage: true` to access new features
3. **Configure Power Management**: Add power-off method configuration
4. **Update API Usage**: Status endpoint response format has changed
5. **Test Thoroughly**: Complete rewrite requires full testing

---

## [v2.1.0] - Interactive Control Page

### Added
- **Interactive Control Page**: Optional web interface for manual WOL control instead of automatic wake-up
- **Real-time Status Updates**: Live progress indicators with detailed status messages during wake process
- **Responsive Design**: Mobile-friendly interface with modern CSS animations and touch-optimized controls
- **Manual Wake Control**: "Turn On Service" button triggers WOL with progress tracking
- **Direct Access Option**: "Go to Service Anyway" button bypasses health checks for immediate access
- **REST API Endpoints**: New `/_wol/wake`, `/_wol/status`, and `/_wol/redirect` endpoints for control page functionality
- **Background Processing**: Non-blocking WOL operations with real-time progress updates via JavaScript polling

### Changed
- **Default Behavior**: Plugin can now operate in two modes - automatic wake (default) or interactive control page
- **Configuration**: Added `enableControlPage` option to enable interactive dashboard

### Technical Details
- **API Endpoints**: Added REST endpoints for dashboard functionality
- **JavaScript Integration**: Client-side polling for real-time status updates
- **CSS Animations**: Smooth transitions and loading indicators
- **Mobile Support**: Touch-optimized buttons and responsive layout

---

## [v2.0.1] - Yaegi Compatibility Fix

### Fixed
- **Bugfix**: Fix Yaegi interpreter compatibility issue with UDP socket type assertion
- **Plugin Loading**: Resolve plugin loading failure that prevented middleware from being available
- **Compatibility**: Ensure plugin works properly in Traefik's Yaegi environment

### Technical Details
- **Type Assertion**: Fixed Go type assertion issue specific to Yaegi interpreter
- **UDP Socket Handling**: Corrected socket type handling for Traefik's plugin environment
- **Runtime Compatibility**: Ensured all Go features used are compatible with Yaegi

---

## [v2.0.0] - Container/LXC Optimization

### Added
- **Container/LXC Optimization**: Enhanced WOL packet delivery for containerized environments
- **Broadcast Support**: Automatic network interface discovery and broadcast address calculation
- **Smart Health Check Caching**: Configurable health check intervals to reduce redundant checks
- **Multi-Interface Support**: Send WOL packets on multiple network interfaces for maximum reliability

### Changed
- **Network Strategy**: Improved WOL packet delivery strategy for container environments
- **Health Check Strategy**: Added caching to reduce unnecessary health check requests

### Technical Details
- **Broadcast Discovery**: Automatic detection of available network interfaces and broadcast addresses
- **Container Compatibility**: Enhanced support for Docker, LXC, and other containerized deployments
- **Connection Pooling**: Improved HTTP client with connection reuse
- **Cache Implementation**: Smart caching for health check results with configurable intervals

### Configuration Changes
- **New Options**: Added `healthCheckInterval` for configurable health check caching
- **Network Options**: Enhanced network configuration options for complex setups

---

## [v1.0.0] - Initial Release

### Added
- **Core WOL Functionality**: Basic Wake-on-LAN packet transmission
- **Health Check Monitoring**: HTTP endpoint monitoring to detect service availability
- **Magic Packet Generation**: IEEE 802.3 compliant Wake-on-LAN magic packet creation
- **MAC Address Support**: Flexible MAC address format parsing (colon, dash, dot, no separator)
- **Configurable Timeouts**: Customizable wake timeout and retry attempts
- **Error Handling**: Comprehensive error handling and logging
- **Yaegi Compatibility**: Plugin designed for Traefik's Yaegi interpreter

### Configuration Options
- **Required**: `healthCheck`, `macAddress`
- **Optional**: `ipAddress`, `port`, `timeout`, `retryAttempts`, `retryInterval`, `debug`

### Technical Implementation
- **Single File Plugin**: Complete implementation in `main.go` for Yaegi compatibility
- **Go Standard Library**: Uses only Go standard library for maximum compatibility
- **Network Protocol**: UDP-based magic packet transmission on configurable port (default: 9)
- **Retry Logic**: Configurable retry attempts with exponential backoff

---

## Version History Summary

| Version | Release Date | Major Features |
|---------|-------------|----------------|
| v3.0.0 | 2024-XX-XX | Complete power management with shutdown capabilities, interactive dashboard, SSH/IPMI support |
| v2.1.0 | 2024-XX-XX | Interactive control page, real-time status updates, manual wake control |
| v2.0.1 | 2024-XX-XX | Yaegi compatibility fix for UDP socket handling |
| v2.0.0 | 2024-XX-XX | Container optimization, broadcast support, smart health check caching |
| v1.0.0 | 2024-XX-XX | Initial release with core WOL functionality |

---

## Upgrade Notes

### From v2.x to v3.0.0
- **Complete configuration rewrite required**
- **New power management features available**
- **Enhanced security and authentication options**
- **API endpoint changes**

### From v1.x to v2.x
- **Backward compatible configuration**
- **New container optimization features**
- **Optional interactive control page**
- **Enhanced network support**

---

## Support and Compatibility

### Traefik Compatibility
- **Minimum Version**: Traefik v2.3+ (plugin support required)
- **Recommended**: Traefik v2.8+ for best performance
- **Plugin System**: Yaegi-based interpreted plugin

### Go Version
- **Development**: Go 1.21+
- **Runtime**: Yaegi interpreter (no Go installation required on Traefik host)

### Operating System Support
- **Linux**: Full support (primary development platform)
- **macOS**: Full support
- **Windows**: Full support
- **Container Platforms**: Docker, LXC, Kubernetes, Docker Compose

---

For detailed upgrade instructions and migration guides, see the [main README](README.md) or create an issue if you need assistance upgrading.