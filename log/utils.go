package log

import (
	"unsafe"
)

const (
	logFormat = "2006-01-02 15:04:05.000 "
)

func s2b(s string) []byte { return unsafe.Slice(unsafe.StringData(s), len(s)) }
func b2s(b []byte) string { return unsafe.String(&b[0], len(b)) }
