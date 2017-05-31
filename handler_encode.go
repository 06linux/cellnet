package cellnet

import "reflect"

type EncodePacketHandler struct {
}

func (self *EncodePacketHandler) Call(ev *SessionEvent) {

	var err error
	ev.Data, ev.MsgID, err = EncodeMessage(ev.Msg)

	ev.SetResult(errToResult(err))
}

var defaultEncodePacketHandler EventHandler = new(EncodePacketHandler)

func StaticEncodePacketHandler() EventHandler {
	return defaultEncodePacketHandler
}

func EncodeMessage(msg interface{}) (data []byte, msgid uint32, err error) {

	meta := MessageMetaByType(reflect.TypeOf(msg))
	if meta != nil {
		msgid = meta.ID
	} else {
		return nil, 0, ErrMessageNotFound
	}

	if meta.Codec == nil {
		return nil, 0, ErrCodecNotFound
	}

	data, err = meta.Codec.Encode(msg)

	return data, msgid, err
}
