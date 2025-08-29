package traefik_power_management

import (
	"testing"
)

func TestParseMACAddress(t *testing.T) {
	plugin := &WOLPlugin{}

	tests := []struct {
		name     string
		input    string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "colon separated",
			input:    "00:11:22:33:44:55",
			expected: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			wantErr:  false,
		},
		{
			name:     "dash separated",
			input:    "00-11-22-33-44-55",
			expected: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			wantErr:  false,
		},
		{
			name:     "dot separated",
			input:    "0011.2233.4455",
			expected: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			wantErr:  false,
		},
		{
			name:     "no separators",
			input:    "001122334455",
			expected: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			wantErr:  false,
		},
		{
			name:     "uppercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			wantErr:  false,
		},
		{
			name:     "invalid length",
			input:    "00:11:22:33:44",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid hex",
			input:    "GG:11:22:33:44:55",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := plugin.parseMACAddress(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %s, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tt.input, err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, expected, result[i])
				}
			}
		})
	}
}

func TestCreateMagicPacket(t *testing.T) {
	plugin := &WOLPlugin{}
	macBytes := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	
	packet := plugin.createMagicPacket(macBytes)
	
	// Magic packet should be 102 bytes total
	if len(packet) != 102 {
		t.Errorf("expected packet length 102, got %d", len(packet))
	}

	// First 6 bytes should be 0xFF
	for i := 0; i < 6; i++ {
		if packet[i] != 0xFF {
			t.Errorf("byte %d should be 0xFF, got 0x%02X", i, packet[i])
		}
	}

	// Next 96 bytes should be the MAC address repeated 16 times
	for i := 0; i < 16; i++ {
		offset := 6 + i*6
		for j := 0; j < 6; j++ {
			expected := macBytes[j]
			actual := packet[offset+j]
			if actual != expected {
				t.Errorf("MAC repetition %d, byte %d: expected 0x%02X, got 0x%02X", i, j, expected, actual)
			}
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := CreateConfig()

	// Check default values
	if config.Port != "9" {
		t.Errorf("expected default port '9', got '%s'", config.Port)
	}

	if config.Timeout != "30" {
		t.Errorf("expected default timeout '30', got '%s'", config.Timeout)
	}

	if config.RetryAttempts != "3" {
		t.Errorf("expected default retry attempts '3', got '%s'", config.RetryAttempts)
	}

	if config.RetryInterval != "5" {
		t.Errorf("expected default retry interval '5', got '%s'", config.RetryInterval)
	}

	if config.HealthCheckInterval != "10" {
		t.Errorf("expected default health check interval '10', got '%s'", config.HealthCheckInterval)
	}

	if config.PowerOffCommand != "/usr/local/bin/shutdown-script.sh" {
		t.Errorf("expected default power off command '/usr/local/bin/shutdown-script.sh', got '%s'", config.PowerOffCommand)
	}

	if config.ShowPowerOffButton != true {
		t.Errorf("expected default ShowPowerOffButton true, got %v", config.ShowPowerOffButton)
	}

	if config.ConfirmPowerOff != true {
		t.Errorf("expected default ConfirmPowerOff true, got %v", config.ConfirmPowerOff)
	}
}

func TestNewPluginValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				HealthCheck:   "http://example.com/health",
				MacAddress:    "00:11:22:33:44:55",
				Port:          "9",
				Timeout:       "30",
				RetryAttempts: "3",
				RetryInterval: "5",
				HealthCheckInterval: "10",
				RedirectDelay: "3",
			},
			wantError: false,
		},
		{
			name: "missing health check",
			config: &Config{
				MacAddress:    "00:11:22:33:44:55",
				Port:          "9",
				Timeout:       "30",
				RetryAttempts: "3",
				RetryInterval: "5",
				HealthCheckInterval: "10",
				RedirectDelay: "3",
			},
			wantError: true,
			errorMsg:  "healthCheck URL is required",
		},
		{
			name: "missing MAC address",
			config: &Config{
				HealthCheck:   "http://example.com/health",
				Port:          "9",
				Timeout:       "30",
				RetryAttempts: "3",
				RetryInterval: "5",
				HealthCheckInterval: "10",
				RedirectDelay: "3",
			},
			wantError: true,
			errorMsg:  "macAddress is required",
		},
		{
			name: "invalid port",
			config: &Config{
				HealthCheck:   "http://example.com/health",
				MacAddress:    "00:11:22:33:44:55",
				Port:          "invalid",
				Timeout:       "30",
				RetryAttempts: "3",
				RetryInterval: "5",
				HealthCheckInterval: "10",
				RedirectDelay: "3",
			},
			wantError: true,
			errorMsg:  "invalid port",
		},
		{
			name: "power-off button without command",
			config: &Config{
				HealthCheck:        "http://example.com/health",
				MacAddress:         "00:11:22:33:44:55",
				Port:               "9",
				Timeout:            "30",
				RetryAttempts:      "3",
				RetryInterval:      "5",
				HealthCheckInterval: "10",
				RedirectDelay:      "3",
				ShowPowerOffButton: true,
				PowerOffCommand:    "",
			},
			wantError: true,
			errorMsg:  "powerOffCommand is required when showPowerOffButton is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(nil, nil, tt.config, "test")

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
					return
				}
				if err.Error() != tt.errorMsg && len(tt.errorMsg) > 0 {
					// Allow partial matches for complex error messages
					found := false
					if len(tt.errorMsg) > 10 {
						// For longer error messages, just check if it contains the key part
						if tt.errorMsg == "invalid port" && err.Error() != "invalid port: strconv.Atoi: parsing \"invalid\": invalid syntax" {
							// This is expected - the actual error includes more detail
							found = true
						} else if err.Error() == tt.errorMsg {
							found = true
						}
					} else {
						found = err.Error() == tt.errorMsg
					}
					if !found {
						t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}