package libp2pvsock

import (
	"context"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/transport"

	mavs "github.com/balena/go-multiaddr-vsock"
	mavsnet "github.com/balena/go-multiaddr-vsock/net"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"
	manet "github.com/multiformats/go-multiaddr/net"
)

var log = logging.Logger("vsock-tpt")

type Option func(*VsockTransport) error

// VsockTransport is the VSOCK transport.
type VsockTransport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	upgrader transport.Upgrader

	rcmgr network.ResourceManager
}

var _ transport.Transport = &VsockTransport{}

// New creates a VSOCK transport object that tracks dialers and listeners
// created.
func New(upgrader transport.Upgrader, rcmgr network.ResourceManager, opts ...Option) (*VsockTransport, error) {
	if rcmgr == nil {
		rcmgr = &network.NullResourceManager{}
	}
	tr := &VsockTransport{
		upgrader: upgrader,
		rcmgr:    rcmgr,
	}
	for _, o := range opts {
		if err := o(tr); err != nil {
			return nil, err
		}
	}
	return tr, nil
}

var dialMatcher = mafmt.And(mafmt.Base(mavs.P_VSOCK), mafmt.Base(mavs.P_XTCP))

// CanDial returns true if this transport believes it can dial the given
// multiaddr.
func (t *VsockTransport) CanDial(addr ma.Multiaddr) bool {
	return dialMatcher.Matches(addr)
}

func (t *VsockTransport) mavsnetDial(ctx context.Context, raddr ma.Multiaddr) (manet.Conn, error) {
	// Context is discarded for now in the hope VSOCK connections
	// complete as fast as possible.
	return mavsnet.DialContext(ctx, raddr)
}

// Dial dials the peer at the remote address.
func (t *VsockTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	connScope, err := t.rcmgr.OpenConnection(network.DirOutbound, true, raddr)
	if err != nil {
		log.Debugw("resource manager blocked outgoing connection", "peer", p, "addr", raddr, "error", err)
		return nil, err
	}

	c, err := t.dialWithScope(ctx, raddr, p, connScope)
	if err != nil {
		connScope.Done()
		return nil, err
	}
	return c, nil
}

func (t *VsockTransport) dialWithScope(ctx context.Context, raddr ma.Multiaddr, p peer.ID, connScope network.ConnManagementScope) (transport.CapableConn, error) {
	if err := connScope.SetPeer(p); err != nil {
		log.Debugw("resource manager blocked outgoing connection for peer", "peer", p, "addr", raddr, "error", err)
		return nil, err
	}
	conn, err := t.mavsnetDial(ctx, raddr)
	if err != nil {
		return nil, err
	}
	direction := network.DirOutbound
	if ok, isClient, _ := network.GetSimultaneousConnect(ctx); ok && !isClient {
		direction = network.DirInbound
	}
	return t.upgrader.Upgrade(ctx, t, conn, direction, p, connScope)
}

func (t *VsockTransport) mavsnetListen(laddr ma.Multiaddr) (manet.Listener, error) {
	return mavsnet.Listen(laddr)
}

// Listen listens on the given multiaddr.
func (t *VsockTransport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	list, err := t.mavsnetListen(laddr)
	if err != nil {
		return nil, err
	}
	return t.upgrader.UpgradeListener(t, list), nil
}

// Protocols returns the list of terminal protocols this transport can dial.
func (t *VsockTransport) Protocols() []int {
	return []int{mavs.P_XTCP}
}

// Proxy always returns false for the VSOCK transport.
func (t *VsockTransport) Proxy() bool {
	return false
}

func (t *VsockTransport) String() string {
	return "VSOCK"
}
