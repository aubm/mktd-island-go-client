package log

import (
	"github.com/Sirupsen/logrus"
)

type LoggerInterface interface {
	Debug(msg string, fields Fields)
	Info(msg string, fields Fields)
	Warn(msg string, fields Fields)
	Error(msg string, fields Fields)
	Fatal(msg string, fields Fields)
}

type Logger struct{}

func (l *Logger) Debug(msg string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Debug(msg)
}

func (l *Logger) Info(msg string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Info(msg)
}

func (l *Logger) Warn(msg string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Warn(msg)
}

func (l *Logger) Error(msg string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Error(msg)
}

func (l *Logger) Fatal(msg string, fields Fields) {
	logrus.WithFields(logrus.Fields(fields)).Fatal(msg)
}

type Fields map[string]interface{}

func ConfigureDebugLevel() {
	logrus.SetLevel(logrus.DebugLevel)
}
