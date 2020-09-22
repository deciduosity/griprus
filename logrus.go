package griprus

import (
	"fmt"
	"math"
	"strings"

	"github.com/deciduosity/grip/level"
	"github.com/deciduosity/grip/message"
	"github.com/deciduosity/grip/send"
	"github.com/sirupsen/logrus"
)

// ConvertPriority takes a grip Priorty and converts it into the
// equivalent logrus level.
func ConvertPriority(p level.Priority) logrus.Level {
	switch {
	case p > level.Warning:
		return logrus.ErrorLevel
	case p > level.Notice:
		return logrus.WarnLevel
	case p > level.Debug:
		return logrus.InfoLevel
	case p > level.Invalid:
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

func fieldsToLogrus(f message.Fields) logrus.Fields {
	return func(data interface{}) logrus.Fields { return data.(map[string]interface{}) }(f)
}

// ConvertMessage takes a grip message instance and converts it to the
// equivalent logrus.Entry.
func ConvertMessage(logger *logrus.Logger, m message.Composer) *logrus.Entry {
	e := logrus.NewEntry(logger)
	if !m.Loggable() {
		// the lower logrus levels, so our approach here is to
		// make it maximally low priority by making the number
		// maximally high.
		e.Level = math.MaxUint32
		return e
	}

	e.Level = ConvertPriority(m.Priority())

	if !message.IsStructured(m) {
		e.Message = m.String()
		return e
	}

	switch payload := m.Raw().(type) {
	case error:
		return e.WithField(logrus.ErrorKey, payload)
	case message.Fields:
		return e.WithFields(fieldsToLogrus(payload))
	case *message.ProcessInfo:
		return e.WithField("procinfo", payload)
	case *message.SystemInfo:
		return e.WithField("sysinfo", payload)
	case *message.GoRuntimeInfo:
		return e.WithField("goinfo", payload)
	case message.StackTrace:
		switch context := payload.Context.(type) {
		case message.Composer:
			e = ConvertMessage(logger, context)
			e.WithField("trace", payload.Frames)
		case string:
			e.Message = context
			e.WithField("trace", payload.Frames)
		case fmt.Stringer:
			e.Message = context.String()
			e.WithField("trace", payload.Frames)
		default:
			e.Message = fmt.Sprintf("%v", context)
			e.WithField("trace", payload.Frames)
		}
		return e
	case *message.Slack:
		e.WithField("target", payload.Target)
		for _, v := range payload.Attachments {
			e.WithField(v.Title, v.Text)
			for _, f := range v.Fields {
				e.WithField(strings.Join([]string{v.Title, f.Title}, "."), f.Value)
			}
		}
		return e
	case *message.Email:
		return e.WithField("message", payload)
	case *message.JiraIssue:
		return e.WithField("issue", payload)
	case *message.GithubStatus:
		return e.WithField("status", payload)
	default:
		return e.WithField("payload", payload)
	}

	return nil
}

// NewLogger produces a logger that writes all logs to the underlying
// grip Sender, configured as an io.Writer.
func NewLogger(s send.Sender) *logrus.Logger {
	logger := logrus.New()
	logger.Out = send.NewWriterSender(s)
	logger.Level = ConvertPriority(s.Level().Default)
	return logger
}
