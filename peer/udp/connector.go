package udp

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/comm"
	"github.com/davyxu/cellnet/peer"
	"github.com/davyxu/cellnet/proc"
	"net"
)

type udpConnector struct {
	peer.CoreSessionManager
	peer.CorePropertySet
	peer.CoreRunningTag
	proc.CoreDuplexEventProc
	peer.CommunicateConfig
	remoteAddr *net.UDPAddr
	conn       *net.UDPConn
}

func (self *udpConnector) Start() cellnet.Peer {

	var err error
	self.remoteAddr, err = net.ResolveUDPAddr("udp", self.Address())

	if err != nil {

		log.Errorf("#resolve udp address failed(%s) %v", self.NameOrAddress(), err.Error())
		return self
	}

	go self.connect()

	return self
}

func (self *udpConnector) connect() {

	var err error
	self.conn, err = net.DialUDP("udp", nil, self.remoteAddr)
	if err != nil {

		log.Errorf("#connect failed(%s) %v", self.NameOrAddress(), err.Error())
		return
	}

	var running = true

	ses := newUDPSession(nil, self.conn, self, func() {
		running = false
	})

	ses.Start()

	self.CallInboundProc(&cellnet.RecvMsgEvent{ses, &comm.SessionConnected{}})

	buff := make([]byte, 4096)
	for running {

		n, remoteAddr, err := self.conn.ReadFromUDP(buff)
		if err != nil {

			log.Errorln("disconnected:", remoteAddr.String())
			break
		}

		ses.Recv(buff[:n])
	}
}

func (self *udpConnector) Stop() {

	if self.conn != nil {
		self.conn.Close()
	}
}

func (self *udpConnector) TypeName() string {
	return "udp.Connector"
}

func init() {

	peer.RegisterPeerCreator(func() cellnet.Peer {
		p := &udpConnector{}

		return p
	})
}
