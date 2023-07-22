package internal

import (
	"log"
	"os"

	"github.com/xorgal/xtun-core/pkg/config"
)

func Fatal(v ...any) {
	log.Print(v...)
	handleFatalErr()
}

func Fatalf(format string, v ...any) {
	log.Printf(format, v...)
	handleFatalErr()
}

func Fatalln(v ...any) {
	log.Println(v...)
	handleFatalErr()
}

func handleFatalErr() {
	if !config.AppConfig.GUIMode {
		os.Exit(1)
	}
}
