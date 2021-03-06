package opay

import (
	"log"
	"runtime"
	"sync"
	"time"
)

type (
	// 订单队列
	Queue interface {
		SetCap(int)
		Push(Request) (respChan <-chan *Response)
		Pull() Request
		GetOpay() *Opay
	}

	OrderChan struct {
		c    chan Request
		mu   sync.RWMutex
		opay *Opay
	}
)

const (
	DEFAULT_QUEUE_CAP = 1024 //队列默认容量
)

func newOrderChan(queueCapacity int, opay *Opay) Queue {
	if queueCapacity <= 0 {
		queueCapacity = DEFAULT_QUEUE_CAP
	}
	return &OrderChan{
		c:    make(chan Request, queueCapacity),
		opay: opay,
	}
}

// 设置队列容量
func (oc *OrderChan) SetCap(queueCapacity int) {
	if queueCapacity <= 0 {
		queueCapacity = DEFAULT_QUEUE_CAP
	}
	close(oc.c)
	if len(oc.c) > 0 {
		log.Println("Waiting for the completion of the remaining order processing...")
		for len(oc.c) > 0 {
			runtime.Gosched()
		}
	}
	oc.mu.Lock()
	oc.c = make(chan Request, queueCapacity)
	oc.mu.Unlock()

	log.Println("Successfully set the queue capacity.")
}

// 推送一条订单
func (oc *OrderChan) Push(req Request) (respChan <-chan *Response) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	respChan, err := req.prepare(oc.GetOpay())
	if err != nil {
		req.setError(err)
		req.writeback()
		return
	}

	timeout, err := checkTimeout(req.Deadline)

	if err != nil {
		// 已超时，取消处理
		req.setError(err)
		req.writeback()
		return
	}

	if timeout > 0 {
		// 未超时
		select {
		case oc.c <- req:
		case <-time.After(timeout):
			err = ErrTimeout
			req.setError(err)
			req.writeback()
		}

	} else {
		// 无超时限制
		oc.c <- req
	}

	return
}

// 读出一条订单
// 无限等待，直到取出一个有效订单
// 超时订单，自动处理
func (oc *OrderChan) Pull() Request {
	var (
		req Request
		c   chan Request
	)

	for {
		oc.mu.RLock()
		c = oc.c
		oc.mu.RUnlock()

		req = <-c
		if req.isNil() {
			continue
		}

		// If timeout, cancel the order.
		if _, err := checkTimeout(req.Deadline); err != nil {
			req.setError(err)
			req.writeback()
			continue
		}
		break
	}

	return req
}

func (oc *OrderChan) GetOpay() *Opay {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return oc.opay
}
