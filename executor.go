package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Executor struct {
	cmd      *exec.Cmd
	exe      string
	args     []string
	running  bool
	exitChan chan int
}

func NewExecutor(cmdStr string) *Executor {
	if cmdStr == "" {
		panic("cmd can't be enpty")
	}
	cmds := strings.Split(cmdStr, " ")
	var args []string
	if len(cmds) > 1 {
		args = cmds[1:]
	}

	return &Executor{
		exe:      cmds[0],
		args:     args,
		exitChan: make(chan int),
	}
}

func (e *Executor) printOutput(r io.Reader) {
	reader := bufio.NewReader(r)
	defer fmt.Println("stop print output")
	for e.running {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		fmt.Printf(">>> %v\n", string(l))
	}
}

func (e *Executor) Start() error {
	e.running = true
	cmd := exec.Command(e.exe, e.args...)
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Env = os.Environ()

	e.cmd = cmd
	go e.printOutput(r)

	go func() {
		var code int
		err := e.cmd.Run()
		if err != nil {
			code = 1
			if e.cmd.ProcessState != nil && e.cmd.ProcessState.Sys() != nil {
				if ws, ok := e.cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
					code = ws.ExitStatus()
				}
			} else {
				fmt.Println(err.Error())
			}
		}
		w.Close()
		// r.Close()
		e.running = false
		// e.exitChan <- code
		_ = code
		e.cmd = nil
	}()

	return nil
}

func (e *Executor) Stop() error {
	if !e.running || e.cmd == nil || e.cmd.Process == nil {
		return nil
	}

	return e.cmd.Process.Kill()
}

func (e *Executor) Restart() error {
	err := e.Stop()
	if err != nil {
		return err
	}
	return e.Start()
}
