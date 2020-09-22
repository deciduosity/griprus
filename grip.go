// Package griprus provides a bidirectional compatibility layer
// between grip and logrus.
package griprus

import (
	"github.com/deciduosity/grip"
	"github.com/deciduosity/grip/level"
	"github.com/deciduosity/grip/message"
	"github.com/deciduosity/grip/send"
	"github.com/sirupsen/logrus"
)

// ConvertLevel converts a logrus level to a grip priority.
func ConvertLevel(l logrus.Level) level.Priority {
	switch l {
	case logrus.PanicLevel, logrus.FatalLevel:
		return level.Emergency
	case logrus.ErrorLevel:
		return level.Error
	case logrus.WarnLevel:
		return level.Warning
	case logrus.InfoLevel:
		return level.Info
	case logrus.DebugLevel:
		return level.Debug
	case logrus.TraceLevel:
		return level.Trace
	default:
		return level.Info
	}
}

// ConvertEntry produces converts logrus entry into the equivalent
// grip message. The implementation uses grip's "Fields" message type.
func ConvertEntry(e *logrus.Entry) message.Composer {
	return message.NewFieldsMessage(ConvertLevel(e.Level), e.Message, fieldsToGrip(e.Data))
}

func fieldsToGrip(f logrus.Fields) message.Fields {
	return func(data interface{}) message.Fields { return data.(map[string]interface{}) }(f)
}

type logrusSender struct {
	logger *logrus.Logger
	*send.Base
}

// NewSender produces a Sender implementation that wraps a logger
// implementation. The name associated with the sender is derived from
// the global grip logger name, which is probably the process name.
func NewSender(logger *logrus.Logger) send.Sender {
	return &logrusSender{
		Base:   send.NewBase(grip.Name()),
		logger: logger,
	}
}

func (s *logrusSender) Send(m message.Composer) {
	if s.Base.Level().ShouldLog(m) {
		e := ConvertMessage(s.logger, m)
		e.Log(e.Level, e.Message)
	}
}
