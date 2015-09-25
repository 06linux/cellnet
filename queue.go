/*
	dispatcher包提供消息队列, 消息注册+派发
	封装消息解包, 打包的过程

*/
package cellnet

type EvQueue struct {
	// 保证注册发生在初始化, 读取发生在之后可以不用锁
	contextMap map[int][]func(interface{})

	queue chan interface{}

	inject func(interface{}) bool
}

// 注册事件回调
func (self *EvQueue) RegisterCallback(id int, f func(interface{})) {

	// 事件
	em, ok := self.contextMap[id]

	// 新建
	if !ok {

		em = make([]func(interface{}), 0)

	}

	em = append(em, f)

	self.contextMap[id] = em
}

// 注入回调, 返回false时表示不再投递
func (self *EvQueue) Inject(f func(interface{}) bool) {

	self.inject = f
}

func (self *EvQueue) Exists(id int) bool {
	_, ok := self.contextMap[id]

	return ok
}

// 派发到队列
func (self *EvQueue) Post(data interface{}) {
	self.queue <- data
}

func (self *EvQueue) Count() int {
	return len(self.contextMap)
}

func (self *EvQueue) CountByID(id int) int {
	if v, ok := self.contextMap[id]; ok {
		return len(v)
	}

	return 0
}

type contentIndexer interface {
	ContextID() int
}

// 通过数据接口调用
func (self *EvQueue) call(data interface{}) {

	// 先处理注入
	if self.inject != nil && !self.inject(data) {
		return
	}

	// 再投递消息
	if ci, ok := data.(contentIndexer); ok {

		if carr, ok := self.contextMap[ci.ContextID()]; ok {

			for _, c := range carr {

				c(data)
			}
		}
	}

}

const queueLength = 10

func newEvQueue() *EvQueue {
	self := &EvQueue{
		contextMap: make(map[int][]func(interface{})),
		queue:      make(chan interface{}, queueLength),
	}

	return self

}
