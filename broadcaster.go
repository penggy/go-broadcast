package broadcast

import (
	"sync"
)

type broadcaster struct {
	input   chan interface{}
	reg     chan chan<- interface{}
	regLock sync.RWMutex
	unreg   chan chan<- interface{}

	outputs map[chan<- interface{}]bool
}

type Broadcaster interface {
	Register(chan<- interface{})
	Unregister(chan<- interface{})
	Close()
	Submit(interface{})
}

func (b *broadcaster) broadcast(m interface{}) {
	for ch := range b.outputs {
		select {
		case ch <- m:
		default:
		}
	}
}

func (b *broadcaster) run() {
	for {
		select {
		case m := <-b.input:
			b.broadcast(m)
		case ch, ok := <-b.reg:
			if ok {
				b.outputs[ch] = true
			} else {
				return
			}
		case ch := <-b.unreg:
			delete(b.outputs, ch)
		}
	}
}

func NewBroadcaster(buflen int) Broadcaster {
	b := &broadcaster{
		input:   make(chan interface{}, buflen),
		reg:     make(chan chan<- interface{}),
		unreg:   make(chan chan<- interface{}),
		outputs: make(map[chan<- interface{}]bool),
	}

	go b.run()

	return b
}

func (b *broadcaster) Register(newch chan<- interface{}) {
	b.regLock.RLock()
	if b.reg != nil {
		b.reg <- newch
	}
	b.regLock.RUnlock()
}

func (b *broadcaster) Unregister(newch chan<- interface{}) {
	b.regLock.RLock()
	if b.reg != nil {
		b.unreg <- newch
	}
	b.regLock.RUnlock()
}

func (b *broadcaster) Close() {
	b.regLock.Lock()
	if b.reg != nil {
		close(b.reg)
		b.reg = nil
	}
	b.regLock.Unlock()
}

func (b *broadcaster) Submit(m interface{}) {
	b.regLock.RLock()
	if b.reg != nil {
		b.input <- m
	}
	b.regLock.RUnlock()
}
