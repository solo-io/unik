package qemu

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"net"
	"path/filepath"
)

func startDebuggerListener(port int) error {
	addr := fmt.Sprintf(":%v", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.New("establishing tcp listener on "+addr, err)
	}
	logrus.Info("listening on " + addr)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				logrus.WithError(err).Warnf("failed to accept debugger connection")
				continue
			}
			go connectDebugger(conn)
		}
	}()
	return nil
}

func connectDebugger(conn net.Conn) {
	if debuggerTargetImageName == "" {
		logrus.Error("no debug instance is currently running")
		return
	}
	container := unikutil.NewContainer("rump-debugger-qemu").
		WithNet("host").
		WithVolume(filepath.Dir(getKernelPath(debuggerTargetImageName)), "/opt/prog/").
		Interactive(true)

	cmd := container.BuildCmd(
		"/opt/gdb-7.11/gdb/gdb",
		"-ex", "target remote 192.168.99.1:1234",
		"/opt/prog/program.bin",
	)
	conn.Read([]byte("GET / HTTP/1.0\r\n\r\n"))
	logrus.WithField("command", cmd.Args).Info("running debug command")
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn
	if err := cmd.Start(); err != nil {
		logrus.WithError(err).Error("error starting debugger container")
		return
	}
	defer func() {
		//reset debugger target
		debuggerTargetImageName = ""
		container.Stop()
	}()

	for {
		if _, err := conn.Write([]byte{0}); err != nil {
			logrus.Debug("debugger disconnected: %v", err)
			return
		}
	}
}
