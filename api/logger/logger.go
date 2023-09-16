package logger

import (
	"log"
	"os"
)

func Info(message string) {
	log.Printf("INFO: %v", message)
}

func Warn(message string) {
	log.Printf("WARN: %v", message)
}

func Error(message string) {
	log.Printf("ERROR: %v", message)
}

func Fatal(message string) {
	log.Printf("FATAL: %v", message)
	os.Exit(1)
}
