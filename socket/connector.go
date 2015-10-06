package socket

import (
	"github.com/davyxu/cellnet"
	"log"
	"net"
	"time"
)

type socketConnector struct {
	*peerProfile
	*sessionMgr

	conn net.Conn

	autoReconnectSec int // 重连间隔时间, 0为不重连

	closeSignal chan bool

	working bool // 重入锁

	defaultSes cellnet.Session
}

func (self *socketConnector) Start(address string) cellnet.Peer {

	if self.working {
		return self
	}

	go self.connect(address)

	self.PostData(NewPeerEvent(Event_PeerStart, self))

	return self
}

func (self *socketConnector) connect(address string) {
	self.working = true

	for {

		// 开始连接
		cn, err := net.Dial("tcp", address)

		// 连不上
		if err != nil {

			log.Println("[socket] cononect failed", err.Error())

			// 没重连就退出
			if self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)

			// 继续连接
			continue
		}

		log.Println("connected: ", address)

		// 连上了, 记录连接
		self.conn = cn

		// 创建Session
		ses := newSession(NewPacketStream(cn), self.EventQueue, self)
		self.sessionMgr.Add(ses)
		self.defaultSes = ses

		// 内部断开回调
		ses.OnClose = func() {
			self.sessionMgr.Remove(ses)
			self.closeSignal <- true
		}

		// 抛出事件
		self.PostData(NewSessionEvent(Event_SessionConnected, ses, nil))

		if <-self.closeSignal {

			self.conn = nil

			// 没重连就退出
			if self.autoReconnectSec == 0 {
				break
			}

			// 有重连就等待
			time.Sleep(time.Duration(self.autoReconnectSec) * time.Second)

			// 继续连接
			continue

		}

	}

	self.working = false
}

func (self *socketConnector) Stop() {

	if self.conn != nil {

		self.PostData(NewPeerEvent(Event_PeerStop, self))

		self.conn.Close()
	}

}

func (self *socketConnector) DefaultSession() cellnet.Session {
	return self.defaultSes
}

func NewConnector(pipe cellnet.EventPipe) cellnet.Peer {
	self := &socketConnector{
		sessionMgr:  newSessionManager(),
		peerProfile: newPeerProfile(pipe.AddQueue()),
		closeSignal: make(chan bool),
	}

	self.PostData(NewPeerEvent(Event_PeerInit, self))

	return self
}
