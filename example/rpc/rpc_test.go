package rpc

import (
	"testing"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/example"
	"github.com/davyxu/cellnet/proto/pb/gamedef"
	"github.com/davyxu/cellnet/rpc"
	"github.com/davyxu/cellnet/socket"
	"github.com/davyxu/golog"
)

var log *golog.Logger = golog.New("test")

var asyncSignal *test.SignalTester
var syncSignal *test.SignalTester

func server() {

	queue := cellnet.NewEventQueue()

	p := socket.NewAcceptor(queue)
	p.SetName("server")
	p.Start("127.0.0.1:9201")

	rpc.RegisterMessage(p, "gamedef.TestEchoACK", func(ev *cellnet.SessionEvent) {
		msg := ev.Msg.(*gamedef.TestEchoACK)

		log.Debugln("server recv:", msg.String())

		ev.Send(&gamedef.TestEchoACK{
			Content: msg.String(),
		})

	})

	queue.StartLoop()

}

// 异步阻塞调用的rpc: 适用于逻辑服与逻辑服之间互相查询数据
func asyncClient() {

	queue := cellnet.NewEventQueue()

	p := socket.NewConnector(queue)
	p.SetName("client.async")
	p.Start("127.0.0.1:9201")

	cellnet.RegisterMessage(p, "coredef.SessionConnected", func(ev *cellnet.SessionEvent) {

		for i := 0; i < 2; i++ {

			copy := i + 1

			rpc.Call(p, &gamedef.TestEchoACK{
				Content: "async",
			}, "gamedef.TestEchoACK", func(rpcev *cellnet.SessionEvent) {
				msg := rpcev.Msg.(*gamedef.TestEchoACK)

				log.Debugln(copy, "client async recv:", msg.Content)

				asyncSignal.Done(copy)
			})

		}

	})

	queue.StartLoop()

	asyncSignal.WaitAndExpect("async not recv data ", 1, 2)
}

// 同步阻塞调用的rpc: 适用于web后台向逻辑服查询数据后生成页面
func syncClient() {

	queue := cellnet.NewEventQueue()

	p := socket.NewConnector(queue)
	p.SetName("client.sync")
	p.Start("127.0.0.1:9201")

	cellnet.RegisterMessage(p, "coredef.SessionConnected", func(ev *cellnet.SessionEvent) {

		for i := 0; i < 2; i++ {

			// 这里使用goroutine包裹调用原因: 避免当前消息不返回, 无法继续处理rpc的消息接收
			// 正式使用时, CallSync被调用的消息所在的Peer, 与CallSync第一个参数使用Peer一定是不同Peer
			go func(id int) {

				result, err := rpc.CallSync(p, &gamedef.TestEchoACK{
					Content: "sync",
				}, "gamedef.TestEchoACK", 5)

				if err != nil {
					syncSignal.Log(err)
					return
				}

				msg := result.(*gamedef.TestEchoACK)
				log.Debugln("client sync recv:", msg.Content, id*100)

				syncSignal.Done(id * 100)

			}(i + 1)

		}

	})

	queue.StartLoop()

	syncSignal.WaitAndExpect("sync not recv data ", 100, 200)

}

func TestRPC(t *testing.T) {

	asyncSignal = test.NewSignalTester(t)
	syncSignal = test.NewSignalTester(t)

	server()

	asyncClient()

	syncClient()

}
