package wireguard

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	LogLevelError = iota
	LogLevelInfo
	LogLevelDebug
)

type Logger struct {
	Debug *log.Logger
	Info  *log.Logger
	Error *log.Logger
}

func NewLogger(level int, prepend string) *Logger {
	output := os.Stdout
	logger := new(Logger)

	logErr, logInfo, logDebug := func() (io.Writer, io.Writer, io.Writer) {
		if level >= LogLevelDebug {
			return output, output, output
		}
		if level >= LogLevelInfo {
			return output, output, ioutil.Discard
		}
		return output, ioutil.Discard, ioutil.Discard
	}()

	logger.Debug = log.New(logDebug,
		"DEBUG: "+prepend,
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	logger.Info = log.New(logInfo,
		"INFO: "+prepend,
		log.Ldate|log.Ltime,
	)
	logger.Error = log.New(logErr,
		"ERROR: "+prepend,
		log.Ldate|log.Ltime,
	)
	return logger
}
