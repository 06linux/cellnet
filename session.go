package cellnet

type Session interface {

	// 发包
	Send(interface{})

	// 断开
	Close()

	// 标示ID
	ID() int64

	// 归属端
	FromPeer() Peer
}
