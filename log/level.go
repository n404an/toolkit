package log

import "strings"

type level int8

const (
	debugLevel level = iota
	infoLevel
	errorLevel
	fatalLevel
	panicLevel

	muteLevel level = -1
)

func (l level) byte() []byte {
	switch l {
	case debugLevel:
		return levelDebugValue
	case infoLevel:
		return levelInfoValue
	case errorLevel:
		return levelErrorValue
	case fatalLevel:
		return levelFatalValue
	case panicLevel:
		return levelPanicValue
	default:
		return nil
	}
}

func setLocalLevel(l string) {
	switch strings.ToUpper(l) {
	case "DEBUG":
		lvl = debugLevel
	case "INFO":
		lvl = infoLevel
	case "ERROR":
		lvl = errorLevel
	case "FATAL":
		lvl = fatalLevel
	case "PANIC":
		lvl = panicLevel
	case "MUTE":
		lvl = muteLevel
	default:
		lvl = infoLevel
	}
}
