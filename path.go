package main

import (
	"os"
	"path/filepath"
	"strings"
)

func CleanFuzzyPath(paths []string) ([]string, error) {
	return cleanPaths(paths, func(i, j int) (int, error) {
		matched, err := filepath.Match(paths[i], paths[j])
		if err != nil {
			return -1, err
		}
		if matched {
			return j, nil
		}
		matched, err = filepath.Match(paths[j], paths[i])
		if err != nil {
			return -1, err
		}
		if matched {
			return i, nil
		}
		return -1, nil
	})
}

func ParentDir(paths ...string) ([]string, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	var err error
	var exactPaths = make([]string, len(paths))
	for i, path := range paths {
		if !filepath.IsAbs(path) {
			path, err = filepath.Abs(path)
			if err != nil {
				return nil, err
			}
		}
		exactPaths[i] = exactDir(path)
	}
	return cleanPaths(exactPaths, func(i, j int) (int, error) {
		if strings.HasPrefix(exactPaths[i], exactPaths[j]) {
			return i, nil
		}
		if strings.HasPrefix(exactPaths[j], exactPaths[i]) {
			return j, nil
		}
		return -1, nil
	})
}

func cleanPaths(paths []string, cleanFunc func(i, j int) (int, error)) ([]string, error) {
	var cleanMap = map[string]bool{}
	for i := range paths {
		cleanMap[paths[i]] = true
	}
	for i := range paths {
		if _, ok := cleanMap[paths[i]]; !ok {
			continue
		}
		for j := i + 1; j < len(paths); j++ {
			index, err := cleanFunc(i, j)
			if err != nil {
				return nil, err
			}
			if index >= 0 {
				delete(cleanMap, paths[index])
			}
		}
	}

	var clean []string
	for k := range cleanMap {
		clean = append(clean, k)
	}
	return clean, nil
}

func exactDir(path string) string {
	var dirs []string
	if strings.HasPrefix(path, "/") {
		dirs = append(dirs, "/")
	}
	for _, p := range strings.Split(path, "/") {
		dir := filepath.Join(filepath.Join(dirs...), p)
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			break
		}
		dirs = append(dirs, p)
	}
	return filepath.FromSlash(filepath.Join(dirs...))
}

func GroupByWildcard(paths []string) (exact []string, fuzzy []string, err error) {
	for i := range paths {
		if isFuzzyPath(paths[i]) {
			var abs string
			abs, err = filepath.Abs(paths[i])
			if err != nil {
				return 
			}
			fuzzy = append(fuzzy, filepath.FromSlash(abs))
		} else {
			exact = append(exact, filepath.FromSlash(paths[i]))
		}
	}
	return
}

func isFuzzyPath(path string) bool {
	return strings.ContainsAny(path, "[]*?^")
}
