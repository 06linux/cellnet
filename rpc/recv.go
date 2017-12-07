package rpc

import (
	"github.com/davyxu/cellnet"
)

func ProcRPC(userFunc cellnet.EventFunc) cellnet.EventFunc {

	return func(raw cellnet.EventParam) cellnet.EventResult {

		recvEv, ok := raw.(cellnet.RecvMsgEvent)

		if ok {
			switch rpcMsg := recvEv.Msg.(type) {

			case *RemoteCallREQ: // 服务器端收到

				if msg, err := cellnet.DecodeMessage(rpcMsg.MsgID, rpcMsg.Data); err == nil {

					return userFunc(RecvMsgEvent{recvEv.Ses, msg, rpcMsg.CallID})

				} else {
					return err
				}
			case *RemoteCallACK: // 客户端收到

				if msg, err := cellnet.DecodeMessage(rpcMsg.MsgID, rpcMsg.Data); err == nil {

					request := getRequest(rpcMsg.CallID)
					if request != nil {
						request.RecvFeedback(msg)
					}

				} else {
					return err
				}
			}
		}

		return userFunc(raw)
	}
}
