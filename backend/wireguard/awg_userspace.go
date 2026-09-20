package wireguard

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/awg-go/awgctrl-go/wgtypes"
)

var awgRunner = runAWG

func runAWG(args ...string) error {
	cmd := exec.Command("awg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("awg %s: %w: %s", strings.Join(args, " "), err, msg)
		}
		return fmt.Errorf("awg %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func writeAWGConfig(config wgtypes.Config) (string, error) {
	f, err := os.CreateTemp("", "pasarguard-awg-*.conf")
	if err != nil {
		return "", fmt.Errorf("create AWG config: %w", err)
	}
	path := f.Name()
	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(path)
	}

	if err := f.Chmod(0600); err != nil {
		cleanup()
		return "", fmt.Errorf("chmod AWG config: %w", err)
	}

	var b strings.Builder
	b.WriteString("[Interface]\n")

	if config.PrivateKey != nil {
		fmt.Fprintf(&b, "PrivateKey = %s\n", config.PrivateKey.String())
	}
	if config.ListenPort != nil {
		fmt.Fprintf(&b, "ListenPort = %d\n", *config.ListenPort)
	}
	writeAWGInt(&b, "Jc", config.Jc)
	writeAWGInt(&b, "Jmin", config.Jmin)
	writeAWGInt(&b, "Jmax", config.Jmax)
	writeAWGInt(&b, "S1", config.S1)
	writeAWGInt(&b, "S2", config.S2)
	writeAWGInt(&b, "S3", config.S3)
	writeAWGInt(&b, "S4", config.S4)
	writeAWGString(&b, "H1", config.H1)
	writeAWGString(&b, "H2", config.H2)
	writeAWGString(&b, "H3", config.H3)
	writeAWGString(&b, "H4", config.H4)
	writeAWGString(&b, "I1", config.I1)
	writeAWGString(&b, "I2", config.I2)
	writeAWGString(&b, "I3", config.I3)
	writeAWGString(&b, "I4", config.I4)
	writeAWGString(&b, "I5", config.I5)

	if config.RandomTrailers != nil {
		fmt.Fprintf(&b, "RandomTrailers = %s\n", boolWord(*config.RandomTrailers))
	}
	if config.DisableCookies != nil {
		fmt.Fprintf(&b, "DisableCookies = %s\n", boolWord(*config.DisableCookies))
	}

	for _, peer := range config.Peers {
		writeAWGPeer(&b, peer)
	}

	if _, err := f.WriteString(b.String()); err != nil {
		cleanup()
		return "", fmt.Errorf("write AWG config: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close AWG config: %w", err)
	}
	return path, nil
}

func writeAWGInt(b *strings.Builder, name string, value *int) {
	if value != nil {
		fmt.Fprintf(b, "%s = %d\n", name, *value)
	}
}

func writeAWGString(b *strings.Builder, name string, value *string) {
	if value != nil {
		fmt.Fprintf(b, "%s = %s\n", name, *value)
	}
}

func boolWord(value bool) string {
	if value {
		return "on"
	}
	return "off"
}

func writeAWGPeer(b *strings.Builder, peer wgtypes.PeerConfig) {
	fmt.Fprintf(b, "[Peer]\nPublicKey = %s\n", peer.PublicKey.String())
	if peer.PresharedKey != nil {
		fmt.Fprintf(b, "PresharedKey = %s\n", peer.PresharedKey.String())
	}
	if peer.Endpoint != nil {
		fmt.Fprintf(b, "Endpoint = %s\n", formatAWGEndpoint(peer.Endpoint))
	}
	if peer.PersistentKeepaliveInterval != nil {
		fmt.Fprintf(b, "PersistentKeepalive = %d\n", int64(peer.PersistentKeepaliveInterval.Seconds()))
	}
	if peer.ReplaceAllowedIPs || len(peer.AllowedIPs) > 0 {
		fmt.Fprintf(b, "AllowedIPs = %s\n", formatAllowedIPs(peer.AllowedIPs))
	}
	fmt.Fprintf(b, "AdvancedSecurity = %s\n", boolWord(peer.AdvancedSecurity))
}

func formatAWGEndpoint(endpoint *net.UDPAddr) string {
	if endpoint == nil {
		return ""
	}
	return endpoint.String()
}

func formatAllowedIPs(allowedIPs []net.IPNet) string {
	parts := make([]string, 0, len(allowedIPs))
	for _, ip := range allowedIPs {
		parts = append(parts, ip.String())
	}
	return strings.Join(parts, ",")
}

func configureAWGWithSetconf(interfaceName string, config wgtypes.Config) error {
	path, err := writeAWGConfig(config)
	if err != nil {
		return err
	}
	defer os.Remove(path)

	if err := awgRunner("setconf", interfaceName, path); err != nil {
		return fmt.Errorf("configure AmneziaWG %s with awg setconf: %w", interfaceName, err)
	}
	return nil
}

func applyAWGPeers(interfaceName string, peers []wgtypes.PeerConfig) error {
	for _, peer := range peers {
		args := []string{"set", interfaceName, "peer", peer.PublicKey.String()}

		if peer.Remove {
			args = append(args, "remove")
		} else {
			if peer.PresharedKey != nil {
				keyPath, err := writeAWGKeyFile(peer.PresharedKey)
				if err != nil {
					return err
				}
				err = runAWG(append(args, "preshared-key", keyPath)...)
				_ = os.Remove(keyPath)
				if err != nil {
					return fmt.Errorf("configure AmneziaWG peer %s: %w", peer.PublicKey.String(), err)
				}
				args = []string{"set", interfaceName, "peer", peer.PublicKey.String()}
			}
			if peer.Endpoint != nil {
				args = append(args, "endpoint", formatAWGEndpoint(peer.Endpoint))
			}
			if peer.PersistentKeepaliveInterval != nil {
				args = append(args, "persistent-keepalive", strconv.FormatInt(int64(peer.PersistentKeepaliveInterval.Seconds()), 10))
			}
			if peer.ReplaceAllowedIPs || len(peer.AllowedIPs) > 0 {
				args = append(args, "allowed-ips", formatAllowedIPs(peer.AllowedIPs))
			}
			args = append(args, "advanced-security", boolWord(peer.AdvancedSecurity))
		}

		if err := awgRunner(args...); err != nil {
			return fmt.Errorf("configure AmneziaWG peer %s: %w", peer.PublicKey.String(), err)
		}
	}
	return nil
}

func writeAWGKeyFile(key *wgtypes.Key) (string, error) {
	f, err := os.CreateTemp("", "pasarguard-awg-key-*")
	if err != nil {
		return "", fmt.Errorf("create AWG key file: %w", err)
	}
	path := f.Name()
	if err := f.Chmod(0600); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("chmod AWG key file: %w", err)
	}
	if _, err := f.WriteString(key.String() + "\n"); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write AWG key file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close AWG key file: %w", err)
	}
	return path, nil
}

