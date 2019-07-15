package log

import "github.com/sirupsen/logrus"

type Logrus struct {
	*logrus.Entry
}

type LogrusOptions struct {
}

func NewLogrus(logger *logrus.Logger) *Logrus {
	entry := logrus.NewEntry(logger)
	return &Logrus{entry}
}

func (l *Logrus) WithFields(fields Fields) Logger {
	return &Logrus{l.Entry.WithFields(logrus.Fields(fields))}
}
