package main

import (
	"bufio"
	"time"
	"sync/atomic"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type Executor struct {
	cmd      *exec.Cmd
	exe      string
	args     []string
	running  uint32
	exitChan chan int
	lock     sync.Mutex
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

func (e *Executor) printOutput(r io.ReadCloser) {
	defer r.Close()

	reader := bufio.NewReader(r)
	for atomic.LoadUint32(&e.running) == 1 {
		l, _, err := reader.ReadLine()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		fmt.Printf("\t%v\n", string(l))
	}
}

func (e *Executor) Start() error {
	if atomic.LoadUint32(&e.running) == 1 {
		return fmt.Errorf("already start")
	}
	atomic.StoreUint32(&e.running, 1)
	cmd := exec.Command(e.exe, e.args...)
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Env = os.Environ()

	e.lock.Lock()
	e.cmd = cmd
	e.lock.Unlock()

	go e.printOutput(r)

	go func() {
		var code int

		defer func ()  {
			fmt.Printf("exit with status %v..\n", code)
		}()
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
		atomic.StoreUint32(&e.running, 0)
		_ = code
		e.lock.Lock()
		e.cmd = nil
		e.lock.Unlock()
	}()

	return nil
}

func (e *Executor) Stop() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if atomic.LoadUint32(&e.running) == 0 || e.cmd == nil || e.cmd.Process == nil {
		return nil
	}

	return e.cmd.Process.Kill()
}

func (e *Executor) Restart() error {
	err := e.Stop()
	if err != nil {
		return err
	}
	//wait†ˆng
	for atomic.LoadUint32(&e.running) == 1 {
		time.Sleep(time.Millisecond*100)
	}
	return e.Start()
}
