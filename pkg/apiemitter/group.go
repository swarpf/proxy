package apiemitter

import (
	"reflect"
	"sync"
)

// Group marges given subscribed channels into
// on subscribed channel
type Group struct {
	// Cap is capacity to create new channel
	Cap uint

	mu        sync.Mutex
	listeners []listener
	isInit    bool

	stop chan struct{}
	done chan struct{}

	cmu   sync.Mutex
	cases []reflect.SelectCase

	lmu      sync.Mutex
	isListen bool
}

// Flush reset the group to the initial state.
// All references will dropped.
func (g *Group) Flush() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stopIfListen()
	close(g.stop)
	close(g.done)
	g.isInit = false
	g.init()
}

// Add adds channels which were already subscribed to
// some events.
func (g *Group) Add(channels ...<-chan Event) {
	g.mu.Lock()
	defer g.listen()
	defer g.mu.Unlock()
	g.init()

	g.stopIfListen()

	g.cmu.Lock()
	cases := make([]reflect.SelectCase, len(channels))
	for i, ch := range channels {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
	}
	g.cases = append(g.cases, cases...)
	g.cmu.Unlock()
}

// On returns subscribed channel.
func (g *Group) On() <-chan Event {
	g.mu.Lock()
	defer g.listen()
	defer g.mu.Unlock()
	g.init()

	g.stopIfListen()

	l := newListener(g.Cap)
	g.listeners = append(g.listeners, l)
	return l.ch
}

// Off unsubscribed given channels if any or unsubscribed all
// channels in other case
func (g *Group) Off(channels ...<-chan Event) {
	g.mu.Lock()
	defer g.listen()
	defer g.mu.Unlock()
	g.init()

	g.stopIfListen()

	if len(channels) != 0 {
		for _, ch := range channels {
			i := -1
		Listeners:
			for in := range g.listeners {
				if g.listeners[in].ch == ch {
					i = in
					break Listeners
				}
			}
			if i != -1 {
				l := g.listeners[i]
				g.listeners = append(g.listeners[:i], g.listeners[i+1:]...)
				close(l.ch)
			}
		}
	} else {
		g.listeners = make([]listener, 0)
	}
}

func (g *Group) stopIfListen() bool {
	g.lmu.Lock()
	defer g.lmu.Unlock()

	if !g.isListen {
		return false
	}

	g.stop <- struct{}{}
	g.isListen = false
	return true
}

func (g *Group) listen() {
	g.lmu.Lock()
	defer g.lmu.Unlock()
	g.cmu.Lock()
	g.isListen = true

	go func() {
		// unlock cases and isListen flag when func is exit
		defer g.cmu.Unlock()

		for {
			i, val, isOpened := reflect.Select(g.cases)

			// exit if listening is stopped
			if i == 0 {
				return
			}

			if !isOpened && len(g.cases) > i {
				// remove this case
				g.cases = append(g.cases[:i], g.cases[i+1:]...)
			}

			e := val.Interface().(Event)
			// use unblocked mode
			e.Flags = e.Flags | FlagSkip
			// send events to all listeners
			g.mu.Lock()
			for index := range g.listeners {
				l := g.listeners[index]
				// todo(lyrex): we should probably handle errors here
				pushEvent(g.done, l.ch, &e)
			}
			g.mu.Unlock()
		}
	}()
}

func (g *Group) init() {
	if g.isInit {
		return
	}
	g.stop = make(chan struct{})
	g.done = make(chan struct{})
	g.cases = []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(g.stop),
		},
	}
	g.listeners = make([]listener, 0)
	g.isInit = true
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
