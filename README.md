# Custom Traefik WOL Plugin

Enhanced Wake-on-LAN plugin for Traefik with additional features:

- Configurable retry attempts
- Configurable retry intervals
- Debug logging
- Enhanced error handling
- Flexible MAC address formats

## Configuration

```yaml
middlewares:
  my-wol:
    plugin:
      custom-wol:
        healthCheck: http://192.168.1.100:3000/health
        macAddress: "00:11:22:33:44:55"
        ipAddress: "192.168.1.100"
        port: "9"           # Optional, default: 9
        timeout: "30"       # Optional, default: 30 seconds
        retryAttempts: "3"  # Optional, default: 3
        retryInterval: "5"  # Optional, default: 5 seconds
        debug: true         # Optional, default: false
```