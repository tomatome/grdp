package glog

import (
	"log"
	"sync"
)

var (
	logger *log.Logger
	level  LEVEL
	mu     sync.Mutex
)

type LEVEL int

const (
	DEBUG LEVEL = iota
	INFO
	WARN
	ERROR
	NONE
)

func SetLogger(l *log.Logger) {
	logger = l
}

func SetLevel(l LEVEL) {
	level = l
}

func checkLogger() {
	if logger == nil && level != NONE {
		panic("logger not inited")
	}
}

func Debug(v ...interface{}) {
	checkLogger()
	if level <= DEBUG {
		mu.Lock()
		defer mu.Unlock()
		logger.SetPrefix("[DEBUG]")
		logger.Print(v)
	}
}

func Info(v ...interface{}) {
	checkLogger()
	if level <= INFO {
		mu.Lock()
		defer mu.Unlock()
		logger.SetPrefix("[INFO]")
		logger.Print(v)
	}
}

func Warn(v ...interface{}) {
	checkLogger()
	if level <= WARN {
		mu.Lock()
		defer mu.Unlock()
		logger.SetPrefix("[WARN]")
		logger.Print(v)
	}
}

func Error(v ...interface{}) {
	checkLogger()
	if level <= ERROR {
		mu.Lock()
		defer mu.Unlock()
		logger.SetPrefix("[ERROR]")
		logger.Print(v)
	}
}
