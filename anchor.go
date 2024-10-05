package antibody

import (
	"context"
	"net"
	"time"

	"github.com/sagernet/sing/common/buf"
	"github.com/xchacha20-poly1305/anchor"
)

// ProbeAnchor probe anchor service for addr.
// If ctx done and not finished, here will set conn's deadline to now to cancel it.
func ProbeAnchor(ctx context.Context, ip net.IP, conn net.PacketConn, query []byte) (*anchor.Response, error) {
	stop := context.AfterFunc(ctx, func() {
		_ = conn.SetDeadline(time.Now())
	})
	udpAddr := &net.UDPAddr{
		IP:   ip,
		Port: anchor.Port,
	}
	_, err := conn.WriteTo(query, udpAddr)
	if err != nil {
		stop()
		return nil, err
	}
	buffer := buf.NewSize(anchor.MaxResponseSize)
	_, _, err = buffer.ReadPacketFrom(conn)
	if err != nil {
		return nil, err
	}
	return anchor.ParseResponse(buffer.Bytes())
}
