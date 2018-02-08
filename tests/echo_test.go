package tests

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	_ "github.com/davyxu/cellnet/peer/udp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/kcp"
	_ "github.com/davyxu/cellnet/proc/tcp"
	_ "github.com/davyxu/cellnet/proc/udp"
	"github.com/davyxu/cellnet/util"
	"testing"
	"time"
)

type echoContext struct {
	Address   string
	Protocol  string
	Processor string
	Tester    *util.SignalTester
	Acceptor  cellnet.Peer
}

var (
	echoContexts = []*echoContext{
		{
			Address:   "127.0.0.1:7701",
			Protocol:  "tcp",
			Processor: "tcp.ltv",
		},
		{
			Address:   "127.0.0.1:7702",
			Protocol:  "udp",
			Processor: "udp.ltv",
		},
		{
			Address:   "127.0.0.1:7703",
			Protocol:  "udp",
			Processor: "udp.kcp.ltv",
		},
	}
)

func echo_StartServer(context *echoContext) {
	queue := cellnet.NewEventQueue()

	context.Acceptor = peer.NewPeer(context.Protocol + ".Acceptor")
	pset := context.Acceptor.(cellnet.PropertySet)
	pset.SetProperty("Address", context.Address)
	pset.SetProperty("Name", "server")
	pset.SetProperty("Queue", queue)

	proc.BindProcessor(context.Acceptor, context.Processor, func(ev cellnet.Event) {

		switch msg := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			fmt.Println("server accepted")
		case *TestEchoACK:

			fmt.Printf("server recv %+v\n", msg)

			ev.Session().Send(&TestEchoACK{
				Msg:   msg.Msg,
				Value: msg.Value,
			})

		case *cellnet.SessionClosed:
			fmt.Println("session closed: ", ev.Session().ID())
		}

	})

	context.Acceptor.Start()

	queue.StartLoop()
}

func echo_StartClient(echoContext *echoContext) {
	queue := cellnet.NewEventQueue()

	p := peer.NewPeer(echoContext.Protocol + ".Connector")
	pset := p.(cellnet.PropertySet)
	pset.SetProperty("Address", echoContext.Address)
	pset.SetProperty("Name", "client")
	pset.SetProperty("Queue", queue)

	proc.BindProcessor(p, echoContext.Processor, func(ev cellnet.Event) {

		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			fmt.Println("client connected")
			ev.Session().Send(&TestEchoACK{
				Msg:   "hello",
				Value: 1234,
			})
		case *TestEchoACK:

			fmt.Printf("client recv %+v\n", msg)

			echoContext.Tester.Done(1)

		case *cellnet.SessionClosed:
			fmt.Println("client error: ")
		}
	})

	p.Start()

	queue.StartLoop()

	echoContext.Tester.WaitAndExpect("not recv data", 1)
}

func runEcho(t *testing.T, index int) {

	ctx := echoContexts[index]

	ctx.Tester = util.NewSignalTester(t)
	ctx.Tester.SetTimeout(time.Hour)

	echo_StartServer(ctx)

	echo_StartClient(ctx)

	ctx.Acceptor.Stop()
}

func TestEchoTCP(t *testing.T) {

	runEcho(t, 0)
}

func TestEchoUDP(t *testing.T) {

	runEcho(t, 1)
}

func TestEchoKCP(t *testing.T) {

	runEcho(t, 2)
}
