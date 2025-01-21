package logger

import (
	"io"
	"log"
	"os"

	"ask-web/pkg/config"
)

func Init(opts *config.Opts) {
	file, err := os.OpenFile(opts.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	var multiWriter io.Writer
	if opts.LogStderr {
		multiWriter = io.MultiWriter(file, os.Stderr)
	} else {
		multiWriter = file
	}

	log.SetOutput(multiWriter)

	log.SetFlags(log.Ldate | log.Ltime)
}

func Debug(v ...interface{}) {
	log.SetPrefix("[DEBUG] ")
	log.Println(v...)
}

func Info(v ...interface{}) {
	log.SetPrefix("[INFO] ")
	log.Println(v...)
}

func Warn(v ...interface{}) {
	log.SetPrefix("[WARN] ")
	log.Println(v...)
}

func Error(v ...interface{}) {
	log.SetPrefix("[ERROR] ")
	log.Println(v...)
}

func Fatal(v ...interface{}) {
	log.SetPrefix("[FATAL] ")
	log.Fatalln(v...)
}
