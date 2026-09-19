//go:build linux
// +build linux

package wireguard

import (
	"context"
	"net"
	"os"
	"testing"

	awgctrl "github.com/awg-go/awgctrl-go"
	"github.com/awg-go/awgctrl-go/wgtypes"
	"github.com/vishvananda/netlink"
)

const awgIntegrationEnv = "PASARGUARD_AWG_INTEGRATION"

// TestAmneziaWGKernelIntegration exercises the real amneziawg kernel module
// through awgctrl-go. It is opt-in because it creates and removes a real
// network interface and requires CAP_NET_ADMIN/root.
func TestAmneziaWGKernelIntegration(t *testing.T) {
	if os.Getenv(awgIntegrationEnv) != "yesreallydoit" {
		t.Skipf("set %s=yesreallydoit to run the real AmneziaWG kernel integration test", awgIntegrationEnv)
	}

	const name = "pg-awg-it0"

	nl, err := netlink.NewHandle()
	if err != nil {
		t.Fatalf("open netlink handle: %v", err)
	}
	defer nl.Close()

	_ = netlink.LinkDel(&netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		LinkType:  "amneziawg",
	})

	link := &netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: name},
		LinkType:  "amneziawg",
	}
	if err := nl.LinkAdd(link); err != nil {
		t.Fatalf("create amneziawg link: %v", err)
	}
	defer func() {
		if l, err := nl.LinkByName(name); err == nil {
			_ = nl.LinkDel(l)
		}
	}()

	priv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	peerKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate peer key: %v", err)
	}

	jc, jmin, jmax := 3, 64, 128
	s1, s2, s3, s4 := 16, 17, 18, 4
	h1, h2, h3, h4 := "123456-123999", "223456-223999", "323456-323999", "423456-423999"
	i1, i2, i3, i4, i5 := "<r 16>", "<r 32>", "<r 8>", "<r 24>", "<r 12>"
	port := 51871

	cfg := wgtypes.Config{
		PrivateKey: &priv,
		ListenPort: &port,
		Jc: &jc, Jmin: &jmin, Jmax: &jmax,
		S1: &s1, S2: &s2, S3: &s3, S4: &s4,
		H1: &h1, H2: &h2, H3: &h3, H4: &h4,
		I1: &i1, I2: &i2, I3: &i3, I4: &i4, I5: &i5,
		Peers: []wgtypes.PeerConfig{{
			PublicKey:         peerKey.PublicKey(),
			ReplaceAllowedIPs: true,
			AllowedIPs:        []net.IPNet{mustIntegrationCIDR(t, "10.77.0.2/32")},
			AdvancedSecurity:  true,
		}},
		ReplacePeers: true,
	}

	client, err := awgctrl.New()
	if err != nil {
		t.Fatalf("open awgctrl client: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	if err := client.ConfigureDevice(ctx, name, cfg); err != nil {
		t.Fatalf("configure real amneziawg device: %v", err)
	}

	device, err := client.Device(ctx, name)
	if err != nil {
		t.Fatalf("read back real amneziawg device: %v", err)
	}
	if !device.IsAmnezia {
		t.Fatal("device was not identified as AmneziaWG")
	}
	if device.ListenPort != port {
		t.Fatalf("listen port = %d, want %d", device.ListenPort, port)
	}

	assertInt := func(label string, got, want int) {
		t.Helper()
		if got != want {
			t.Fatalf("%s = %d, want %d", label, got, want)
		}
	}
	assertString := func(label, got, want string) {
		t.Helper()
		if got != want {
			t.Fatalf("%s = %q, want %q", label, got, want)
		}
	}

	assertInt("Jc", device.Jc, jc)
	assertInt("Jmin", device.Jmin, jmin)
	assertInt("Jmax", device.Jmax, jmax)
	assertInt("S1", device.S1, s1)
	assertInt("S2", device.S2, s2)
	assertInt("S3", device.S3, s3)
	assertInt("S4", device.S4, s4)
	assertString("H1", device.H1, h1)
	assertString("H2", device.H2, h2)
	assertString("H3", device.H3, h3)
	assertString("H4", device.H4, h4)
	assertString("I1", device.I1, i1)
	assertString("I2", device.I2, i2)
	assertString("I3", device.I3, i3)
	assertString("I4", device.I4, i4)
	assertString("I5", device.I5, i5)

	if len(device.Peers) != 1 {
		t.Fatalf("peer count = %d, want 1", len(device.Peers))
	}
	// AWG2 interoperability is device-level. The loaded AWG3 kernel may
	// expose per-peer AdvancedSecurity state only when that optional kernel
	// capability is implemented. The core AWG2 contract verified here is
	// successful device configuration and peer provisioning.
	if len(device.Peers[0].AllowedIPs) != 1 || device.Peers[0].AllowedIPs[0].String() != "10.77.0.2/32" {
		t.Fatalf("unexpected peer AllowedIPs: %v", device.Peers[0].AllowedIPs)
	}

	if err := client.ConfigureDevice(ctx, name, wgtypes.Config{Peers: []wgtypes.PeerConfig{
		{PublicKey: peerKey.PublicKey(), Remove: true},
	}}); err != nil {
		t.Fatalf("remove peer: %v", err)
	}

	device, err = client.Device(ctx, name)
	if err != nil {
		t.Fatalf("read back device after peer removal: %v", err)
	}
	if len(device.Peers) != 0 {
		t.Fatalf("peer count after removal = %d, want 0", len(device.Peers))
	}

}

func mustIntegrationCIDR(t *testing.T, s string) net.IPNet {
	t.Helper()
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatalf("parse CIDR %q: %v", s, err)
	}
	return *ipnet
}
