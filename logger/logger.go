package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

type Level uint

const (
	OFF   Level = 100
	ERROR Level = 40
	WARN  Level = 30
	INFO  Level = 20
	DEBUG Level = 10
	ALL   Level = 0

	sWARN  string = "warn"
	sINFO  string = "info"
	sERROR string = "error"
	sDEBUG string = "debug"
)

type Logger struct {
	filename      string
	Dir           string
	file          *os.File
	fullPath      string
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
	LogLevel      Level
}

func (l *Logger) getLogFilePath() string {

	filename := l.filename + "_" + time.Now().Format("2006-01-02") + ".log"

	path := path.Join(l.Dir, filename)

	return path
}

func (l *Logger) rotate() {
	var err error
	ticker := time.NewTicker(time.Minute)
	for {
		<-ticker.C
		newpath := l.getLogFilePath()
		if l.fullPath != newpath {
			l.fullPath = newpath
			l.file, err = os.OpenFile(l.fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				log.Fatal(err)
			}

			l.WarningLogger = log.New(l.file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
			l.InfoLogger = log.New(l.file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
			l.ErrorLogger = log.New(l.file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
			l.DebugLogger = log.New(l.file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}
}

func New(filename string) *Logger {
	var err error
	l := Logger{
		filename: filename,
		Dir:      "./logs",
	}

	if _, err := os.Stat(l.Dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(l.Dir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	l.fullPath = l.getLogFilePath()
	l.file, err = os.OpenFile(l.fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	l.InfoLogger = log.New(l.file, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.WarningLogger = log.New(l.file, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.ErrorLogger = log.New(l.file, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	l.DebugLogger = log.New(l.file, "DEBUG: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	l.SetLevelS(os.Getenv("LOG_LEVEL"))
	go l.rotate()
	return &l
}

func (log *Logger) WARN(t ...string) {
	if log.LogLevel <= WARN {
		fmt.Print(time.Now().Format("15:04:05.000000000"))
		fmt.Println(t)
		log.WarningLogger.Output(2, fmt.Sprint(t))
	}
}
func (log *Logger) INFO(t ...string) {
	if log.LogLevel <= INFO {
		fmt.Print(time.Now().Format("15:04:05.000000000"))
		fmt.Println(t)
		log.InfoLogger.Output(2, fmt.Sprint(t))
	}
}
func (log *Logger) ERROR(t ...string) {
	if log.LogLevel <= ERROR {
		fmt.Print(time.Now().Format("15:04:05.000000000"))
		fmt.Println(t)
		log.ErrorLogger.Output(2, fmt.Sprint(t))
	}
}
func (log *Logger) DEBUG(t ...string) {
	if log.LogLevel <= DEBUG {
		fmt.Print(time.Now().Format("15:04:05.000000000"))
		fmt.Println(t)
		log.DebugLogger.Output(2, fmt.Sprint(t))
	}
}

func (log *Logger) SetLevel(lvl Level) {
	log.LogLevel = lvl
}

func (log *Logger) SetLevelS(lvl string) {
	switch lvl {
	case sWARN:
		log.LogLevel = WARN
	case sINFO:
		log.LogLevel = INFO
	case sERROR:
		log.LogLevel = ERROR
	case sDEBUG:
		log.LogLevel = DEBUG
	default:
		log.LogLevel = INFO
	}
}
