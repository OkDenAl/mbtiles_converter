package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"runtime"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type writeHook struct {
	writers   []io.Writer
	logLevels []logrus.Level
}

func (wh *writeHook) Levels() []logrus.Level {
	return wh.logLevels
}

func (wh *writeHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range wh.writers {
		_, err = w.Write([]byte(line))
		if err != nil {
			break
		}
	}
	return err
}

type logger struct {
	*logrus.Logger
}

func Init() Logger {
	l := logrus.New()
	l.SetReportCaller(true)
	l.SetFormatter(&logrus.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			return fmt.Sprintf("%s()", frame.Function), fmt.Sprintf("%s:%d", filename, frame.Line)
		},
		DisableColors: true,
	})
	err := os.MkdirAll("logs", 0777)
	if err != nil {
		l.Fatal(err)
	}
	f, err := os.OpenFile("logs/all.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		l.Fatal(err)
	}
	f.Write([]byte("\n"))
	l.AddHook(&writeHook{
		writers:   []io.Writer{f},
		logLevels: logrus.AllLevels,
	})
	l.SetLevel(logrus.TraceLevel)
	return l
}
