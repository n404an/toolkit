package log

import "time"

type msg struct {
	buf   []byte
	level level
	fn    string
	msg   string
}

func (m *msg) send() { logChan <- m }

func (m *msg) toByte() []byte {
	m.buf = m.buf[:0]
	m.buf = time.Now().UTC().AppendFormat(m.buf, logFormat)

	if level := m.level.byte(); level != nil {
		m.buf = append(m.buf, level...)
	}
	if len(m.fn) > 0 {
		m.buf = append(m.buf, s2b(m.fn)...)
	}
	if len(m.msg) > 0 {
		m.buf = append(m.buf, space)
		m.buf = append(m.buf, s2b(m.msg)...)
	}
	return m.buf
}
