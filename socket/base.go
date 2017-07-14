package socket

import (
	"github.com/davyxu/cellnet"
	"net"
)

// Peer间的共享数据
type socketPeer struct {
	cellnet.EventQueue
	// 会话管理器
	cellnet.SessionManager

	// 共享配置
	*cellnet.BasePeerImplement

	// 自带派发器
	*cellnet.DispatcherHandler

	// socket配置
	*socketOptions

	// 停止过程同步
	stopping chan bool
	// 自定义流
	streamGen func(net.Conn) cellnet.PacketStream
}

func (self *socketPeer) waitStopFinished() {
	// 如果正在停止时, 等待停止完成
	if self.stopping != nil {
		<-self.stopping
		self.stopping = nil
	}
}

func (self *socketPeer) isStopping() bool {
	return self.stopping != nil
}

func (self *socketPeer) startStopping() {
	self.stopping = make(chan bool)
}

func (self *socketPeer) endStopping() {
	select {
	case self.stopping <- true:

	default:
		self.stopping = nil
	}
}
func (self *socketPeer) SetPacketStreamGenerator(callback func(net.Conn) cellnet.PacketStream) {

	self.streamGen = callback
}

func (self *socketPeer) genPacketStream(conn net.Conn) cellnet.PacketStream {

	self.socketOptions.apply(conn)

	if self.streamGen == nil {
		return NewTLVStream(conn)
	}

	return self.streamGen(conn)
}

func (self *socketPeer) Queue() cellnet.EventQueue {
	return self.EventQueue
}

func newSocketPeer(queue cellnet.EventQueue, sm cellnet.SessionManager) *socketPeer {

	self := &socketPeer{
		EventQueue:        queue,
		DispatcherHandler: cellnet.NewDispatcherHandler(),
		SessionManager:    sm,
		socketOptions:     newSocketOptions(),
		BasePeerImplement: cellnet.NewBasePeer(),
	}

	self.BasePeerImplement.SetHandlerList(BuildRecvHandler(self.DispatcherHandler), BuildSendHandler())

	return self
}
