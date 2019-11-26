package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"strings"
)

type Watcher struct {
	exe       *Executor
	fsWatcher *fsnotify.Watcher
	exitChan  chan bool
}

func NewWatcher(files string, exe *Executor) (*Watcher, error) {
	fsWt, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range strings.Split(files, ",") {
		err = fsWt.Add(f)
		if err != nil {
			return nil, err
		}
	}

	w := &Watcher{
		exe:       exe,
		fsWatcher: fsWt,
		exitChan:  make(chan bool),
	}
	return w, nil
}

func (w *Watcher) Watch() {
	defer fmt.Println("stop watch")
	for {
		select {
		case <-w.exitChan:
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			println(111)
			// err := w.onChange()
			// if err != nil {
			// 	fmt.Println(err)
			// }
			fmt.Println("event:", event.Op.String())
			if event.Op&fsnotify.Write == fsnotify.Write {
				// fmt.Println("modified file:", event.Name)
			}
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			fmt.Println("error:", err)
		}
	}
}

func (w *Watcher) onChange() error {
	return w.exe.Restart()
}

func (w *Watcher) Notify(action string) error {
	switch action {
	case "s":
		return w.exe.Start()
	case "p":
		return w.exe.Stop()
	case "r":
		return w.exe.Restart()
	default:
		return fmt.Errorf("unknown action")
	}
}

func (w *Watcher) Close() error {
	err := w.exe.Start()
	if err != nil {
		return err
	}
	return w.fsWatcher.Close()
}
