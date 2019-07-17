package log

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/logging"
	"github.com/entropyx/soul/env"
)

var stackdriverClient *logging.Client

type Stackdriver struct {
	logger   *logging.Logger
	fields   Fields
	severity logging.Severity
}

type StackdriverOptions struct {
	LogName  string
	Severity logging.Severity
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
	s.log(logging.Debug, args...)
}

func (s *Stackdriver) Debugf(format string, args ...interface{}) {
	s.Debug(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Error(args ...interface{}) {
	s.log(logging.Error, args...)
}

func (s *Stackdriver) Errorf(format string, args ...interface{}) {
	s.Error(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Fields() Fields {
	return s.fields
}

func (s *Stackdriver) Info(args ...interface{}) {
	s.log(logging.Info, args...)
}

func (s *Stackdriver) Infof(format string, args ...interface{}) {
	s.Info(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Panic(args ...interface{}) {
	s.log(logging.Critical, args...)
	os.Exit(1)
}

func (s *Stackdriver) Panicf(format string, args ...interface{}) {
	s.Panic(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) Warning(args ...interface{}) {
	s.log(logging.Warning, args...)
}

func (s *Stackdriver) Warningf(format string, args ...interface{}) {
	s.Warning(fmt.Sprintf(format, args...))
}

func (s *Stackdriver) WithField(key string, value interface{}) Logger {
	newLogger := *s
	newLogger.fields[key] = value
	return &newLogger
}

func (s *Stackdriver) log(severity logging.Severity, args ...interface{}) {
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
	entry.Severity = severity
	entry.Labels = map[string]string{
		"env": env.Mode,
	}
	s.logger.Log(entry)
}

func (s *Stackdriver) canLog(severity logging.Severity) bool {
	return s.severity >= severity
}
