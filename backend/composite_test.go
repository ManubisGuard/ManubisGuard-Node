package backend

import "testing"

type passiveCompositeCore struct{ Backend }
type routingCompositeCore struct {
	Backend
	RoutingBackend
}

func TestCompositeRoutingBackendSelection(t *testing.T) {
	mixed := NewComposite([]Backend{passiveCompositeCore{}, routingCompositeCore{}})
	if mixed.RoutingBackend() == nil {
		t.Fatal("mixed node must expose its Xray routing capability")
	}

	wgOnly := NewComposite([]Backend{passiveCompositeCore{}})
	if wgOnly.RoutingBackend() != nil {
		t.Fatal("WireGuard-only node must not expose Xray routing")
	}
}
