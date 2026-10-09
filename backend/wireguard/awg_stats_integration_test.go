package wireguard

import (
	"context"
	"strings"
	"testing"

	"github.com/awg-go/awgctrl-go/wgtypes"
	"github.com/pasarguard/node/config"
	pkgstats "github.com/pasarguard/node/pkg/stats"
)

// This regression test exercises the production AWG path end to end:
// awg dump -> parser -> peer-key/email lookup -> stats tracker.
// The private key is intentionally hidden and differs from the public key,
// matching the official awg-tools dump layout.
func TestAWGTrafficCountersReachStatsTracker(t *testing.T) {
	const devicePublicKey = "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI="
	const peerPublicKey = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="

	oldRunner := awgOutputRunner
	t.Cleanup(func() { awgOutputRunner = oldRunner })

	header := []string{
		"(hidden)", devicePublicKey, "51820",
		"0", "0", "0", "0", "0", "0", "0",
		"1", "2", "3", "4",
		"(null)", "(null)", "(null)", "(null)", "(null)",
		"(none)", "0", "0", "0", "off",
	}
	for len(header) < 29 {
		header = append(header, "off")
	}
	peerRow := []string{
		peerPublicKey, "(none)", "192.0.2.10:54321",
		"10.77.0.2/32", "1700000000", "1234", "5678", "25",
	}
	awgOutputRunner = func(args ...string) ([]byte, error) {
		if strings.Join(args, " ") != "show awg0 dump" {
			t.Fatalf("unexpected AWG command: %v", args)
		}
		return []byte(strings.Join(header, "\t") + "\n" + strings.Join(peerRow, "\t") + "\n"), nil
	}

	parsedPeerKey, err := wgtypes.ParseKey(peerPublicKey)
	if err != nil {
		t.Fatalf("parse test peer key: %v", err)
	}
	peerStore := NewPeerStore()
	peerStore.Init([]*PeerInfo{{Email: "traffic-test@example.invalid", PublicKey: parsedPeerKey}})

	manager := &Manager{
		iFaceName: "awg0",
		linkType:  "amneziawg",
		client:    &fakeWGClient{},
	}
	wg := &WireGuard{
		manager:      manager,
		cfg:          &config.Config{},
		config:       &Config{},
		peerStore:    peerStore,
		statsTracker: pkgstats.New(),
	}

	wg.updateConnectedPeers(context.Background())

	entries := wg.statsTracker.GetStatsEntries([]string{peerPublicKey})
	entry, ok := entries[peerPublicKey]
	if !ok {
		t.Fatal("AWG peer was not recorded in the stats tracker")
	}
	if entry.Email != "traffic-test@example.invalid" {
		t.Fatalf("stats email = %q, want test peer email", entry.Email)
	}
	if entry.CurrentRx != 1234 || entry.CurrentTx != 5678 {
		t.Fatalf("AWG traffic counters lost: rx=%d tx=%d, want rx=1234 tx=5678", entry.CurrentRx, entry.CurrentTx)
	}
}
