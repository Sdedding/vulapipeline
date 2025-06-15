package core

import (
	"io"
	"log"
	"os"
)

var (
	debugLog = log.New(io.Discard, "DEBUG:", 0)
	infoLog  = log.New(os.Stderr, "INFO:", 0)
	warnLog  = log.New(os.Stderr, "WARN:", 0)
)

func LogDebugf(format string, args ...any) {
	debugLog.Printf(format, args...)
}

func LogInfof(format string, args ...any) {
	infoLog.Printf(format, args...)
}

func LogWarnf(format string, args ...any) {
	warnLog.Printf(format, args...)
}

func LogDebug(v ...any) {
	debugLog.Print(v...)
}

func LogInfo(v ...any) {
	infoLog.Print(v...)
}

func LogWarn(v ...any) {
	warnLog.Print(v...)
}

func SetDebugLogOutput(w io.Writer) {
	debugLog.SetOutput(w)
}
