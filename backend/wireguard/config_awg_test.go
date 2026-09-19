package wireguard

import "testing"

func TestNewConfigAmneziaWGValidation(t *testing.T) {
	cfg := `{
		"interface_name":"awg0",
		"private_key":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"listen_port":51820,
		"address":["10.0.0.1/24"],
		"amneziawg":true,
		"jc":3,
		"jmin":64,
		"jmax":128,
		"s1":16,
		"s2":17,
		"s3":18,
		"s4":4,
		"h1":"123456-123999",
		"h2":"223456-223999",
		"i1":"<r 16>"
	}`
	got, err := NewConfig(cfg)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if !got.AmneziaWG {
		t.Fatal("AmneziaWG flag was not preserved")
	}
	if got.ListenPort != 51820 {
		t.Fatalf("ListenPort = %d, want 51820", got.ListenPort)
	}
	if got.Jc == nil || *got.Jc != 3 {
		t.Fatalf("Jc = %v, want 3", got.Jc)
	}
	if got.I1 == nil || *got.I1 != "<r 16>" {
		t.Fatalf("I1 = %v, want <r 16>", got.I1)
	}
}

func TestNewConfigRejectsInvalidAmneziaWG(t *testing.T) {
	cfg := `{
		"interface_name":"awg0",
		"private_key":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"listen_port":51820,
		"address":["10.0.0.1/24"],
		"amneziawg":true,
		"jc":11
	}`
	if _, err := NewConfig(cfg); err == nil {
		t.Fatal("NewConfig() accepted invalid Jc=11")
	}
}
