package vmess

import (
	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

type Account struct {
	ID     *protocol.ID
	Server *Server
}

type Server struct {
	Target      string
	TCPFastOpen bool
}
