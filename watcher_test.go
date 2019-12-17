package main

import (
	"github.com/hzxiao/goutil/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func mkdir(t *testing.T, dir string) string {
	dir = filepath.Join("testdata", dir)
	err := os.Mkdir(dir, os.ModePerm)
	assert.NoError(t, err)

	return dir
}

func remove(t *testing.T, s string) {
	info, err := os.Stat(s)
	assert.NoError(t, err)
	if info.IsDir() {
		err = os.RemoveAll(s)
	} else {
		err = os.Remove(s)
	}
	assert.NoError(t, err)
}

func write2file(t *testing.T, fp string, content string) {
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	assert.NoError(t, err)
	defer f.Close()

	if content != "" {
		_, err = f.WriteString(content)
		assert.NoError(t, err)
	}
}

func TestFSWatcher_Watch(t *testing.T) {
	//init
	c := mkdir(t, "c")
	d := mkdir(t, "d")
	e := mkdir(t, "e")

	var changedFile string
	var ch = make(chan bool)
	var onChange = func(e Event) error {
		fsEvent, ok := e.(*FSEvnent)
		assert.True(t, ok)
		changedFile = fsEvent.Filename
		ch <- true
		return nil
	}

	cur, err := filepath.Abs(".")
	assert.NoError(t, err)

	w, err := NewFSWatcher("testdata/c/*.txt,testdata/d/*.conf,testdata/e", onChange)
	assert.NoError(t, err)
	assert.Equal(t, []string{filepath.FromSlash(filepath.Join(cur, "testdata/c/*.txt")), filepath.FromSlash(filepath.Join(cur, "testdata/d/*.conf"))}, w.fuzzyFiles)
	go w.Watch()

	changedFile = ""
	time.Sleep(1 * time.Second)
	write2file(t, filepath.Join("testdata", "c", "c.txt"), "")
	<-ch
	assert.Equal(t, filepath.Join(cur, "testdata", "c", "c.txt"), changedFile)

	changedFile = ""
	time.Sleep(1 * time.Second)
	write2file(t, filepath.Join("testdata", "c", "c.txt"), "abc")
	<-ch
	assert.Equal(t, filepath.Join(cur, "testdata", "c", "c.txt"), changedFile)

	changedFile = ""
	time.Sleep(1 * time.Second)
	write2file(t, filepath.Join("testdata", "d", "d.conf"), "")
	<-ch
	assert.Equal(t, filepath.Join(cur, "testdata", "d", "d.conf"), changedFile)

	changedFile = ""
	time.Sleep(1 * time.Second)
	write2file(t, filepath.Join("testdata", "e", "e.conf"), "")
	<-ch
	assert.Equal(t, filepath.Join("testdata", "e", "e.conf"), changedFile)

	remove(t, c)
	remove(t, d)
	remove(t, e)
}
