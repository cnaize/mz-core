package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log *logger

const timeLayout = "2006-01-02 15:04:05 -07"

type logger struct {
	config Config
	logrus *logrus.Logger
}

func Init(config Config) {
	if log != nil {
		return
	}
	log = newLogger(config)

	Info("=========== Logger initialized ===========")
}

func Debug(format string, args ...interface{}) {
	get().log(logrus.DebugLevel, format, args...)
}

func Info(format string, args ...interface{}) {
	get().log(logrus.InfoLevel, format, args...)
}

func Warn(format string, args ...interface{}) {
	get().log(logrus.WarnLevel, format, args...)
}

func Error(format string, args ...interface{}) {
	get().log(logrus.ErrorLevel, format, args...)
}

func Fatal(format string, args ...interface{}) {
	get().log(logrus.FatalLevel, format, args...)
}

func Panic(format string, args ...interface{}) {
	get().log(logrus.PanicLevel, format, args...)
}

func newLogger(config Config) *logger {
	lgr := logrus.New()
	lgr.Out = os.Stdout
	lgr.Level = logrus.Level(config.Lvl)
	lgr.Formatter = &logrus.TextFormatter{FullTimestamp: true, TimestampFormat: timeLayout}

	return &logger{
		config: config,
		logrus: lgr,
	}
}

func get() *logger {
	if log == nil {
		panic("logger not initialized")
	}
	return log
}

func (l *logger) log(level logrus.Level, format string, args ...interface{}) {
	if level > logrus.Level(l.config.Lvl) {
		return
	}

	switch level {
	case logrus.PanicLevel:
		l.logrus.Panicf(format, args...)
	case logrus.FatalLevel:
		l.logrus.Fatalf(format, args...)
	case logrus.ErrorLevel:
		l.logrus.Errorf(format, args...)
	case logrus.WarnLevel:
		l.logrus.Warnf(format, args...)
	case logrus.InfoLevel:
		l.logrus.Infof(format, args...)
	default:
		l.logrus.Debugf(format, args...)
	}
}
