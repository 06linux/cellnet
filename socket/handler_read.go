package socket

import (
	"github.com/davyxu/cellnet"
	"time"
)

type ReadPacketHandler struct {
}

func (self *ReadPacketHandler) Call(ev *cellnet.Event) {

	switch ev.Type {
	case cellnet.Event_Recv:

		rawSes := ev.Ses.(*socketSession)

		// 读超时
		read, _ := rawSes.FromPeer().(SocketOptions).SocketDeadline()

		if read != 0 {
			rawSes.stream.Raw().SetReadDeadline(time.Now().Add(read))
		}

		msgid, data, err := rawSes.stream.Read()

		if err != nil {

			ev.SetResult(errToResult(err))

			// 外部会根据Result断开连接并抛出错误

		} else {

			ev.MsgID = msgid
			// 逻辑封包
			ev.Data = data
		}

	}

}

var defaultReadPacketHandler = new(ReadPacketHandler)

func StaticReadPacketHandler() cellnet.EventHandler {
	return defaultReadPacketHandler
}
