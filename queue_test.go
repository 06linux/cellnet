package cellnet

import (
	"testing"
)

func TestQueue(t *testing.T) {

	q := NewEvQueue()

	q.RegisterCallback(1, func(d interface{}) {

		t.Log(d)
	})

	q.RegisterCallback(1, func(d interface{}) {

		t.Log(d)
	})

	q.Post(Packet{MsgID: 1})

}
