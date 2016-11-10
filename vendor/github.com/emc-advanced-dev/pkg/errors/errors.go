package errors


import (
	"fmt"
	"runtime"
	"strings"
)

type lxerror struct {
	error
	message string
	err     error
	file    string
}

func New(message string, err error) *lxerror {
	return &lxerror{
		message: message,
		err:     err,
		file: getTrace(),
	}
}

func (e *lxerror) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s %s: {%s}", e.file, e.message, e.err.Error())
	}
	return fmt.Sprintf("%s %s", e.file, e.message)
}

func getTrace() string {
	_, fn, line, _ := runtime.Caller(2)
	pathComponents := strings.Split(fn, "/")
	var truncatedPath string
	if len(pathComponents) > 3 {
		truncatedPath = strings.Join(pathComponents[len(pathComponents) - 2:], "/")
	} else {
		truncatedPath = strings.Join(pathComponents, "/")
	}
	file := fmt.Sprintf("[%s:%d]", truncatedPath, line)
	return file
}
