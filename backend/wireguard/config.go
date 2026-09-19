package wireguard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/advanced-wg/awgctrl-go/wgtypes"
)

// Config represents the WireGuard configuration
type Config struct {
	InterfaceName string         `json:"interface_name"`
	PrivateKey    string         `json:"private_key"`
	PreSharedKey  string         `json:"pre_shared_key,omitempty"`
	ListenPort    int            `json:"listen_port"`
	Address       []string       `json:"address"`
	Latency       *LatencyConfig `json:"latency,omitempty"`
	AmneziaWG      bool    `json:"amneziawg,omitempty"`
	Jc         *int    `json:"jc,omitempty"`
	Jmin       *int    `json:"jmin,omitempty"`
	Jmax       *int    `json:"jmax,omitempty"`
	S1         *int    `json:"s1,omitempty"`
	S2         *int    `json:"s2,omitempty"`
	S3         *int    `json:"s3,omitempty"`
	S4         *int    `json:"s4,omitempty"`
	H1         *string `json:"h1,omitempty"`
	H2         *string `json:"h2,omitempty"`
	H3         *string `json:"h3,omitempty"`
	H4         *string `json:"h4,omitempty"`
	I1         *string `json:"i1,omitempty"`
	I2         *string `json:"i2,omitempty"`
	I3         *string `json:"i3,omitempty"`
	I4         *string `json:"i4,omitempty"`
	I5         *string `json:"i5,omitempty"`

	privateKeyValue   wgtypes.Key
	privateKeySet     bool
	presharedKeyValue *wgtypes.Key
	presharedKeySet   bool

	mu sync.RWMutex
}

type LatencyConfig struct {
	TestURL        string `json:"test_url,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// PeerInfo stores information about a WireGuard peer
type PeerInfo struct {
	Email      string      `json:"email"`
	PublicKey  wgtypes.Key `json:"public_key"`
	AllowedIPs []net.IPNet `json:"allowed_ips"`
}

func clonePeerInfo(peer *PeerInfo) *PeerInfo {
	if peer == nil {
		return nil
	}

	return &PeerInfo{
		Email:      peer.Email,
		PublicKey:  peer.PublicKey,
		AllowedIPs: append([]net.IPNet(nil), peer.AllowedIPs...),
	}
}

// NewConfig creates a new WireGuard configuration from JSON
func NewConfig(config string) (*Config, error) {
	var wgConfig Config
	err := json.Unmarshal([]byte(config), &wgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if wgConfig.InterfaceName == "" {
		wgConfig.InterfaceName = "wg0"
	}

	if wgConfig.ListenPort <= 0 {
		wgConfig.ListenPort = 51820
	}
	if wgConfig.Latency == nil {
		wgConfig.Latency = &LatencyConfig{}
	}
	if strings.TrimSpace(wgConfig.Latency.TestURL) == "" {
		wgConfig.Latency.TestURL = "https://www.gstatic.com/generate_204"
	}
	if wgConfig.Latency.TimeoutSeconds <= 0 {
		wgConfig.Latency.TimeoutSeconds = 5
	}
	if err := wgConfig.ValidateAmnezia(); err != nil {
		return nil, fmt.Errorf("invalid AmneziaWG configuration: %w", err)
	}

	return &wgConfig, nil
}

func (c *Config) AmneziaConfig() wgtypes.Config {
	return wgtypes.Config{
		Jc: c.Jc, Jmin: c.Jmin, Jmax: c.Jmax,
		S1: c.S1, S2: c.S2, S3: c.S3, S4: c.S4,
		H1: c.H1, H2: c.H2, H3: c.H3, H4: c.H4,
		I1: c.I1, I2: c.I2, I3: c.I3, I4: c.I4, I5: c.I5,
	}
}

func (c *Config) ValidateAmnezia() error {
	if !c.AmneziaWG {
		return nil
	}
	cfg := c.AmneziaConfig()\n\treturn cfg.Validate()
}

// InterfaceNetworks returns CIDR prefixes parsed from the node's core `address` list.
// Used to restrict peer AllowedIPs to subnets this interface actually serves.
func (c *Config) InterfaceNetworks() []*net.IPNet {
	if c == nil {
		return nil
	}
	var out []*net.IPNet
	for _, addr := range c.Address {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(addr)
		if err != nil {
			continue
		}
		out = append(out, ipNet)
	}
	return out
}

// GetPrivateKey returns the parsed WireGuard private key.
// The parsed key is stored in memory and reused after first successful parse.
func (c *Config) GetPrivateKey() (wgtypes.Key, error) {
	c.mu.RLock()
	if c.privateKeySet {
		key := c.privateKeyValue
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.privateKeySet {
		return c.privateKeyValue, nil
	}

	if c.PrivateKey == "" {
		// Return Error
		return wgtypes.Key{}, errors.New("private key is empty")
	}

	key, err := wgtypes.ParseKey(c.PrivateKey)
	if err != nil {
		return wgtypes.Key{}, err
	}
	c.privateKeyValue = key
	c.privateKeySet = true
	return key, nil
}

// GetPreSharedKey parses and caches the pre-shared key, optionally returning nil if not set.
func (c *Config) GetPreSharedKey() (*wgtypes.Key, error) {
	c.mu.RLock()
	if c.presharedKeySet {
		c.mu.RUnlock()
		return c.presharedKeyValue, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.presharedKeySet {
		return c.presharedKeyValue, nil
	}

	if c.PreSharedKey == "" {
		c.presharedKeySet = true
		c.presharedKeyValue = nil
		return nil, nil
	}

	key, err := wgtypes.ParseKey(c.PreSharedKey)
	if err != nil {
		return nil, err
	}
	c.presharedKeyValue = &key
	c.presharedKeySet = true
	return &key, nil
}

// GenerateKeyPair generates a new WireGuard key pair
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}

	return privKey.String(), privKey.PublicKey().String(), nil
}
