package log

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/logging"
	"github.com/entropyx/soul/env"
)

type Severity uint8

const (
	Debug Severity = iota
	Info
	Warning
	Error
	Panic
)

var stackdriverClient *logging.Client

type Stackdriver struct {
	logger   *logging.Logger
	fields   Fields
	severity Severity
}

type StackdriverOptions struct {
	LogName  string
	Severity Severity
}

func StartStackdriver() {
	var err error
	ctx := context.Background()
	stackdriverClient, err = logging.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func NewStackdriver(opts *StackdriverOptions) *Stackdriver {
	logger := stackdriverClient.Logger(opts.LogName)
	return &Stackdriver{logger: logger, severity: opts.Severity, fields: Fields{}}
}

func (s *Stackdriver) Debug(args ...interface{}) {
	s.log(Debug, args...)
}

func (s *Stackdriver) Debugf(format string, args ...interface{}) {
	s.Debug(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Error(args ...interface{}) {
	s.log(Error, args...)
}

func (s *Stackdriver) Errorf(format string, args ...interface{}) {
	s.Error(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Fields() Fields {
	return s.fields
}

func (s *Stackdriver) Info(args ...interface{}) {
	s.log(Info, args...)
}

func (s *Stackdriver) Infof(format string, args ...interface{}) {
	s.Info(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Panic(args ...interface{}) {
	s.log(Panic, args...)
	os.Exit(1)
}

func (s *Stackdriver) Panicf(format string, args ...interface{}) {
	s.Panic(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Warning(args ...interface{}) {
	s.log(Warning, args...)
}

func (s *Stackdriver) Warningf(format string, args ...interface{}) {
	s.Warning(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) WithField(key string, value interface{}) Logger {
	newLogger := *s
	newLogger.fields[key] = value
	return &newLogger
}

func (s *Stackdriver) log(severity Severity, args ...interface{}) {
	if !s.canLog(severity) {
		return
	}
	fields := s.fields
	fields["message"] = fmt.Sprint(args...)
	entry := logging.Entry{}
	if trace, ok := fields["trace"]; ok {
		entry.Trace = trace.(string)
		delete(fields, "trace")
	}
	entry.Payload = fields
	entry.Severity = setSeverity(severity)
	entry.Labels = map[string]string{
		"env": env.Mode,
	}
	s.logger.Log(entry)
}

func (s *Stackdriver) canLog(severity Severity) bool {
	return severity >= s.severity
}

func setSeverity(severity Severity) logging.Severity {
	switch severity {
	case Debug:
		return logging.Debug
	case Info:
		return logging.Info
	case Warning:
		return logging.Warning
	case Panic:
		return logging.Critical
	default:
		return logging.Default
	}
}
