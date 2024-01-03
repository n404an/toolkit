package log

var (
	levelDebugValue = []byte("[DEBUG] ")
	levelInfoValue  = []byte("[INFO] ")
	levelErrorValue = []byte("[ERROR] ")
	levelFatalValue = []byte("[FATAL] ")
	levelPanicValue = []byte("[PANIC] ")

	space   = byte(' ')
	lnLabel = "\n"
)
