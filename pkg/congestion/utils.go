package congestion

import (
	"github.com/SianHH/frp-package/pkg/congestion/bbr"
	"github.com/apernet/quic-go"
)

func UseBBR(conn *quic.Conn) {
	conn.SetCongestionControl(bbr.NewBbrSender(
		bbr.DefaultClock{},
		bbr.GetInitialPacketSize(conn.RemoteAddr()),
	))
}
