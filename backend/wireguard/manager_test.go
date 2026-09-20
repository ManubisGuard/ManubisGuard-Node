package wireguard

import (
	"errors"
	"os"
	"net"
	"strings"
	"testing"

	"github.com/vishvananda/netlink"
	"github.com/awg-go/awgctrl-go/wgtypes"
)

func linkNotFoundError() error {
	var err netlink.LinkNotFoundError
	return err
}

type fakeWGClient struct {
	configureDeviceFn func(name string, cfg wgtypes.Config) error
	deviceFn          func(name string) (*wgtypes.Device, error)
	closeFn           func() error
}

func (c *fakeWGClient) ConfigureDevice(name string, cfg wgtypes.Config) error {
	if c.configureDeviceFn == nil {
		return nil
	}
	return c.configureDeviceFn(name, cfg)
}

func (c *fakeWGClient) Device(name string) (*wgtypes.Device, error) {
	if c.deviceFn == nil {
		return &wgtypes.Device{}, nil
	}
	return c.deviceFn(name)
}

func (c *fakeWGClient) Close() error {
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

type mockNetlinkOps struct {
	parseAddrFn func(string) (*netlink.Addr, error)
	linkAddFn   func(netlink.Link) error
	linkByName  func(string) (netlink.Link, error)
	addrAddFn   func(netlink.Link, *netlink.Addr) error
	linkSetUpFn func(netlink.Link) error
	linkDelFn   func(netlink.Link) error
}

func (m mockNetlinkOps) ParseAddr(address string) (*netlink.Addr, error) {
	if m.parseAddrFn == nil {
		return nil, errors.New("ParseAddr was not mocked")
	}
	return m.parseAddrFn(address)
}

func (m mockNetlinkOps) LinkAdd(link netlink.Link) error {
	if m.linkAddFn == nil {
		return errors.New("LinkAdd was not mocked")
	}
	return m.linkAddFn(link)
}

func (m mockNetlinkOps) LinkByName(name string) (netlink.Link, error) {
	if m.linkByName == nil {
		return nil, errors.New("LinkByName was not mocked")
	}
	return m.linkByName(name)
}

func (m mockNetlinkOps) AddrAdd(link netlink.Link, addr *netlink.Addr) error {
	if m.addrAddFn == nil {
		return errors.New("AddrAdd was not mocked")
	}
	return m.addrAddFn(link, addr)
}

func (m mockNetlinkOps) LinkSetUp(link netlink.Link) error {
	if m.linkSetUpFn == nil {
		return errors.New("LinkSetUp was not mocked")
	}
	return m.linkSetUpFn(link)
}

func (m mockNetlinkOps) LinkDel(link netlink.Link) error {
	if m.linkDelFn == nil {
		return errors.New("LinkDel was not mocked")
	}
	return m.linkDelFn(link)
}

func TestManagerInitializeConfigureFailureCleansInterface(t *testing.T) {
	linkByNameCalls := 0
	linkDelCalls := 0

	mock := mockNetlinkOps{
		parseAddrFn: func(_ string) (*netlink.Addr, error) {
			return &netlink.Addr{}, nil
		},
		linkAddFn: func(_ netlink.Link) error {
			return nil
		},
		linkByName: func(name string) (netlink.Link, error) {
			linkByNameCalls++
			if linkByNameCalls == 1 {
				return nil, linkNotFoundError()
			}
			return &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name}}, nil
		},
		addrAddFn: func(_ netlink.Link, _ *netlink.Addr) error {
			return nil
		},
		linkSetUpFn: func(_ netlink.Link) error {
			return nil
		},
		linkDelFn: func(_ netlink.Link) error {
			linkDelCalls++
			return nil
		},
	}

	manager := &Manager{
		iFaceName: "wg-test",
		client:    &fakeWGClient{},
		nl:        mock,
		configure: func(_ wgClient, _ string, _ wgtypes.Config) error {
			return errors.New("configure boom")
		},
	}
	err := manager.InitializeWithPeers(wgtypes.Key{}, 51820, []string{"10.0.0.1/24"}, nil)
	if err == nil {
		t.Fatal("expected initialize error, got nil")
	}
	if !strings.Contains(err.Error(), "configure boom") {
		t.Fatalf("unexpected error: %v", err)
	}
	if linkDelCalls != 1 {
		t.Fatalf("expected cleanup link delete to be called once, got %d", linkDelCalls)
	}
}

func TestManagerInitializeAddrAddFailureCleansInterface(t *testing.T) {
	linkByNameCalls := 0
	linkDelCalls := 0

	mock := mockNetlinkOps{
		parseAddrFn: func(_ string) (*netlink.Addr, error) {
			return &netlink.Addr{}, nil
		},
		linkAddFn: func(_ netlink.Link) error {
			return nil
		},
		linkByName: func(name string) (netlink.Link, error) {
			linkByNameCalls++
			if linkByNameCalls == 1 {
				return nil, linkNotFoundError()
			}
			return &netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{Name: name, Flags: net.FlagUp},
			}, nil
		},
		addrAddFn: func(_ netlink.Link, _ *netlink.Addr) error {
			return errors.New("addr add boom")
		},
		linkSetUpFn: func(_ netlink.Link) error {
			return nil
		},
		linkDelFn: func(_ netlink.Link) error {
			linkDelCalls++
			return nil
		},
	}

	manager := &Manager{
		iFaceName: "wg-test",
		client:    &fakeWGClient{},
		nl:        mock,
		configure: func(_ wgClient, _ string, _ wgtypes.Config) error {
			return nil
		},
	}
	err := manager.InitializeWithPeers(wgtypes.Key{}, 51820, []string{"10.0.0.1/24"}, nil)
	if err == nil {
		t.Fatal("expected initialize error, got nil")
	}
	if !strings.Contains(err.Error(), "addr add boom") {
		t.Fatalf("unexpected error: %v", err)
	}
	if linkDelCalls != 1 {
		t.Fatalf("expected cleanup link delete to be called once, got %d", linkDelCalls)
	}
}

func TestManagerInitializeWithPeersConfiguresReplacePeers(t *testing.T) {
	linkByNameCalls := 0
	configureCalls := 0

	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	parsedKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}
	peers := []wgtypes.PeerConfig{
		{PublicKey: parsedKey},
	}

	mock := mockNetlinkOps{
		parseAddrFn: func(_ string) (*netlink.Addr, error) {
			return &netlink.Addr{}, nil
		},
		linkAddFn: func(_ netlink.Link) error {
			return nil
		},
		linkByName: func(name string) (netlink.Link, error) {
			linkByNameCalls++
			if linkByNameCalls == 1 {
				return nil, linkNotFoundError()
			}
			return &netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{Name: name, Flags: net.FlagUp},
			}, nil
		},
		addrAddFn: func(_ netlink.Link, _ *netlink.Addr) error {
			return nil
		},
		linkSetUpFn: func(_ netlink.Link) error {
			return nil
		},
		linkDelFn: func(_ netlink.Link) error {
			return nil
		},
	}

	manager := &Manager{
		iFaceName: "wg-test",
		client:    &fakeWGClient{},
		nl:        mock,
		configure: func(_ wgClient, _ string, cfg wgtypes.Config) error {
			configureCalls++
			if cfg.PrivateKey == nil {
				t.Fatal("expected private key to be set")
			}
			if cfg.ListenPort == nil {
				t.Fatal("expected listen port to be set")
			}
			if !cfg.ReplacePeers {
				t.Fatal("expected ReplacePeers=true for initialize with peers")
			}
			if len(cfg.Peers) != len(peers) {
				t.Fatalf("unexpected peers length: got %d want %d", len(cfg.Peers), len(peers))
			}
			return nil
		},
	}

	err = manager.InitializeWithPeers(wgtypes.Key{}, 51820, []string{"10.0.0.1/24"}, peers)
	if err != nil {
		t.Fatalf("unexpected initialize error: %v", err)
	}
	if configureCalls != 1 {
		t.Fatalf("expected one configure call, got %d", configureCalls)
	}
}

func TestManagerCleanupExistingInterfaceReturnsLookupError(t *testing.T) {
	manager := &Manager{
		iFaceName: "wg-test",
		nl: mockNetlinkOps{
			linkByName: func(_ string) (netlink.Link, error) {
				return nil, errors.New("permission denied")
			},
		},
	}

	err := manager.cleanupExistingInterface()
	if err == nil {
		t.Fatal("expected cleanup error, got nil")
	}
	if !strings.Contains(err.Error(), "link lookup") {
		t.Fatalf("unexpected cleanup error: %v", err)
	}
}

func TestManagerInitializeNilClient(t *testing.T) {
	manager := &Manager{
		iFaceName: "wg-test",
	}

	err := manager.InitializeWithPeers(wgtypes.Key{}, 51820, []string{"10.0.0.1/24"}, nil)
	if err == nil {
		t.Fatal("expected initialize error, got nil")
	}
	if !strings.Contains(err.Error(), "wgctrl client is not initialized") {
		t.Fatalf("unexpected initialize error: %v", err)
	}
}

func TestManagerGetDeviceNilClient(t *testing.T) {
	manager := &Manager{
		iFaceName: "wg-test",
	}

	_, err := manager.GetDevice()
	if err == nil {
		t.Fatal("expected get device error, got nil")
	}
	if !strings.Contains(err.Error(), "wgctrl client is not initialized") {
		t.Fatalf("unexpected get device error: %v", err)
	}
}

func TestManagerApplyPeersReplaceAllSetsReplacePeers(t *testing.T) {
	calls := 0
	manager := &Manager{
		iFaceName: "wg-test",
		client: &fakeWGClient{
			configureDeviceFn: func(name string, cfg wgtypes.Config) error {
				calls++
				if name != "wg-test" {
					t.Fatalf("unexpected interface: %s", name)
				}
				if !cfg.ReplacePeers {
					t.Fatal("expected ReplacePeers=true")
				}
				if len(cfg.Peers) != 1 {
					t.Fatalf("expected one peer config, got %d", len(cfg.Peers))
				}
				return nil
			},
		},
	}

	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	parsed, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		t.Fatalf("failed to parse key: %v", err)
	}

	err = manager.ApplyPeersReplaceAll([]wgtypes.PeerConfig{
		{PublicKey: parsed},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one configure call, got %d", calls)
	}
}

func TestManagerApplyPeersReplaceAllNilClient(t *testing.T) {
	manager := &Manager{iFaceName: "wg-test"}
	err := manager.ApplyPeersReplaceAll(nil)
	if err == nil {
		t.Fatal("expected error with nil client")
	}
	if !strings.Contains(err.Error(), "wgctrl client is not initialized") {
		t.Fatalf("unexpected error: %v", err)
	}
}


func TestManagerInitializeAmneziaWGCreatesAmneziaLinkAndConfiguresSecurity(t *testing.T) {
	var addedLink netlink.Link
	var awgArgs []string
	var configText string
	oldRunner := awgRunner
	awgRunner = func(args ...string) error {
		awgArgs = append([]string(nil), args...)
		if len(args) == 3 && args[0] == "setconf" {
			data, err := os.ReadFile(args[2])
			if err != nil {
				t.Fatalf("read generated AWG config: %v", err)
			}
			configText = string(data)
		}
		return nil
	}
	defer func() { awgRunner = oldRunner }()

	mock := mockNetlinkOps{
		parseAddrFn: func(_ string) (*netlink.Addr, error) { return &netlink.Addr{}, nil },
		linkAddFn: func(link netlink.Link) error {
			addedLink = link
			return nil
		},
		linkByName: func(name string) (netlink.Link, error) {
			return &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: name, Flags: net.FlagUp}}, nil
		},
		addrAddFn: func(_ netlink.Link, _ *netlink.Addr) error { return nil },
		linkSetUpFn: func(_ netlink.Link) error { return nil },
		linkDelFn: func(_ netlink.Link) error { return nil },
	}

	manager := &Manager{
		iFaceName: "awg-test",
		linkType: "amneziawg",
		client: &fakeWGClient{},
		nl: mock,
		configure: defaultConfigureDevice,
	}

	jc, jmin, jmax := 3, 64, 128
	s1, s2, s3, s4 := 16, 17, 18, 4
	h1, h2, h3, h4 := "123456-123999", "223456-223999", "323456-323999", "423456-423999"
	i1, i2, i3, i4, i5 := "<r 16>", "<r 32>", "<r 8>", "<r 24>", "<r 12>"
	extra := wgtypes.Config{
		Jc: &jc, Jmin: &jmin, Jmax: &jmax,
		S1: &s1, S2: &s2, S3: &s3, S4: &s4,
		H1: &h1, H2: &h2, H3: &h3, H4: &h4,
		I1: &i1, I2: &i2, I3: &i3, I4: &i4, I5: &i5,
	}

	if err := manager.InitializeWithPeersAndConfig(wgtypes.Key{}, 51820, []string{"10.0.0.1/24"}, nil, extra); err != nil {
		t.Fatalf("unexpected initialize error: %v", err)
	}

	generic, ok := addedLink.(*netlink.GenericLink)
	if !ok {
		t.Fatalf("expected GenericLink for AmneziaWG, got %T", addedLink)
	}
	if generic.LinkType != "amneziawg" {
		t.Fatalf("link type = %q, want amneziawg", generic.LinkType)
	}
	if len(awgArgs) != 3 || awgArgs[0] != "setconf" || awgArgs[1] != "awg-test" {
		t.Fatalf("unexpected awg invocation: %v", awgArgs)
	}
	for _, want := range []string{
		"Jc = 3", "Jmin = 64", "Jmax = 128",
		"S1 = 16", "S4 = 4",
		"H1 = 123456-123999", "H4 = 423456-423999",
		"I1 = <r 16>", "I5 = <r 12>",
	} {
		if !strings.Contains(configText, want) {
			t.Fatalf("generated AWG config missing %q:\n%s", want, configText)
		}
	}
}


func TestReplaceAWGPeersUsesOutputRunner(t *testing.T) {
	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}
	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}

	oldOutputRunner := awgOutputRunner
	oldRunner := awgRunner
	defer func() {
		awgOutputRunner = oldOutputRunner
		awgRunner = oldRunner
	}()

	var setconfPath string
	awgOutputRunner = func(args ...string) ([]byte, error) {
		if len(args) != 2 || args[0] != "showconf" || args[1] != "awg-test" {
			t.Fatalf("unexpected showconf args: %v", args)
		}
		return []byte("[Interface]\nListenPort = 51820\n"), nil
	}
	awgRunner = func(args ...string) error {
		if len(args) != 3 || args[0] != "setconf" || args[1] != "awg-test" {
			t.Fatalf("unexpected setconf args: %v", args)
		}
		setconfPath = args[2]
		data, err := os.ReadFile(setconfPath)
		if err != nil {
			return err
		}
		text := string(data)
		if !strings.Contains(text, "[Interface]") || !strings.Contains(text, "ListenPort = 51820") {
			t.Fatalf("replacement config missing interface settings: %s", text)
		}
		if !strings.Contains(text, "PublicKey = "+key.String()) {
			t.Fatalf("replacement config missing peer: %s", text)
		}
		return nil
	}

	if err := replaceAWGPeers("awg-test", []wgtypes.PeerConfig{{PublicKey: key}}); err != nil {
		t.Fatalf("replaceAWGPeers() error = %v", err)
	}
	if setconfPath == "" {
		t.Fatal("expected setconf to be invoked")
	}
}
