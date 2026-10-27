package wireguard

import (
	"errors"
	"strings"
	"testing"

	"github.com/advanced-wg/awgctrl-go/wgtypes"
	"github.com/vishvananda/netlink"
)

func TestManagerUsesAmneziaWGLinkKind(t *testing.T) {
	var added netlink.Link
	lookups := 0
	manager := &Manager{
		iFaceName: "awg-test", client: &fakeWGClient{}, amneziaWG: true,
		nl: mockNetlinkOps{
			parseAddrFn: func(string) (*netlink.Addr, error) { return &netlink.Addr{}, nil },
			linkAddFn:   func(link netlink.Link) error { added = link; return nil },
			linkByName: func(name string) (netlink.Link, error) {
				lookups++
				if lookups == 1 {
					return nil, linkNotFoundError()
				}
				return &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name}}, nil
			},
			linkSetUpFn: func(netlink.Link) error { return nil },
			linkDelFn:   func(netlink.Link) error { return nil },
		},
	}
	if err := manager.InitializeWithPeers(wgtypes.Key{}, 51830, nil, nil); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	link, ok := added.(*netlink.GenericLink)
	if !ok || link.LinkType != "amneziawg" {
		t.Fatalf("expected AmneziaWG link, got %#v", added)
	}
}

func TestSetPeerAdvancedSecurityEnablesAWGPeers(t *testing.T) {
	peers := setPeerAdvancedSecurity([]wgtypes.PeerConfig{{}}, true)
	if len(peers) != 1 || !peers[0].AdvancedSecurity {
		t.Fatalf("expected AWG advanced-security flag to be set, got %#v", peers)
	}
	peers = setPeerAdvancedSecurity(peers, false)
	if peers[0].AdvancedSecurity {
		t.Fatal("expected standard WireGuard peer flag to remain disabled")
	}
}

func TestAmneziaWGInterfaceFailureExplainsHostModuleRequirement(t *testing.T) {
	manager := &Manager{
		iFaceName: "awg-test", client: &fakeWGClient{}, amneziaWG: true,
		nl: mockNetlinkOps{
			linkAddFn:  func(netlink.Link) error { return errors.New("operation not supported") },
			linkByName: func(string) (netlink.Link, error) { return nil, linkNotFoundError() },
		},
	}
	err := manager.InitializeWithPeers(wgtypes.Key{}, 51830, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "AmneziaWG module installed and loaded") {
		t.Fatalf("expected explicit host module requirement, got: %v", err)
	}
}
