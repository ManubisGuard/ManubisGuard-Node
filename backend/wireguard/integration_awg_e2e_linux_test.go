//go:build linux
// +build linux

package wireguard

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"testing"
	"time"

	awgctrl "github.com/awg-go/awgctrl-go"
	"github.com/awg-go/awgctrl-go/wgtypes"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const awgE2EIntegrationEnv = "PASARGUARD_AWG_INTEGRATION"

// TestAmneziaWGAWG2HandshakeAndTrafficIntegration provisions two real
// AmneziaWG kernel interfaces in isolated network namespaces and verifies
// an AWG2 handshake plus bidirectional UDP traffic.
//
// It is intentionally opt-in because it creates network namespaces,
// veth devices and real kernel interfaces and requires root/CAP_NET_ADMIN.
func TestAmneziaWGAWG2HandshakeAndTrafficIntegration(t *testing.T) {
	if os.Getenv(awgE2EIntegrationEnv) != "yesreallydoit" {
		t.Skipf("set %s=yesreallydoit to run the real AWG2 handshake/traffic integration test", awgE2EIntegrationEnv)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origNS, err := netns.Get()
	if err != nil {
		t.Fatalf("get original network namespace: %v", err)
	}
	defer origNS.Close()

	const (
		nsAName = "pg-awg-e2e-a"
		nsBName = "pg-awg-e2e-b"
		vethA  = "pg-awg-e2e-a0"
		vethB  = "pg-awg-e2e-b0"
		awgA   = "pg-awg-e2e0"
		awgB   = "pg-awg-e2e1"

		transportA = "192.0.2.1"
		transportB = "192.0.2.2"
		wgAddrA   = "10.77.1.1"
		wgAddrB   = "10.77.2.1"
		portA     = 51881
		portB     = 51882
	)

	_ = netns.DeleteNamed(nsAName)
	_ = netns.DeleteNamed(nsBName)

	nsA, err := netns.NewNamed(nsAName)
	if err != nil {
		t.Fatalf("create namespace A: %v", err)
	}
	defer nsA.Close()

	if err := netns.Set(origNS); err != nil {
		t.Fatalf("restore original namespace after creating A: %v", err)
	}

	nsB, err := netns.NewNamed(nsBName)
	if err != nil {
		t.Fatalf("create namespace B: %v", err)
	}
	defer nsB.Close()

	if err := netns.Set(origNS); err != nil {
		t.Fatalf("restore original namespace after creating B: %v", err)
	}
	defer func() {
		_ = netns.Set(origNS)
		_ = netns.DeleteNamed(nsAName)
		_ = netns.DeleteNamed(nsBName)
	}()

	root, err := netlink.NewHandle()
	if err != nil {
		t.Fatalf("open root netlink handle: %v", err)
	}
	defer root.Close()

	// Create a point-to-point transport network between the two namespaces.
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: vethA},
		PeerName:  vethB,
	}
	if err := root.LinkAdd(veth); err != nil {
		t.Fatalf("create veth pair: %v", err)
	}

	aLink, err := root.LinkByName(vethA)
	if err != nil {
		t.Fatalf("lookup veth A: %v", err)
	}
	bLink, err := root.LinkByName(vethB)
	if err != nil {
		t.Fatalf("lookup veth B: %v", err)
	}
	if err := root.LinkSetNsFd(aLink, int(nsA)); err != nil {
		t.Fatalf("move veth A into namespace A: %v", err)
	}
	if err := root.LinkSetNsFd(bLink, int(nsB)); err != nil {
		t.Fatalf("move veth B into namespace B: %v", err)
	}

	privA, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate private key A: %v", err)
	}
	privB, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate private key B: %v", err)
	}

	pubA := privA.PublicKey()
	pubB := privB.PublicKey()

	if err := configureAWG2Endpoint(nsA, vethA, awgA, transportA, transportB, wgAddrA, wgAddrB, portA, portB, privA, pubB); err != nil {
		t.Fatalf("configure AWG2 endpoint A: %v", err)
	}
	if err := configureAWG2Endpoint(nsB, vethB, awgB, transportB, transportA, wgAddrB, wgAddrA, portB, portA, privB, pubA); err != nil {
		t.Fatalf("configure AWG2 endpoint B: %v", err)
	}

	stopEcho := make(chan struct{})
	echoReady := make(chan error, 1)
	echoDone := make(chan struct{})
	go runAWG2EchoServer(nsB, wgAddrB, 40001, stopEcho, echoReady, echoDone)

	if err := <-echoReady; err != nil {
		t.Fatalf("start AWG2 echo server: %v", err)
	}
	defer func() {
		close(stopEcho)
		<-echoDone
	}()

	if err := netns.Set(nsA); err != nil {
		t.Fatalf("switch to namespace A: %v", err)
	}

	conn, err := net.DialUDP("udp4",
		&net.UDPAddr{IP: net.ParseIP(wgAddrA)},
		&net.UDPAddr{IP: net.ParseIP(wgAddrB), Port: 40001},
	)
	if err != nil {
		t.Fatalf("open UDP socket in namespace A: %v", err)
	}
	defer conn.Close()

	payload := []byte("pasarguard-awg2-e2e")
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("send UDP payload through AWG2: %v", err)
	}

	buf := make([]byte, 128)
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("receive echoed UDP payload through AWG2: %v", err)
	}
	if string(buf[:n]) != string(payload) {
		t.Fatalf("echo payload = %q, want %q", string(buf[:n]), string(payload))
	}

	if err := assertAWG2Handshake(nsA, awgA, pubB); err != nil {
		t.Fatalf("endpoint A handshake verification: %v", err)
	}
	if err := assertAWG2Handshake(nsB, awgB, pubA); err != nil {
		t.Fatalf("endpoint B handshake verification: %v", err)
	}

	if err := assertAWG2TrafficCounters(nsA, awgA, pubB); err != nil {
		t.Fatalf("endpoint A traffic counters: %v", err)
	}
	if err := assertAWG2TrafficCounters(nsB, awgB, pubA); err != nil {
		t.Fatalf("endpoint B traffic counters: %v", err)
	}

	if err := netns.Set(origNS); err != nil {
		t.Fatalf("restore original namespace: %v", err)
	}
}

func configureAWG2Endpoint(ns netns.NsHandle, vethName, awgName, localTransport, remoteTransport, localWG, remoteWG string, listenPort int, remotePort int, privateKey, peerPublicKey wgtypes.Key) error {
	if err := netns.Set(ns); err != nil {
		return err
	}

	nl, err := netlink.NewHandle()
	if err != nil {
		return err
	}
	defer nl.Close()

	veth, err := nl.LinkByName(vethName)
	if err != nil {
		return err
	}

	transportIP := &net.IPNet{IP: net.ParseIP(localTransport), Mask: net.CIDRMask(30, 32)}
	if err := nl.AddrAdd(veth, &netlink.Addr{IPNet: transportIP}); err != nil {
		return fmt.Errorf("add transport address %s: %w", localTransport, err)
	}
	if err := nl.LinkSetUp(veth); err != nil {
		return fmt.Errorf("set transport link up: %w", err)
	}

	awgLink := &netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: awgName},
		LinkType:  "amneziawg",
	}
	if err := nl.LinkAdd(awgLink); err != nil {
		return fmt.Errorf("add amneziawg link: %w", err)
	}
	if err := nl.LinkSetUp(awgLink); err != nil {
		return fmt.Errorf("set amneziawg link up: %w", err)
	}

	wgIP := &net.IPNet{IP: net.ParseIP(localWG), Mask: net.CIDRMask(32, 32)}
	if err := nl.AddrAdd(awgLink, &netlink.Addr{IPNet: wgIP}); err != nil {
		return fmt.Errorf("add AWG address %s: %w", localWG, err)
	}

	remoteWGIP := &net.IPNet{IP: net.ParseIP(remoteWG), Mask: net.CIDRMask(32, 32)}
	if err := nl.RouteAdd(&netlink.Route{
		LinkIndex: awgLink.Attrs().Index,
		Dst:       remoteWGIP,
		Scope:     netlink.SCOPE_LINK,
	}); err != nil {
		return fmt.Errorf("add AWG route to %s: %w", remoteWG, err)
	}

	jc, jmin, jmax := 3, 64, 128
	s1, s2, s3, s4 := 16, 17, 18, 4
	h1, h2, h3, h4 := "123456-123999", "223456-223999", "323456-323999", "423456-423999"
	i1, i2, i3, i4, i5 := "<r 16>", "<r 32>", "<r 8>", "<r 24>", "<r 12>"
	keepalive := 1 * time.Second

	cfg := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
		Jc: &jc, Jmin: &jmin, Jmax: &jmax,
		S1: &s1, S2: &s2, S3: &s3, S4: &s4,
		H1: &h1, H2: &h2, H3: &h3, H4: &h4,
		I1: &i1, I2: &i2, I3: &i3, I4: &i4, I5: &i5,
		Peers: []wgtypes.PeerConfig{{
			PublicKey:         peerPublicKey,
			Endpoint:           &net.UDPAddr{IP: net.ParseIP(remoteTransport), Port: remotePort},
			PersistentKeepaliveInterval: &keepalive,
			ReplaceAllowedIPs: true,
			AllowedIPs:        []net.IPNet{*remoteWGIP},
			AdvancedSecurity:  true,
		}},
		ReplacePeers: true,
	}

	client, err := awgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	// Configure the device first, then add the peer separately. This makes
	// kernel/netlink failures attributable to the device or peer operation.
	deviceCfg := cfg
	deviceCfg.Peers = nil
	deviceCfg.ReplacePeers = false
	if err := client.ConfigureDevice(context.Background(), awgName, deviceCfg); err != nil {
		return fmt.Errorf("ConfigureDevice(%s) device config: %w", awgName, err)
	}

	// Add the peer in progressively richer stages so an EINVAL can be
	// attributed to one specific peer attribute.
	basePeer := cfg.Peers[0]
	basePeer.Endpoint = nil
	basePeer.PersistentKeepaliveInterval = nil
	basePeer.AdvancedSecurity = false
	peerCfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{basePeer},
		ReplacePeers: true,
	}
	if err := client.ConfigureDevice(context.Background(), awgName, peerCfg); err != nil {
		return fmt.Errorf("ConfigureDevice(%s) peer base (key/allowed-ips): %w", awgName, err)
	}

	endpoint := cfg.Peers[0].Endpoint
	if err := client.ConfigureDevice(context.Background(), awgName, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{{
			PublicKey: cfg.Peers[0].PublicKey,
			Endpoint: endpoint,
		}},
	}); err != nil {
		return fmt.Errorf("ConfigureDevice(%s) peer endpoint: %w", awgName, err)
	}

	keepalivePtr := cfg.Peers[0].PersistentKeepaliveInterval
	if err := client.ConfigureDevice(context.Background(), awgName, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{{
			PublicKey: cfg.Peers[0].PublicKey,
			PersistentKeepaliveInterval: keepalivePtr,
		}},
	}); err != nil {
		return fmt.Errorf("ConfigureDevice(%s) peer keepalive: %w", awgName, err)
	}

	if err := client.ConfigureDevice(context.Background(), awgName, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{{
			PublicKey: cfg.Peers[0].PublicKey,
			AdvancedSecurity: true,
		}},
	}); err != nil {
		return fmt.Errorf("ConfigureDevice(%s) peer advanced-security: %w", awgName, err)
	}
	return nil
}

func runAWG2EchoServer(ns netns.NsHandle, bindIP string, port int, stop <-chan struct{}, ready chan<- error, done chan<- struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(done)

	if err := netns.Set(ns); err != nil {
		ready <- err
		return
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(bindIP), Port: port})
	if err != nil {
		ready <- err
		return
	}
	defer conn.Close()

	ready <- nil

	buf := make([]byte, 2048)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				select {
				case <-stop:
					return
				default:
					continue
				}
			}
			return
		}
		if _, err := conn.WriteToUDP(buf[:n], addr); err != nil {
			return
		}
	}
}

func assertAWG2Handshake(ns netns.NsHandle, awgName string, peerPublicKey wgtypes.Key) error {
	if err := netns.Set(ns); err != nil {
		return err
	}

	client, err := awgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		device, err := client.Device(context.Background(), awgName)
		if err != nil {
			return err
		}
		for _, peer := range device.Peers {
			if peer.PublicKey == peerPublicKey {
				if peer.LastHandshakeTime.IsZero() {
					time.Sleep(100 * time.Millisecond)
					break
				}
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("peer %s has no completed handshake within 5s", peerPublicKey.String())
}

func assertAWG2TrafficCounters(ns netns.NsHandle, awgName string, peerPublicKey wgtypes.Key) error {
	if err := netns.Set(ns); err != nil {
		return err
	}

	client, err := awgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	device, err := client.Device(context.Background(), awgName)
	if err != nil {
		return err
	}
	for _, peer := range device.Peers {
		if peer.PublicKey == peerPublicKey {
			if peer.ReceiveBytes <= 0 || peer.TransmitBytes <= 0 {
				return fmt.Errorf("peer counters RX=%d TX=%d, expected both > 0", peer.ReceiveBytes, peer.TransmitBytes)
			}
			return nil
		}
	}
	return fmt.Errorf("peer %s not found", peerPublicKey.String())
}
