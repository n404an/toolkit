package log

import (
	"context"
	"fmt"
	"github.com/spf13/cast"
	"sync"
)

// Start logLevel: "MUTE", "DEBUG", "INFO", "ERROR", "FATAL", "PANIC"
func Start(ctx context.Context, wg *sync.WaitGroup, logLevel string) { start(ctx, wg, logLevel) }

func Debug(fn ...string) *msg { return newMsg(debugLevel, fn...) }
func Info(fn ...string) *msg  { return newMsg(infoLevel, fn...) }
func Error(fn ...string) *msg { return newMsg(errorLevel, fn...) }
func Fatal(fn ...string) *msg { return newMsg(fatalLevel, fn...) }
func Panic(fn ...string) *msg { return newMsg(panicLevel, fn...) }

func (m *msg) Ln(args ...any) { m.message(true, args...) }
func (m *msg) F(args ...any)  { m.message(false, args...) }

func (m *msg) message(isLn bool, args ...any) {
	switch len(args) {
	case 0:
		m.msg = lnLabel
	default:
		switch isLn {
		case true:
			m.msg = fmt.Sprintln(args...)
		case false:
			m.msg = fmt.Sprintf(cast.ToString(args[0])+lnLabel, args[1:]...)
		}
	}
	m.send()
}

func start(ctx context.Context, wg *sync.WaitGroup, logLevel string) {
	baseCtx = ctx
	baseWg = wg
	setLocalLevel(logLevel)

	baseWg.Add(1)
	if lvl == muteLevel {
		go discard()
	} else {
		go read()
	}
}

var (
	baseCtx context.Context
	baseWg  *sync.WaitGroup
	lvl     level
)
