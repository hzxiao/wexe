package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Wexe struct {
	w Watcher
	e *Executor
}

func NewWexe(watchType string, watchObj string, cmd string) (*Wexe, error) {
	var err error
	we := &Wexe{}
	switch watchType {
	case watchFile:
		we.w, err = NewFSWatcher(watchObj, we.onChange)
	case watchTime:
		we.w, err = NewTimeWatcher(watchObj, we.onChange)
	default:
		return nil, fmt.Errorf("unknown watch type")
	}
	if err != nil {
		return nil, err
	}

	we.e = NewExecutor(cmd)
	return we, nil
}

func (we *Wexe) Run() {
	we.w.Watch()
}

func (we *Wexe) onChange(e Event) error {
	return we.e.Restart()
}

func (we *Wexe) Action(action string) error {
	switch action {
	case "s":
		fmt.Println("start to exec...")
		return we.e.Start()
	case "p":
		fmt.Println("stop exec...")
		return we.e.Stop()
	case "r":
		fmt.Println("restart exec...")
		return we.e.Restart()
	default:
		return fmt.Errorf("unknown action")
	}
}

func (we *Wexe) Close() error {
	err := we.w.Close()
	if err != nil {
		return err
	}
	return we.e.Stop()
}

var (
	wf  string
	wt string
	cmd string
)

func init() {
	flag.StringVar(&wf, "wf", "", "files to watch")
	flag.StringVar(&wt, "wt", "", "time duration to watch")
	flag.StringVar(&cmd, "cmd", "", "command")
}

func main() {
	flag.Parse()

	if wf == "" && wt == "" {
		fmt.Println("you need to watch files or watch time")
		os.Exit(1)
	}
	if wf != "" && wt != "" {
		fmt.Println("you should to watch files only or watch time only")
		os.Exit(1)
	}
	
	var watchType string
	var watchObj string
	if wf != "" {
		watchType = watchFile
		watchObj = wf
	} else {
		watchType = watchTime
		watchObj = wt
	}
	we, err := NewWexe(watchType, watchObj, cmd)
	if err != nil {
		panic(err)
	}

	go we.Run()

	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			buf, err := r.ReadSlice('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}
			action := strings.TrimSpace(string(buf))
			if action == "" {
				continue
			}
			err = we.Action(action)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		fmt.Printf("wexe get a signal %s\n", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			we.Close()
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
