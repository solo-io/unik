package daemon

import (
	"github.com/Sirupsen/logrus"
	"io"
	"fmt"
	"runtime"
	"strings"
)

type unikLogrusHook struct {
	w io.Writer
	logContext string
}

func NewUnikLogrusHook(w io.Writer, logContext string) *unikLogrusHook {
	return &unikLogrusHook{
		w: w,
		logContext: logContext,
	}
}


func (h *unikLogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *unikLogrusHook) Fire(entry *logrus.Entry) error {
	entry.Message = addTrace(h.logContext, entry.Message, 1)
	logger := logrus.New()
	logger.Out = h.w
	switch entry.Level {
	case logrus.PanicLevel:
		logger.WithFields(entry.Data).Panic(entry.Message)
		break;
	case logrus.FatalLevel:
		logger.WithFields(entry.Data).Fatal(entry.Message)
		break;
	case logrus.ErrorLevel:
		logger.WithFields(entry.Data).Error(entry.Message)
		break;
	case logrus.WarnLevel:
		logger.WithFields(entry.Data).Warnf(entry.Message)
		break;
	case logrus.InfoLevel:
		logger.WithFields(entry.Data).Info(entry.Message)
		break;
	case logrus.DebugLevel:
		logger.WithFields(entry.Data).Debug(entry.Message)
		break;
	}
	return nil
}

func addTrace(logContext, message string, trace int) string {
	pc, fn, line, _ := runtime.Caller(trace)
	pathComponents := strings.Split(fn, "/")
	var truncatedPath string
	if len(pathComponents) > 3 {
		truncatedPath = strings.Join(pathComponents[len(pathComponents) - 2:], "/")
	} else {
		truncatedPath = strings.Join(pathComponents, "/")
	}
	fnName := runtime.FuncForPC(pc).Name()
	fnNameComponents := strings.Split(fnName, "/")
	truncatedFnName := fnNameComponents[len(fnNameComponents) - 1]

	file := fmt.Sprintf("(%s): %s[%s:%d] ", logContext, truncatedFnName, truncatedPath, line)

	return file + message
}