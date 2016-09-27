package socket

import (
	"reflect"

	"github.com/davyxu/cellnet"
)

// 注册连接消息
func RegisterSessionMessage(eq cellnet.EventQueue, msgName string, userHandler func(interface{}, cellnet.Session)) *cellnet.MessageMeta {

	msgMeta := cellnet.MessageMetaByName(msgName)

	if msgMeta == nil {
		log.Errorf("message register failed, %s", msgName)
		return nil
	}

	eq.RegisterCallback(msgMeta.ID, func(data interface{}) {

		if ev, ok := data.(*SessionEvent); ok {

			rawMsg, err := cellnet.ParsePacket(ev.Packet, msgMeta.Type)

			if err != nil {
				log.Errorln("unmarshaling error:\n", err)
				return
			}

			userHandler(rawMsg, ev.Ses)

		}

	})

	return msgMeta
}

// 注册连接消息
func RegisterPeerMessage(eq cellnet.EventQueue, msgName string, userHandler func(interface{}, cellnet.Peer)) *cellnet.MessageMeta {

	msgMeta := cellnet.MessageMetaByName(msgName)

	if msgMeta == nil {
		log.Errorf("message register failed, %s", msgName)
		return nil
	}

	eq.RegisterCallback(msgMeta.ID, func(data interface{}) {

		if ev, ok := data.(*PeerEvent); ok {

			rawMsg := reflect.New(msgMeta.Type).Interface()

			userHandler(rawMsg, ev.P)

		}

	})

	return msgMeta
}
