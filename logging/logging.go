package logging

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	logger := logrus.New()
	tf := &prefixed.TextFormatter{}
	tf.FullTimestamp = true
	logger.Formatter = tf
	return &Logger{logger}
}
