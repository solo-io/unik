package util

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"math"
	"os/exec"
	"runtime"
	"strings"
)

type AddTraceHook struct {
	Full bool
}

func (h *AddTraceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *AddTraceHook) Fire(entry *logrus.Entry) error {
	entry.Message = h.addTrace(entry.Message)
	return nil
}

func (h *AddTraceHook) addTrace(message string) string {
	skip := 2
	ok := true
	var stackTrace []string
	for {
		var pc uintptr
		var fn string
		var line int
		pc, fn, line, ok = runtime.Caller(skip)
		if !ok {
			break
		}
		skip++

		fnName := runtime.FuncForPC(pc).Name()
		if strings.Contains(fnName, "logrus.") {
			continue
		}
		fnNameComponents := strings.Split(fnName, "/")
		truncatedFnName := fnNameComponents[len(fnNameComponents)-1]

		pathComponents := strings.Split(fn, "/")
		var truncatedPath string
		if len(pathComponents) > 3 {
			truncatedPath = strings.Join(pathComponents[len(pathComponents)-2:], "/")
		} else {
			truncatedPath = strings.Join(pathComponents, "/")
		}
		stackTrace = append(stackTrace, fmt.Sprintf("%s[%s:%d]", truncatedFnName, truncatedPath, line))
		if !h.Full {
			break
		}
	}

	maxLen := int(math.Max(float64(len(stackTrace)-2), 1))
	for i := 0; i < maxLen/2; i++ {
		tmp := stackTrace[i]
		stackTrace[i] = stackTrace[maxLen-i-1]
		stackTrace[maxLen-i-1] = tmp
	}
	file := strings.Join(stackTrace[:maxLen], "\n")
	message = file + "\n" + message
	return message
}

func LogCommand(cmd *exec.Cmd, asDebug bool) {
	logrus.WithField("command", cmd.Args).Debugf("running command")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			if asDebug {
				logrus.Debugf(in.Text())
			} else {
				logrus.Infof(in.Text())
			}
		}
	}()
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			logrus.Debugf(in.Text())
		}
	}()
}

type TeeHook struct {
	W io.Writer
}

func (h *TeeHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TeeHook) Fire(entry *logrus.Entry) error {
	logger := logrus.New()
	logger.Out = h.W
	switch entry.Level {
	case logrus.PanicLevel:
		logger.WithFields(entry.Data).Panic(entry.Message)
		break
	case logrus.FatalLevel:
		logger.WithFields(entry.Data).Fatal(entry.Message)
		break
	case logrus.ErrorLevel:
		logger.WithFields(entry.Data).Error(entry.Message)
		break
	case logrus.WarnLevel:
		logger.WithFields(entry.Data).Warnf(entry.Message)
		break
	case logrus.InfoLevel:
		logger.WithFields(entry.Data).Info(entry.Message)
		break
	case logrus.DebugLevel:
		logger.WithFields(entry.Data).Info(entry.Message)
		break
	}
	return nil
}
