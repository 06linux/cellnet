package cellnet

import (
	"runtime/debug"
	"time"
)

type EventQueue interface {
	StartLoop()

	StopLoop(result int)

	// 等待退出
	Wait() int

	// 投递事件, 通过队列到达消费者端
	Post(evd EventDispatcher, data interface{})

	// 延时投递
	DelayPost(evd EventDispatcher, dur time.Duration, data interface{})
}

type queueData struct {
	evd  EventDispatcher
	data interface{}
}

type evQueue struct {
	queue chan queueData

	exitSignal chan int

	capturePanic bool
}

// 派发到队列
func (self *evQueue) Post(evd EventDispatcher, data interface{}) {

	self.queue <- queueData{evd: evd, data: data}
}

func (self *evQueue) DelayPost(evd EventDispatcher, dur time.Duration, data interface{}) {
	go func() {

		time.AfterFunc(dur, func() {

			self.Post(evd, data)
		})

	}()
}

func (self *evQueue) protectedCall(evd EventDispatcher, data interface{}) {

	if self.capturePanic {
		defer func() {

			if err := recover(); err != nil {
				log.Fatalln(err)
				debug.PrintStack()
			}

		}()
	}

	if evd != nil {
		evd.CallData(data)
	} else if f, ok := data.(func()); ok {
		f()
	}

}

func (self *evQueue) StartLoop() {

	go func() {
		for v := range self.queue {
			self.protectedCall(v.evd, v.data)
		}
	}()
}

func (self *evQueue) StopLoop(result int) {
	self.exitSignal <- result
}

func (self *evQueue) Wait() int {
	return <-self.exitSignal
}

func NewEventQueue() EventQueue {
	self := &evQueue{
		queue:      make(chan queueData, 10),
		exitSignal: make(chan int),
	}

	return self

}
