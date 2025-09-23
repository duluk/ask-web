package logger

import (
	"fmt"
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
	quiet  bool
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
			quiet:  opts.Quiet,
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
	// I don't actually like how I'm doing 'quiet' right now, but I haven't
	// decided on a better way and I want it to work. I just want fatals to be
	// loggged to stdout...unless don't.
	if !logInst.quiet {
		fmt.Printf("FATAL: ")
		fmt.Println(v...)
	}

	log.SetPrefix("[FATAL] ")
	log.Fatalln(v...)
}
