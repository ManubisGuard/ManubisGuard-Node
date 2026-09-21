package wireguard

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/awg-go/awgctrl-go/wgtypes"
)

func readAWGDevice(interfaceName string) (*wgtypes.Device, error) {
	out, err := awgOutputRunner("show", interfaceName, "dump")
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return nil, fmt.Errorf("awg show %s dump: %w: %s", interfaceName, err, msg)
		}
		return nil, fmt.Errorf("awg show %s dump: %w", interfaceName, err)
	}
	return parseAWGDeviceDump(interfaceName, string(out))
}

func parseAWGDeviceDump(interfaceName, output string) (*wgtypes.Device, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, fmt.Errorf("awg show %s dump returned empty output", interfaceName)
	}

	fields := strings.Split(lines[0], "\t")
	if len(fields) < 4 {
		return nil, fmt.Errorf("invalid AWG device dump: expected at least 4 interface fields, got %d", len(fields))
	}

	listenPort, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, fmt.Errorf("invalid AWG listen port %q: %w", fields[3], err)
	}

	device := &wgtypes.Device{
		Name:       interfaceName,
		Type:       wgtypes.LinuxKernel,
		ListenPort: listenPort,
	}

	if fields[2] != "(none)" && fields[2] != "(hidden)" {
		key, err := wgtypes.ParseKey(fields[2])
		if err != nil {
			return nil, fmt.Errorf("invalid AWG public key: %w", err)
		}
		device.PublicKey = key
	}

	if fields[1] != "(none)" && fields[1] != "(hidden)" {
		key, err := wgtypes.ParseKey(fields[1])
		if err != nil {
			return nil, fmt.Errorf("invalid AWG private key: %w", err)
		}
		device.PrivateKey = key
	}

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		peer, err := parseAWGPeerDump(line)
		if err != nil {
			return nil, err
		}
		device.Peers = append(device.Peers, peer)
	}

	return device, nil
}

func parseAWGPeerDump(line string) (wgtypes.Peer, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 8 {
		return wgtypes.Peer{}, fmt.Errorf("invalid AWG peer dump: expected at least 8 fields, got %d", len(fields))
	}

	publicKey, err := wgtypes.ParseKey(fields[0])
	if err != nil {
		return wgtypes.Peer{}, fmt.Errorf("invalid AWG peer public key: %w", err)
	}

	peer := wgtypes.Peer{PublicKey: publicKey}

	if fields[1] != "(none)" && fields[1] != "(hidden)" {
		psk, err := wgtypes.ParseKey(fields[1])
		if err != nil {
			return wgtypes.Peer{}, fmt.Errorf("invalid AWG peer preshared key: %w", err)
		}
		peer.PresharedKey = psk
	}

	if fields[2] != "(none)" {
		endpoint, err := net.ResolveUDPAddr("udp", fields[2])
		if err != nil {
			return wgtypes.Peer{}, fmt.Errorf("invalid AWG peer endpoint %q: %w", fields[2], err)
		}
		peer.Endpoint = endpoint
	}

	if fields[3] != "(none)" {
		for _, rawCIDR := range strings.Split(fields[3], ",") {
			rawCIDR = strings.TrimSpace(rawCIDR)
			if rawCIDR == "" {
				continue
			}
			_, ipnet, err := net.ParseCIDR(rawCIDR)
			if err != nil {
				return wgtypes.Peer{}, fmt.Errorf("invalid AWG allowed IP %q: %w", rawCIDR, err)
			}
			peer.AllowedIPs = append(peer.AllowedIPs, *ipnet)
		}
	}

	handshake, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return wgtypes.Peer{}, fmt.Errorf("invalid AWG handshake timestamp %q: %w", fields[4], err)
	}
	if handshake > 0 {
		peer.LastHandshakeTime = time.Unix(handshake, 0)
	}

	receiveBytes, err := strconv.ParseInt(fields[5], 10, 64)
	if err != nil {
		return wgtypes.Peer{}, fmt.Errorf("invalid AWG receive bytes %q: %w", fields[5], err)
	}
	transmitBytes, err := strconv.ParseInt(fields[6], 10, 64)
	if err != nil {
		return wgtypes.Peer{}, fmt.Errorf("invalid AWG transmit bytes %q: %w", fields[6], err)
	}
	peer.ReceiveBytes = receiveBytes
	peer.TransmitBytes = transmitBytes

	if fields[7] != "off" {
		seconds, err := strconv.ParseInt(fields[7], 10, 64)
		if err != nil {
			return wgtypes.Peer{}, fmt.Errorf("invalid AWG persistent keepalive %q: %w", fields[7], err)
		}
		peer.PersistentKeepaliveInterval = time.Duration(seconds) * time.Second
	}

	return peer, nil
}
