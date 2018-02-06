package rpc

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer"
)

type RemoteCallMsg interface {
	GetMsgID() uint16
	GetMsgData() []byte
	GetCallID() int64
}

type EventHooker struct {
}

func (self EventHooker) OnInboundEvent(ev cellnet.Event) {

	rpcMsg, ok := ev.Message().(RemoteCallMsg)
	if !ok {
		return
	}

	msg, meta, err := codec.DecodeMessage(int(rpcMsg.GetMsgID()), rpcMsg.GetMsgData())

	if err != nil {
		return
	}

	peerInfo := ev.BaseSession().Peer().(interface {
		Name() string
	})

	log.Debugf("#rpc recv(%s)@%d %s(%d) | %s",
		peerInfo.Name(),
		ev.BaseSession().(cellnet.Session).ID(),
		meta.TypeName(),
		meta.ID,
		cellnet.MessageToString(msg))

	poster := ev.BaseSession().Peer().(peer.MessagePoster)

	switch ev.Message().(type) {
	case *RemoteCallREQ: // 服务端收到客户端的请求

		poster.PostEvent(&RecvMsgEvent{ev.BaseSession().(cellnet.Session), msg, rpcMsg.GetCallID()})

	case *RemoteCallACK: // 客户端收到服务器的回应
		request := getRequest(rpcMsg.GetCallID())
		if request != nil {
			request.RecvFeedback(msg)
		}
	}

}

func (self EventHooker) OnOutboundEvent(ev cellnet.Event) {
	rpcMsg, ok := ev.Message().(RemoteCallMsg)
	if !ok {
		return
	}

	msg, meta, err := codec.DecodeMessage(int(rpcMsg.GetMsgID()), rpcMsg.GetMsgData())

	if err != nil {
		return
	}

	peerInfo := ev.BaseSession().Peer().(interface {
		Name() string
	})

	log.Debugf("#rpc send(%s)@%d %s(%d) | %s",
		peerInfo.Name(),
		ev.BaseSession().(cellnet.Session).ID(),
		meta.TypeName(),
		meta.ID,
		cellnet.MessageToString(msg))

}

func init() {
	msglog.BlockMessageLog("cellnet.RemoteCallREQ")
	msglog.BlockMessageLog("cellnet.RemoteCallACK")
}
