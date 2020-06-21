package apiemitter

import (
	"path"
	"sync"

	"github.com/lyrex/swarpf/pkg/events"
)

// Flag used to describe what behavior you do expect.
type Flag int

//noinspection GoUnusedConst
const (
	// FlagReset only to clear previously defined flags.
	// Example:
	// ee.Use("*", Reset) // clears flags for this pattern
	FlagReset Flag = 0
	// FlagSkip indicates to skip sending if channel is blocked.
	FlagSkip Flag = 1 << iota
)

// New returns just created Emitter struct. Capacity argument
// will be used to create channels with given capacity
func New(capacity uint) *Emitter {
	return &Emitter{
		Cap:       capacity,
		listeners: make(map[string][]listener),
		isInit:    true,
	}
}

// Emitter is a struct that allows to emit, receive
// event, close receiver channel, get info
// about topics and listeners
type Emitter struct {
	Cap       uint
	mu        sync.Mutex
	listeners map[string][]listener
	isInit    bool
}

func newListener(capacity uint) listener {
	return listener{
		ch: make(chan Event, capacity),
	}
}

type listener struct {
	ch chan Event
}

func (e *Emitter) init() {
	if !e.isInit {
		e.listeners = make(map[string][]listener)
		e.isInit = true
	}
}

// On returns a channel that will receive events.
func (e *Emitter) On(topic string) <-chan Event {
	e.mu.Lock()
	e.init()
	l := newListener(e.Cap)
	if listeners, ok := e.listeners[topic]; ok {
		e.listeners[topic] = append(listeners, l)
	} else {
		e.listeners[topic] = []listener{l}
	}
	e.mu.Unlock()
	return l.ch
}

// Off unsubscribes all listeners which were covered by topic, it can be a pattern as well.
func (e *Emitter) Off(topic string, channels ...<-chan Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.init()
	match, _ := e.matched(topic)

	for _, _topic := range match {
		if listeners, ok := e.listeners[_topic]; ok {

			if len(channels) == 0 {
				for i := len(listeners) - 1; i >= 0; i-- {
					close(listeners[i].ch)
					listeners = drop(listeners, i)
				}

			} else {
				for chi := range channels {
					curr := channels[chi]
					for i := len(listeners) - 1; i >= 0; i-- {
						if curr == listeners[i].ch {
							close(listeners[i].ch)
							listeners = drop(listeners, i)
						}
					}
				}
			}
			e.listeners[_topic] = listeners
		}
		if len(e.listeners[_topic]) == 0 {
			delete(e.listeners, _topic)
		}
	}
}

// Listeners returns slice of listeners which were covered by
// topic(it can be pattern) and error if pattern is invalid.
func (e *Emitter) Listeners(topic string) []<-chan Event {
	e.mu.Lock()
	e.init()
	defer e.mu.Unlock()
	var acc []<-chan Event
	match, _ := e.matched(topic)

	for _, _topic := range match {
		list := e.listeners[_topic]
		for i := range e.listeners[_topic] {
			acc = append(acc, list[i].ch)
		}
	}

	return acc
}

// Topics returns all existing topics.
func (e *Emitter) Topics() []string {
	e.mu.Lock()
	e.init()
	defer e.mu.Unlock()
	acc := make([]string, len(e.listeners))
	i := 0
	for k := range e.listeners {
		acc[i] = k
		i++
	}
	return acc
}

// Emit emits an event with the rest arguments to all
// listeners which were covered by topic(it can be pattern).
func (e *Emitter) Emit(topic string, apiEvent events.ApiEventMsg) chan struct{} {
	e.mu.Lock()
	e.init()
	done := make(chan struct{}, 1)

	match, _ := e.matched(topic)

	var wg sync.WaitGroup
	var haveToWait bool
	for _, _topic := range match {
		listeners := e.listeners[_topic]
		event := Event{
			Topic:         _topic,
			OriginalTopic: topic,
			ApiEvent:      apiEvent,
		}

		for i := len(listeners) - 1; i >= 0; i-- {
			lstnr := listeners[i]
			evn := *(&event) // copy the event

			wg.Add(1)
			haveToWait = true
			go func(lstnr listener, event *Event) {
				e.mu.Lock()
				_, remove := pushEvent(done, lstnr.ch, event)
				if remove {
					defer e.Off(event.Topic, lstnr.ch)
				}
				wg.Done()
				e.mu.Unlock()
			}(lstnr, &evn)

		}

	}
	if haveToWait {
		go func(done chan struct{}) {
			defer func() { recover() }()
			wg.Wait()
			close(done)
		}(done)
	} else {
		close(done)
	}

	e.mu.Unlock()
	return done
}

func pushEvent(
	done chan struct{},
	lstnr chan Event,
	event *Event,
) (success, remove bool) {
	sent, canceled := send(
		done,
		lstnr,
		*event,
		true,
	)
	success = sent

	// todo(lyrex): this looks sketchy. make this nicer
	if !sent && !canceled {
		remove = false
		// if not sent
	} else if !canceled {
		// if event was sent successfully
		remove = false
	}
	return
}

func (e *Emitter) matched(topic string) ([]string, error) {
	var acc []string
	var err error
	for k := range e.listeners {
		if matched, err := path.Match(topic, k); err != nil {
			return []string{}, err
		} else if matched {
			acc = append(acc, k)
		} else {
			if matched, _ := path.Match(k, topic); matched {
				acc = append(acc, k)
			}
		}
	}
	return acc, err
}

func drop(l []listener, i int) []listener {
	return append(l[:i], l[i+1:]...)
}

func send(done chan struct{}, ch chan Event, e Event, wait bool, ) (sent, canceled bool) {

	defer func() {
		if r := recover(); r != nil {
			canceled = false
			sent = false
		}
	}()

	if !wait {
		select {
		case <-done:
			break
		case ch <- e:
			sent = true
			return
		default:
			return
		}

	} else {
		select {
		case <-done:
			break
		case ch <- e:
			sent = true
			return
		}

	}
	canceled = true
	return
}

// Channel based event emitter for Golang
// Copyright (C) 2015 Oleg Lebedev
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
