package wireguard

import (
	"net"
	"testing"
	"time"

	"github.com/awg-go/awgctrl-go/wgtypes"
)

func TestBuildAddConfigUsesProductionPersistentKeepalive(t *testing.T) {
	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}

	_, peerIP, err := net.ParseCIDR("10.0.0.2/32")
	if err != nil {
		t.Fatalf("ParseCIDR() error = %v", err)
	}

	cfg := buildAddConfig(key, []net.IPNet{*peerIP}, nil)
	if cfg.PersistentKeepaliveInterval == nil {
		t.Fatal("PersistentKeepaliveInterval = nil, want 25s")
	}
	if *cfg.PersistentKeepaliveInterval != 25*time.Second {
		t.Fatalf("PersistentKeepaliveInterval = %v, want 25s", *cfg.PersistentKeepaliveInterval)
	}
}

func TestBuildAddConfigPreservesNonZeroPersistentKeepalive(t *testing.T) {
	old := tempKeepAlive
	defer func() { tempKeepAlive = old }()

	tempKeepAlive = 25 * time.Second

	cfg := buildAddConfig(wgtypes.Key{}, nil, nil)
	if cfg.PersistentKeepaliveInterval == nil {
		t.Fatal("PersistentKeepaliveInterval = nil, want 25s")
	}
	if *cfg.PersistentKeepaliveInterval != 25*time.Second {
		t.Fatalf("PersistentKeepaliveInterval = %v, want 25s", *cfg.PersistentKeepaliveInterval)
	}
}

func TestAmneziaWGPeerConfigDoesNotSendZeroPersistentKeepalive(t *testing.T) {
	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}

	_, peerIP, err := net.ParseCIDR("10.0.0.2/32")
	if err != nil {
		t.Fatalf("ParseCIDR() error = %v", err)
	}

	wg := &WireGuard{config: &Config{AmneziaWG: true}}
	cfg, err := wg.buildAddConfigFromPeerInfo(&PeerInfo{
		Email:      "awg@example.com",
		PublicKey:  key,
		AllowedIPs: []net.IPNet{*peerIP},
	}, nil)
	if err != nil {
		t.Fatalf("buildAddConfigFromPeerInfo() error = %v", err)
	}

	if !cfg.AdvancedSecurity {
		t.Fatal("expected AdvancedSecurity=true for AmneziaWG")
	}
	if cfg.PersistentKeepaliveInterval != nil {
		t.Fatalf("PersistentKeepaliveInterval = %v, want nil for AmneziaWG zero keepalive", *cfg.PersistentKeepaliveInterval)
	}
}
