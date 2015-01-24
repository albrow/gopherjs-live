package main

import (
	"errors"
	"fmt"
	"github.com/OneOfOne/xxhash/native"
	"github.com/howeyc/fsnotify"
	"github.com/wsxiaoys/terminal/color"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	workingDir string
)

// a map of known file hashes. This is required to
// determine whether a file actually changed. It is a workaround
// to fix the bug that occurs when a text editor uses atomic saves,
// which triggers multiple watch events even though the file was
// only saved once.
var fileHashes = map[string][]byte{}

func main() {
	defer recovery()

	// Get working directory
	var err error
	workingDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	done := make(chan bool)

	// Define how to process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if changed, err := fileDidChange(ev.Name); err != nil {
					chimeError(err)
				} else if changed {
					if filepath.Base(ev.Name)[0] == '.' {
						// Ignore hidden system files
						continue
					}
					if filepath.Ext(ev.Name) != ".go" {
						// Ignore anything that's not a go file
						continue
					}
					color.Printf("@gCHANGE: %s\n", ev.Name)
					if err := recompile(); err != nil {
						chimeError(err)
					}
				}
			case err := <-watcher.Error:
				chimeError(err)
			}
		}
	}()

	// Recursively calculate paths to watch
	pathsToWatch, err := getPaths()
	if err != nil {
		panic(err)
	}
	for _, path := range pathsToWatch {
		if err := watcher.Watch(path); err != nil {
			panic(err)
		}
	}
	color.Println("@cWatching for changes...")

	// Don't exit early
	<-done
	watcher.Close()
}

func recompile() error {
	// Execute the gopherjs command with
	args := []string{"build"}
	// Pass through all the args/flags
	args = append(args, os.Args[1:]...)
	cmd := exec.Command("gopherjs", args...)
	fmt.Println("Recompiling...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) != 0 {
			return errors.New(string(out))
		} else {
			return err
		}
	} else if len(out) != 0 {
		return errors.New(string(out))
	}
	return nil
}

func getPaths() ([]string, error) {
	watched := []string{}
	if err := filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name()[0] == '.' {
			// ignore hidden system files
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			watched = append(watched, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return watched, nil
}

func recovery() {
	if err := recover(); err != nil {
		// convert err to a string
		var errMsg string
		if e, ok := err.(error); ok {
			errMsg = e.Error()
		} else {
			errMsg = fmt.Sprint(err)
		}
		chimeError(errMsg)
	}
}

// chimeError outputs the bell character and then the error message,
// colored red and formatted.
func chimeError(err interface{}) {
	fmt.Print("\a")
	color.Printf("@rERROR: %s\n", err)
}

// chimeErrorf outputs the bell character and then the error message,
// colored red and formatted according to format and args. It works
// just like fmt.Printf.
func chimeErrorf(format string, args ...interface{}) {
	fmt.Print("\a")
	color.Printf("@rERROR: %s\n", fmt.Sprintf(format, args...))
}

// fileDidChange uses the last known hash to determine whether or
// not the file actually changed. It solves the problem of false positives
// coming from fsnotify when used with a text editor that uses atomic saves.
func fileDidChange(path string) (bool, error) {
	if hash, found := fileHashes[path]; !found {
		// we have not hashed the file before.
		// hash it now and store the value
		if newHash, exists, err := calculateHashForPath(path); err != nil {
			return false, err
		} else if exists {
			fileHashes[path] = newHash
		}
		return true, nil
	} else {
		if newHash, exists, err := calculateHashForPath(path); err != nil {
			return false, err
		} else if !exists {
			// if the file no longer exists, it has been deleted
			// we should consider that a change and recompile
			delete(fileHashes, path)
			return true, nil
		} else if string(newHash) != string(hash) {
			// if the file does exist and has a different hash, there
			// was an actual change and we should recompile
			fileHashes[path] = newHash
			return true, nil
		}
		return false, nil
	}
}

// calculateHashForPath calculates a hash for the file at the given path.
// If the file does not exist, the second return value will be false.
func calculateHashForPath(path string) ([]byte, bool, error) {
	h := xxhash.New64()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}
	io.Copy(h, f)
	result := h.Sum(nil)
	return result, true, nil
}
