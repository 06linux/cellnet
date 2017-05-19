package cellnet

import "reflect"

type EncodePacketHandler struct {
	BaseEventHandler
}

func (self *EncodePacketHandler) Call(ev *SessionEvent) {

	var err error
	ev.Data, ev.MsgID, err = EncodeMessage(ev.Msg)

	if err != nil {
		log.Debugln(err, ev.String())
	}

}

func EncodeMessage(msg interface{}) (data []byte, msgid uint32, err error) {

	fullName := MessageFullName(reflect.TypeOf(msg))

	meta := MessageMetaByName(fullName)
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

func NewEncodePacketHandler() EventHandler {
	return &EncodePacketHandler{}
}
