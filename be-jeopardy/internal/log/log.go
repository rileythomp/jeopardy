package log

import (
	"fmt"
	"log"
	"os"
)

func Debugf(s string, args ...any) {
	log.Printf("[DEBUG] %s\n", fmt.Sprintf(s, args...))
}

func Infof(s string, args ...any) {
	log.Printf("[INF0] %s\n", fmt.Sprintf(s, args...))
}

func Warnf(s string, args ...any) {
	log.Printf("[WARN] %s\n", fmt.Sprintf(s, args...))
}

func Errorf(s string, args ...any) {
	log.Printf("[ERROR] %s\n", fmt.Sprintf(s, args...))
}

func Fatalf(s string, args ...any) {
	log.Printf("[FATAL] %s\n", fmt.Sprintf(s, args...))
	os.Exit(1)
}

func Panicf(s string, args ...any) {
	log.Printf("[PANIC] %s\n", fmt.Sprintf(s, args...))
	panic(fmt.Sprintf(s, args...))
}
