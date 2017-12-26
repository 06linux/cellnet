package internal

import (
	"github.com/davyxu/cellnet"
	"sync"
)

// 通讯端共享的数据
type PeerShare struct {
	cellnet.PeerConfig
	SessionManager

	// 单独保存的保存cellnet.Peer接口
	peerInterface cellnet.Peer
	tag           interface{}
	// 运行状态
	running      bool
	runningGuard sync.RWMutex

	// 停止过程同步
	stopping chan bool

	InboundEvent  cellnet.EventFunc
	OutboundEvent cellnet.EventFunc
}

func (self *PeerShare) SetEventFunc(inboundEvent, outboundEvent cellnet.EventFunc) {
	self.InboundEvent = inboundEvent
	self.OutboundEvent = outboundEvent
}

func (self *PeerShare) SetConfig(config cellnet.PeerConfig) {
	self.PeerConfig = config
}

func (self *PeerShare) IsConnector() bool {
	return false
}

func (self *PeerShare) IsAcceptor() bool {
	return false
}

func (self *PeerShare) IsRunning() bool {

	self.runningGuard.RLock()
	defer self.runningGuard.RUnlock()

	return self.running
}

func (self *PeerShare) Tag() interface{} {
	return self.tag
}

func (self *PeerShare) SetTag(tag interface{}) {
	self.tag = tag
}

func (self *PeerShare) SetRunning(v bool) {
	self.runningGuard.Lock()
	self.running = v
	self.runningGuard.Unlock()
}

// socket包内部派发事件
func (self *PeerShare) InvokeInboundEvent(ev interface{}) interface{} {

	if self.InboundEvent == nil {
		return nil
	}

	return self.InboundEvent(ev)
}

// socket包内部派发事件
func (self *PeerShare) InvokeOutboundEvent(ev interface{}) interface{} {

	if self.OutboundEvent == nil {
		return nil
	}

	return self.OutboundEvent(ev)
}

func (self *PeerShare) NameOrAddress() string {
	if self.PeerName != "" {
		return self.PeerName
	}

	return self.PeerAddress
}

func (self *PeerShare) Peer() cellnet.Peer {
	return self.peerInterface
}

func (self *PeerShare) Init(p cellnet.Peer) {
	self.SessionManager = NewSessionManager()
	self.peerInterface = p
}

func (self *PeerShare) WaitStopFinished() {
	// 如果正在停止时, 等待停止完成
	if self.stopping != nil {
		<-self.stopping
		self.stopping = nil
	}
}

func (self *PeerShare) IsStopping() bool {
	return self.stopping != nil
}

func (self *PeerShare) StartStopping() {
	self.stopping = make(chan bool)
}

func (self *PeerShare) EndStopping() {
	select {
	case self.stopping <- true:

	default:
		self.stopping = nil
	}
}
