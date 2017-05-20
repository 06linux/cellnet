package rpc

import (
	"github.com/davyxu/cellnet"
)

// ud: peer/session,   reqMsg:请求用的消息, userCallback: 返回消息类型回调 func( ackMsg *ackMsgType)
func Call(sesOrPeer interface{}, reqMsg interface{}, ackMsgName string, userCallback func(ev *cellnet.SessionEvent)) error {

	ses, p, err := getPeerSession(sesOrPeer)

	if err != nil {
		return err
	}

	rpcid, err := buildRecvHandler(p, ackMsgName, cellnet.NewCallbackHandler(userCallback))

	if err != nil {
		return err
	}

	// 发送RPC请求
	ev := cellnet.NewSessionEvent(cellnet.SessionEvent_Send, ses)
	ev.TransmitTag = rpcid
	ev.Msg = reqMsg
	ses.RawSend(getSendHandler(), ev)

	return nil
}
