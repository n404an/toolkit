package log

import (
	"os"
)

var logChan = make(chan *msg, 10000)

func discard() {
	defer baseWg.Done()
	for {
		select {
		case <-baseCtx.Done():
			return
		case m := <-logChan:
			m.put()
		}
	}
}

func read() {
	defer baseWg.Done()
	for {
		select {
		case <-baseCtx.Done():
			return
		case m := <-logChan:
			if m.level >= lvl {
				m.print()
			}
			m.put()
		}
	}
}

func (m *msg) print() {
	byteMsg := m.toByte()
	if m.level <= infoLevel {
		_, _ = os.Stdout.Write(byteMsg)
	} else {
		_, _ = os.Stderr.Write(byteMsg)
	}
}
