package logger

import (
	"io"
	"log"
	"os"
	"sync"

	"ask-web/pkg/config"
)

var (
	logInst *SimpleLogger
	once    sync.Once
)

type SimpleLogger struct {
	logger *log.Logger
}

func GetLogger() *SimpleLogger {
	return logInst
}

func Init(opts *config.Opts) error {
	var err error
	once.Do(func() {
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
		logInst = &SimpleLogger{
			logger: log.New(multiWriter, "", 0),
		}
	})

	return err
}

func (*SimpleLogger) Debug(v ...interface{}) {
	log.SetPrefix("[DEBUG] ")
	log.Println(v...)
}

func (*SimpleLogger) Info(v ...interface{}) {
	log.SetPrefix("[INFO] ")
	log.Println(v...)
}

func (*SimpleLogger) Warn(v ...interface{}) {
	log.SetPrefix("[WARN] ")
	log.Println(v...)
}

func (*SimpleLogger) Error(v ...interface{}) {
	log.SetPrefix("[ERROR] ")
	log.Println(v...)
}

func (*SimpleLogger) Fatal(v ...interface{}) {
	log.SetPrefix("[FATAL] ")
	log.Fatalln(v...)
}
