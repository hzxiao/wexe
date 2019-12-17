package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	watchFile = "file"
	watchTime = "time"
)

type Event interface {
	String() string
}

type FSEvnent struct {
	Filename string
	Op fsnotify.Op
}

func (e FSEvnent) String() string {
	return fmt.Sprintf("event: filename = %v, op = %v", e.Filename, e.Op)
}

type TimeEvent struct {
	Time time.Time
}

func (e TimeEvent) String() string {
	return e.Time.String()
}
type Watcher interface {
	Watch()
	Close() error
}

type FSWatcher struct {
	fuzzyFiles   []string
	exactWatcher *fsnotify.Watcher
	fuzzyWatcher *fsnotify.Watcher
	exitChan     chan bool
	onChange     func(e Event) error
}

func NewFSWatcher(file string, onChange func(e Event) error) (*FSWatcher, error) {
	w := &FSWatcher{
		onChange: onChange,
		exitChan: make(chan bool),
	}
	var err error
	w.exactWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	files := strings.Split(file, ",")
	exactFiles, fuzzyFiles, err := GroupByWildcard(files)
	if err != nil {
		return nil, err
	}
	for i := range exactFiles {
		err = w.exactWatcher.Add(filepath.ToSlash(exactFiles[i]))
		if err != nil {
			return nil, err
		}
	}

	w.fuzzyWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if len(fuzzyFiles) > 0 {
		for i := range fuzzyFiles {
			w.fuzzyFiles = append(w.fuzzyFiles, filepath.ToSlash(fuzzyFiles[i]))
		}
		cleanFluzzyPaths, err := CleanFuzzyPath(fuzzyFiles)
		if err != nil {
			return nil, err
		}
		dirs, err := ParentDir(cleanFluzzyPaths...)
		if err != nil {
			return nil, err
		}
		for i := range dirs {
			err = w.fuzzyWatcher.Add(filepath.ToSlash(dirs[i]))
			if err != nil {
				return nil, err
			}
		}
	}
	return w, nil
}

func (w *FSWatcher) Watch() {
	defer fmt.Println("stop watch")

	exeTime := time.Now()
	var lock sync.Mutex

	var onEvent = func(event fsnotify.Event, fuzzy bool) error {
		if fuzzy && !w.matchFuzzyFile(event.Name) {
			return nil
		}
		lock.Lock()
		defer lock.Unlock()
		d := time.Now().Sub(exeTime).Seconds()
		if d < 1 {
			return nil
		}
		fmt.Printf("file: %v : %v\n", event.Name, event.Op)
		err := w.onChange(&FSEvnent{Filename: event.Name, Op: event.Op})
		if err != nil {
			return err
		}
		exeTime = time.Now()
		return nil
	}

	for {
		select {
		case <-w.exitChan:
			return
		case event, ok := <-w.exactWatcher.Events:
			if !ok {
				return
			}
			err := onEvent(event, false)
			if err != nil {
				fmt.Println(err)
			}
		case event, ok := <-w.fuzzyWatcher.Events:
			if !ok {
				break
			}
			err := onEvent(event, true)
			if err != nil {
				fmt.Println(err)
			}
		case err, ok := <-w.exactWatcher.Errors:
			if !ok {
				return
			}
			fmt.Println("error:", err)
		case err, ok := <-w.fuzzyWatcher.Errors:
			if !ok {
				return
			}
			fmt.Println("error:", err)
		}
	}
}

func (w *FSWatcher) matchFuzzyFile(file string) bool {
	for _, pattern := range w.fuzzyFiles {
		ok, _ := filepath.Match(pattern, file)
		if ok {
			return true
		}
	}
	return false
}

func (w *FSWatcher) Close() error {
	w.exitChan <- true
	w.exactWatcher.Close()
	w.fuzzyWatcher.Close()
	return nil 
}

type TimeWatcher struct {
	duration time.Duration
	onChange func(e Event) error
	exitChan     chan bool
}

func NewTimeWatcher(dur string, onChange func(e Event) error) (*TimeWatcher, error) {
	var err error
	w := &TimeWatcher{
		onChange: onChange,
		exitChan: make(chan bool),
	}

	w.duration, err = time.ParseDuration(dur)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *TimeWatcher) Watch()  {
	defer fmt.Println("stop watch")

	ticker := time.NewTicker(w.duration)
	defer ticker.Stop()

	for {
		select{
		case <- w.exitChan:
			return
		case  t := <- ticker.C:
			err := w.onChange(&TimeEvent{Time: t})
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (w *TimeWatcher) Close() error {
	w.exitChan <- true
	return nil 
}