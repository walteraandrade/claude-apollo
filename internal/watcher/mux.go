package watcher

import "sync"

type Mux struct {
	out  chan Event
	done chan struct{}
	wg   sync.WaitGroup
}

func NewMux() *Mux {
	return &Mux{
		out:  make(chan Event, 4),
		done: make(chan struct{}),
	}
}

func (m *Mux) Add(ch <-chan Event) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.done:
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				select {
				case m.out <- ev:
				case <-m.done:
					return
				}
			}
		}
	}()
}

func (m *Mux) Events() <-chan Event {
	return m.out
}

func (m *Mux) Close() {
	close(m.done)
	m.wg.Wait()
	close(m.out)
}
