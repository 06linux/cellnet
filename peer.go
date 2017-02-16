package cellnet

type Peer interface {

	// 开启/关闭
	Start(address string) Peer
	Stop()

	// 名字
	SetName(string)
	Name() string

	// 地址
	Address() string

	// Session最大包大小, 超过这个数字, 接收视为错误, 断开连接
	SetMaxPacketSize(size int)
	MaxPacketSize() int

	// 派发器
	EventDispatcher

	// 连接管理
	SessionManager

	//  Handler
	SetHandler(recv, send EventHandler)
	GetHandler() (recv, send EventHandler)

	// Codec
	PacketCodec() Codec
	SetPacketCodec(c Codec)
}

type Connector interface {

	// 连接后的Session
	DefaultSession() Session

	// 自动重连间隔, 0表示不重连, 默认不重连
	SetAutoReconnectSec(sec int)
}

type SessionManager interface {

	// 获取一个连接
	GetSession(int64) Session

	// 遍历连接
	VisitSession(func(Session) bool)

	// 连接数量
	SessionCount() int
}
