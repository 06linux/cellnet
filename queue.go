package cellnet

import (
	"runtime/debug"
	"sync"
)

type EventQueue interface {
	StartLoop()

	StopLoop()

	// 等待退出
	Wait()

	// 投递事件, 通过队列到达消费者端
	Post(callback func())

	// 是否捕获异常
	EnableCapturePanic(v bool)
}

type eventQueue struct {
	queue chan func()

	endSignal sync.WaitGroup

	capturePanic bool
}

// 启动崩溃捕获
func (self *eventQueue) EnableCapturePanic(v bool) {
	self.capturePanic = v
}

// 派发事件处理回调到队列中
func (self *eventQueue) Post(callback func()) {

	if callback == nil {
		return
	}

	self.queue <- callback
}

// 保护调用用户函数
func (self *eventQueue) protectedCall(callback func()) {

	if self.capturePanic {
		defer func() {

			if err := recover(); err != nil {

				debug.PrintStack()
			}

		}()
	}

	callback()
}

// 开启事件循环
func (self *eventQueue) StartLoop() {

	go func() {

		for callback := range self.queue {

			if callback == nil {
				break
			}

			self.protectedCall(callback)
		}

		self.endSignal.Done()
	}()
}

// 停止事件循环
func (self *eventQueue) StopLoop() {
	self.queue <- nil
}

// 等待退出消息
func (self *eventQueue) Wait() {
	self.endSignal.Wait()
}

const DefaultQueueSize = 100

// 创建默认长度的队列
func NewEventQueue() EventQueue {

	return &eventQueue{
		queue: make(chan func(), DefaultQueueSize),
	}
}

func QueuedCall(ses Session, callback func()) {
	if ses == nil {
		return
	}

	q := ses.Peer().EventQueue()

	// Peer有队列时，在队列线程调用用户处理函数
	if q != nil {
		q.Post(callback)

	} else {

		// 在I/O线程调用用户处理函数
		callback()
	}
}
