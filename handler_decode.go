package cellnet

import (
	"errors"
	"reflect"
)

type DecodePacketHandler struct {
	BaseEventHandler
}

func (self *DecodePacketHandler) Call(ev *SessionEvent) {

	// 系统消息不做处理
	if !ev.IsSystemEvent() {

		var err error
		ev.Msg, err = DecodeMessage(ev.MsgID, ev.Data)

		if err != nil {
			log.Errorln(err)
		}
	}

}

var ErrMessageNotFound = errors.New("message not found")
var ErrCodecNotFound = errors.New("codec not found")

func DecodeMessage(msgid uint32, data []byte) (interface{}, error) {
	meta := MessageMetaByID(msgid)

	if meta == nil {
		return nil, ErrMessageNotFound
	}
	if meta.Codec == nil {
		return nil, ErrCodecNotFound
	}

	// 创建消息
	msg := reflect.New(meta.Type).Interface()

	// 解析消息
	err := meta.Codec.Decode(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func NewDecodePacketHandler() EventHandler {

	return &DecodePacketHandler{}
}
