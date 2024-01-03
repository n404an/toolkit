package log

import (
	"github.com/n404an/toolkit/pool"
	"strings"
)

var msgPool = pool.New(func() *msg {
	return &msg{
		buf: make([]byte, 0, 1024)}
})

func newMsg(lvl level, fn ...string) *msg {
	m := msgPool.Get()
	m.level = lvl

	switch len(fn) {
	case 0:
		m.fn = ""
	case 1:
		m.fn = fn[0]
	default:
		m.fn = strings.Join(fn, "_")
	}
	m.reset()
	return m
}

func (m *msg) put()   { msgPool.Put(m) }
func (m *msg) reset() { m.msg = "" }
