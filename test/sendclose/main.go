package main

import (
	"flag"
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/coredef"
	"github.com/davyxu/cellnet/socket"
	"github.com/golang/protobuf/proto"
	"log"
	"runtime"
	"strconv"
	"sync"
)

var done = make(chan bool)

// 测试客户端连接数量
const connCount = 10

func runClient() {

	pipe := cellnet.NewEvPipe()

	// 同步量
	var endAcc sync.WaitGroup

	// 启动N个连接
	for i := 0; i < connCount; i++ {

		endAcc.Add(1)

		p := socket.NewConnector(pipe).Start("127.0.0.1:7235")

		p.SetName(fmt.Sprintf("%d", i))

		socket.RegisterSessionMessage(p, coredef.TestEchoACK{}, func(ses cellnet.Session, content interface{}) {
			msg := content.(*coredef.TestEchoACK)

			log.Println("client recv:", msg.String())

			// 正常收到
			endAcc.Done()
		})

		socket.RegisterSessionMessage(p, coredef.SessionConnected{}, func(ses cellnet.Session, content interface{}) {

			id, _ := strconv.Atoi(ses.FromPeer().Name())

			// 连接上发包
			ses.Send(&coredef.TestEchoACK{
				Content: proto.String(fmt.Sprintf("data#%d", id)),
			})

		})

	}

	pipe.Start()

	log.Println("waiting server msg...")

	// 等待完成
	endAcc.Wait()

}

func runServer() {
	pipe := cellnet.NewEvPipe()

	p := socket.NewAcceptor(pipe).Start("127.0.0.1:7235")

	// 计数器, 应该按照connCount倍数递增
	var counter int

	socket.RegisterSessionMessage(p, coredef.TestEchoACK{}, func(ses cellnet.Session, content interface{}) {
		msg := content.(*coredef.TestEchoACK)

		if p.Get(ses.ID()) != ses {
			panic("1: session not exist in SessionManager")
		}

		counter++
		log.Printf("No. %d: server recv: %v", counter, msg.String())

		// 发包后关闭
		ses.Send(&coredef.TestEchoACK{
			Content: proto.String(msg.GetContent()),
		})

		if p.Get(ses.ID()) != ses {
			panic("2: session not exist in SessionManager")
		}

		ses.Close()

		if p.Get(ses.ID()) != ses {
			panic("3: session not exist in SessionManager")
		}

	})

	pipe.Start()

	done <- true

}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	mode := flag.String("mode", "", "specify the mode of this test")

	flag.Parse()

	if mode != nil && *mode == "client" {
		runClient()
	} else {
		runServer()
	}

}
