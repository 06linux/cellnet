package main

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/ltvsocket"
	"github.com/davyxu/cellnet/proto/coredef"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

type IError interface {
	Error() string
}

var done = make(chan bool)

func server() {
	ltvsocket.SpawnAcceptor("127.0.0.1:8001", func(cm interface{}) {

		switch v := cm.(type) {
		case ltvsocket.SocketCreateSession:

			ltvsocket.SpawnSession(v.Stream, v.Type, func(sescm interface{}) {

				switch ev := sescm.(type) {
				case ltvsocket.SocketData:

					pkt := ev.GetPacket()

					var ack coredef.TestEchoACK
					if err := proto.Unmarshal(pkt.Data, &ack); err != nil {
						log.Println(err)
					} else {
						log.Println("server recv:", ack.String())

					}

					cellnet.Send(ev.Session, cellnet.BuildPacket(&coredef.TestEchoACK{
						Content: proto.String("world"),
					}))

				}

			})

		case IError:
			log.Println(cellnet.ReflectContent(v))
		}

	})
}

func client() {

	ltvsocket.SpawnConnector("127.0.0.1:8001", func(cm interface{}) {

		switch v := cm.(type) {
		case ltvsocket.SocketCreateSession:

			// new session
			ltvsocket.SpawnSession(v.Stream, v.Type, func(sescm interface{}) {

				switch ev := sescm.(type) {
				case ltvsocket.SocketNewSession:

					cellnet.Send(ev.Session, cellnet.BuildPacket(&coredef.TestEchoACK{
						Content: proto.String("hello"),
					}))

				case ltvsocket.SocketData:

					pkt := ev.GetPacket()

					var ack coredef.TestEchoACK
					if err := proto.Unmarshal(pkt.Data, &ack); err != nil {
						log.Println(err)
					} else {
						log.Println("client recv:", ack.String())

						done <- true
					}

				}

			})

		case IError:
			log.Println(cellnet.ReflectContent(v))

		}

	})

}

func main() {

	server()

	client()

	select {
	case <-done:

	case <-time.After(2 * time.Second):
		log.Println("time out")
	}

}
