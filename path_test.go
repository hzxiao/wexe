package main

import (
	"path/filepath"
	"testing"

	"github.com/hzxiao/goutil/assert"
)

func TestGroupByWildcard(t *testing.T) {
	cur, err := filepath.Abs(".")
	assert.NoError(t, err)

	exact, fuzzy, err := GroupByWildcard([]string{"wexe", "wexe/*.go"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"wexe"}, exact)
	assert.Equal(t, []string{filepath.FromSlash(filepath.Join(cur, "wexe/*.go"))}, fuzzy)
}

func TestCleanFuzzyPath(t *testing.T) {
	var tables = []struct {
		Paths  []string
		Cleans []string
	}{
		{
			Paths:  []string{"wexe"},
			Cleans: []string{"wexe"},
		},
		{
			Paths:  []string{"wexe/*.go", "wexe/go.sum", "wexe/wexe.go"},
			Cleans: []string{"wexe/*.go", "wexe/go.sum"},
		},
	}

	for i := range tables {
		cleans, err := CleanFuzzyPath(tables[i].Paths)
		assert.NoError(t, err)
		assert.Equal(t, tables[i].Cleans, cleans)
	}
}

func TestParentDir(t *testing.T) {
	cur, err := filepath.Abs(".")
	assert.NoError(t, err)

	var join = func(paths ...string) string {
		return filepath.FromSlash(filepath.Join(paths...))
	}

	var tables = []struct {
		Paths []string
		Dirs  []string
	}{
		{
			Paths: []string{"./*.go"},
			Dirs:  []string{join(cur)},
		},
		{
			Paths: []string{"testdata/a/*.txt", "testdata/a/aa/*.txt", "testdata/b/*.go"},
			Dirs:  []string{join(cur, "testdata", "a"), join(cur, "testdata", "b")},
		},
	}

	for i := range tables {
		dirs, err := ParentDir(tables[i].Paths...)
		assert.NoError(t, err)
		assert.Equal(t, tables[i].Dirs, dirs)
	}
}
