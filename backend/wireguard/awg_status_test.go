package wireguard

import (
	"strings"
	"testing"
	"time"
)

func TestParseAWGDeviceDump(t *testing.T) {
	const key = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	const peerKey = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="
	const dump = key + "\t" + key + "\t" + key + "\t51820\t3\t20\t50\t15\t64\t25\t8\t1-2\t3-4\t5-6\t7-8\t(null)\t(null)\t(null)\t(null)\t(null)\t(none)\t0\t0\t0\toff\n" +
		peerKey + "\t(none)\t192.0.2.1:54321\t10.0.0.2/32,10.0.0.3/32\t1700000000\t1234\t5678\t25\n"

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
	if peer.PersistentKeepaliveInterval == nil || *peer.PersistentKeepaliveInterval != 25*time.Second {
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
		return []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\tAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\tAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\t51820\t0\t0\t0\t0\t0\t0\t0\t1\t2\t3\t4\t(null)\t(null)\t(null)\t(null)\t(null)\t(none)\t0\toff\n"), nil
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
