package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

const modulePath = "github.com/jsternberg/flux-lang/parser"

// Package is the package file from go list.
// This is only annotated with the fields I need and is
// non-exhaustive.
type Package struct {
	Dir string
	ImportPath string
}

func listPackage(path string) (*Package, error) {
	cmd := exec.Command("go", "list", "-json", path)
	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var pkg Package
	dec := json.NewDecoder(rc)
	if err := dec.Decode(&pkg); err != nil {
		// Ensure stdout is closed so the command doesn't wait
		// on writing to it and the program finishes.
		_ = rc.Close()
		_ = cmd.Wait()
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// todo: I think this is needed, but trying it out without doing this first.
func makeDirWriteable(dir string) (reset func() error, err error) {
	// Walk through this directory and mark every
	// folder that we make writeable so we can undo this.
	markedDirs := make(map[string]os.FileMode)
	reset = func() error {
		var firstErr error
		for dir, mode := range markedDirs {
			if err := os.Chmod(dir, mode); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		fullpath := filepath.Join(path, info.Name())
		mode := info.Mode()
		if mode & 0200 == 0 {
			if err := os.Chmod(fullpath, mode | 0200); err != nil {
				return err
			}
		}
		markedDirs[fullpath] = mode
		return nil
	}); err != nil {
		_ = reset()
		panic(err)
	}
	return reset, nil
}

func main() {
	// Find the directory where this module is located.
	pkg, err := listPackage(modulePath)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("cargo", "build", "--release")
	cmd.Dir = pkg.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
