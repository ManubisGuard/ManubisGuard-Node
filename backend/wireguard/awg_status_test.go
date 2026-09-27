package wireguard

import (
	"strings"
	"testing"
	"time"
)

func TestParseAWGDeviceDump(t *testing.T) {
	const key = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	const peerKey = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="
	const dump = key + "	" + key + "	51820	3	20	50	15	64	25	8	1-2	3-4	5-6	7-8	(null)	(null)	(null)	(null)	(null)	(none)	0	0	0	off
" +
		peerKey + "	(none)	192.0.2.1:54321	10.0.0.2/32,10.0.0.3/32	1700000000	1234	5678	25
"

	device, err := parseAWGDeviceDump("awg0", dump)
	if err != nil {
		t.Fatalf("parseAWGDeviceDump() error = %v", err)
	}
	if device.Name != "awg0" {
		t.Fatalf("device name = %q, want awg0", device.Name)
	}
	if device.ListenPort != 51820 {
		t.Fatalf("listen port = %d, want 51820", device.ListenPort)
	}
	if device.PublicKey.String() != key {
		t.Fatalf("device public key = %q, want %q", device.PublicKey.String(), key)
	}
	if device.PrivateKey.String() != key {
		t.Fatalf("device private key = %q, want %q", device.PrivateKey.String(), key)
	}
	if len(device.Peers) != 1 {
		t.Fatalf("peer count = %d, want 1", len(device.Peers))
	}

	peer := device.Peers[0]
	if peer.PublicKey.String() != peerKey {
		t.Fatalf("peer public key = %q, want %q", peer.PublicKey.String(), peerKey)
	}
	if peer.Endpoint == nil || peer.Endpoint.String() != "192.0.2.1:54321" {
		t.Fatalf("unexpected endpoint: %v", peer.Endpoint)
	}
	if len(peer.AllowedIPs) != 2 {
		t.Fatalf("allowed IP count = %d, want 2", len(peer.AllowedIPs))
	}
	if peer.LastHandshakeTime != time.Unix(1700000000, 0) {
		t.Fatalf("unexpected handshake time: %v", peer.LastHandshakeTime)
	}
	if peer.ReceiveBytes != 1234 || peer.TransmitBytes != 5678 {
		t.Fatalf("unexpected transfer counters: rx=%d tx=%d", peer.ReceiveBytes, peer.TransmitBytes)
	}
	if peer.PersistentKeepaliveInterval != 25*time.Second {
		t.Fatalf("unexpected keepalive: %v", peer.PersistentKeepaliveInterval)
	}
}

func TestReadAWGDeviceUsesDumpRunner(t *testing.T) {
	old := awgOutputRunner
	defer func() { awgOutputRunner = old }()

	called := false
	awgOutputRunner = func(args ...string) ([]byte, error) {
		called = true
		if strings.Join(args, " ") != "show awg0 dump" {
			t.Fatalf("unexpected awg command: %v", args)
		}
		return []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=	51820	0	0	0	0	0	0	0	1	2	3	4	(null)	(null)	(null)	(null)	(null)	(none)	0	off
"), nil
	}

	device, err := readAWGDevice("awg0")
	if err != nil {
		t.Fatalf("readAWGDevice() error = %v", err)
	}
	if !called {
		t.Fatal("expected awg output runner to be called")
	}
	if device.ListenPort != 51820 {
		t.Fatalf("listen port = %d, want 51820", device.ListenPort)
	}
}
