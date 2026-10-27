package wireguard

import (
	"strings"
	"testing"
)

func TestNewConfigValidatesAndPreservesAmneziaWGProfile(t *testing.T) {
	raw := `{"interface_name":"awg-test","listen_port":51830,"address":["10.77.0.1/24"],"private_key":"REPLACE_PRIVATE_KEY","amneziawg":true,"jc":3,"jmin":20,"jmax":50,"s1":15,"s2":64,"s3":25,"s4":8,"h1":"100","h2":"200","h3":"300","h4":"400"}`
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	// privateKey is substituted into the JSON fixture below.
	raw = strings.Replace(raw, "REPLACE_PRIVATE_KEY", privateKey, 1)

	cfg, err := NewConfig(raw)
	if err != nil {
		t.Fatalf("expected valid AmneziaWG profile, got: %v", err)
	}
	if !cfg.AmneziaWG {
		t.Fatal("expected AmneziaWG mode to be enabled")
	}
	awgConfig, err := buildAmneziaDeviceConfig(cfg)
	if err != nil {
		t.Fatalf("build AmneziaWG device config: %v", err)
	}
	if awgConfig.Jc == nil || *awgConfig.Jc != 3 {
		t.Fatalf("expected Jc=3, got %#v", awgConfig.Jc)
	}
	if awgConfig.H4 == nil || *awgConfig.H4 != "400" {
		t.Fatalf("expected H4=400, got %#v", awgConfig.H4)
	}
}

func TestNewConfigRejectsAmneziaWGWithoutObfuscationParameters(t *testing.T) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	raw := `{"interface_name":"awg-test","listen_port":51830,"address":["10.77.0.1/24"],"private_key":"` + privateKey + `","amneziawg":true}`
	if _, err := NewConfig(raw); err == nil || !strings.Contains(err.Error(), "requires Jc") {
		t.Fatalf("expected missing AWG parameters to be rejected, got: %v", err)
	}
}

func TestNewConfigAllowsIndividuallyConfiguredAmneziaWGInitPackets(t *testing.T) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	raw := `{"interface_name":"awg-test","listen_port":51830,"address":["10.77.0.1/24"],"private_key":"` + privateKey + `","amneziawg":true,"jc":3,"jmin":20,"jmax":50,"s1":15,"s2":64,"s3":25,"s4":8,"h1":"100","h2":"200","h3":"300","h4":"400","i1":"<r 16>"}`
	cfg, err := NewConfig(raw)
	if err != nil {
		t.Fatalf("expected an individual I1 packet to be accepted, got: %v", err)
	}
	awgConfig, err := buildAmneziaDeviceConfig(cfg)
	if err != nil {
		t.Fatalf("build AmneziaWG device config: %v", err)
	}
	if awgConfig.I1 == nil || *awgConfig.I1 != "<r 16>" {
		t.Fatalf("expected I1 to be preserved, got %#v", awgConfig.I1)
	}
	if awgConfig.I2 != nil || awgConfig.I3 != nil || awgConfig.I4 != nil || awgConfig.I5 != nil {
		t.Fatal("unset I2-I5 parameters should remain unset")
	}
}
