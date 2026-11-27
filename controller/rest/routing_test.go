package rest

import (
	"testing"

	"github.com/pasarguard/node/backend"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nonRoutingBackend satisfies backend.Backend (methods promoted from the nil
// embedded interface — never called here) but NOT backend.RoutingBackend.
type nonRoutingBackend struct{ backend.Backend }

// routingFake satisfies both interfaces.
type routingFake struct {
	backend.Backend
	backend.RoutingBackend
}

func TestAsRoutingBackend(t *testing.T) {
	if _, err := asRoutingBackend(nonRoutingBackend{}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("non-routing backend: code = %v, want Unimplemented", status.Code(err))
	}
	if _, err := asRoutingBackend(routingFake{}); err != nil {
		t.Fatalf("routing backend: unexpected err %v", err)
	}
}

// A multi-core node must expose Xray routing even when Backend() is composite.
func TestAsRoutingBackendFromComposite(t *testing.T) {
	routingCore := routingFake{}
	composite := backend.NewComposite([]backend.Backend{
		nonRoutingBackend{}, routingCore,
	})
	if _, err := asRoutingBackend(composite); err != nil {
		t.Fatalf("composite Xray routing should be accessible: %v", err)
	}
	wgOnly := backend.NewComposite([]backend.Backend{nonRoutingBackend{}})
	if _, err := asRoutingBackend(wgOnly); status.Code(err) != codes.Unimplemented {
		t.Fatalf("wireguard-only composite should not offer routing: %v", err)
	}
}
