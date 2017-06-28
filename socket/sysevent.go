package socket

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/binary/coredef"
)

var (
	Meta_SessionConnected = cellnet.MessageMetaByName("coredef.SessionConnected")
	Meta_SessionAccepted  = cellnet.MessageMetaByName("coredef.SessionAccepted")
)

func systemEvent(ses cellnet.Session, e cellnet.EventType, hlist []cellnet.EventHandler) {

	ev := cellnet.NewEvent(e, ses)

	var meta *cellnet.MessageMeta
	switch e {
	case cellnet.Event_Accepted:
		meta = Meta_SessionAccepted
	case cellnet.Event_Connected:
		meta = Meta_SessionConnected
	}

	ev.FromMeta(meta)

	cellnet.HandlerChainCall(hlist, ev)
}

func systemError(ses cellnet.Session, e cellnet.EventType, r cellnet.Result, hlist []cellnet.EventHandler) {

	ev := cellnet.NewEvent(e, ses)

	// 直接放在这里, decoder里遇到系统事件不会进行decode操作
	switch e {
	case cellnet.Event_Closed:
		ev.Msg = &coredef.SessionClosed{Result: r}
	case cellnet.Event_AcceptFailed:
		ev.Msg = &coredef.SessionAcceptFailed{Result: r}
	case cellnet.Event_ConnectFailed:
		ev.Msg = &coredef.SessionConnectFailed{Result: r}
	default:
		panic("unknown system error")
	}

	var encodeErr error
	ev.Data, ev.MsgID, encodeErr = cellnet.EncodeMessage(ev.Msg)

	if encodeErr != nil {
		panic("system error encode error: " + encodeErr.Error())
	}

	ev.Type = e

	cellnet.HandlerChainCall(hlist, ev)
}
