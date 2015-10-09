package sendclose

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/coredef"
	"github.com/davyxu/cellnet/socket"
	"github.com/davyxu/cellnet/test"
	"github.com/golang/protobuf/proto"
	"log"
	"testing"
)

var signal *test.SignalTester

func runServer() {
	pipe := cellnet.NewEventPipe()

	p := socket.NewAcceptor(pipe).Start("127.0.0.1:7201")

	socket.RegisterSessionMessage(p, coredef.TestEchoACK{}, func(content interface{}, ses cellnet.Session) {
		msg := content.(*coredef.TestEchoACK)

		// 发包后关闭
		ses.Send(&coredef.TestEchoACK{
			Content: proto.String(msg.GetContent()),
		})

		if msg.GetContent() != "noclose" {
			ses.Close()
		}

	})

	pipe.Start()

}

// 客户端连接上后, 主动断开连接, 确保连接正常关闭
func testConnActiveClose() {

	pipe := cellnet.NewEventPipe()

	p := socket.NewConnector(pipe).Start("127.0.0.1:7201")

	socket.RegisterSessionMessage(p, coredef.SessionConnected{}, func(content interface{}, ses cellnet.Session) {

		signal.Done(1)
		// 连接上发包,告诉服务器不要断开
		ses.Send(&coredef.TestEchoACK{
			Content: proto.String("noclose"),
		})

	})

	socket.RegisterSessionMessage(p, coredef.TestEchoACK{}, func(content interface{}, ses cellnet.Session) {
		msg := content.(*coredef.TestEchoACK)

		log.Println("client recv:", msg.String())
		signal.Done(2)

		// 客户端主动断开
		ses.Close()

	})

	socket.RegisterSessionMessage(p, coredef.SessionClosed{}, func(content interface{}, ses cellnet.Session) {

		log.Println("close ok!")
		// 正常断开
		signal.Done(3)

	})

	pipe.Start()

	signal.WaitAndExpect(1, "TestConnActiveClose not connected")
	signal.WaitAndExpect(2, "TestConnActiveClose not recv msg")
	signal.WaitAndExpect(3, "TestConnActiveClose not close")
}

// 接收封包后被断开
func testRecvDisconnected() {

	pipe := cellnet.NewEventPipe()

	p := socket.NewConnector(pipe).Start("127.0.0.1:7201")

	socket.RegisterSessionMessage(p, coredef.SessionConnected{}, func(content interface{}, ses cellnet.Session) {

		// 连接上发包
		ses.Send(&coredef.TestEchoACK{
			Content: proto.String("data"),
		})

		signal.Done(1)
	})

	socket.RegisterSessionMessage(p, coredef.TestEchoACK{}, func(content interface{}, ses cellnet.Session) {
		msg := content.(*coredef.TestEchoACK)

		log.Println("client recv:", msg.String())

		signal.Done(2)

	})

	socket.RegisterSessionMessage(p, coredef.SessionClosed{}, func(content interface{}, ses cellnet.Session) {

		// 断开
		signal.Done(3)
	})

	pipe.Start()

	signal.WaitAndExpect(1, "TestRecvDisconnected not connected")
	signal.WaitAndExpect(2, "TestRecvDisconnected not recv msg")
	signal.WaitAndExpect(3, "TestRecvDisconnected not closed")

}

func TestClose(t *testing.T) {

	signal = test.NewSignalTester(t)

	runServer()

	testConnActiveClose()
	testRecvDisconnected()
}
