package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fakeBinaryDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("membuat binary palsu %s: %v", name, err)
		}
	}
	return dir
}

func TestRunGofmtMissingBinary(t *testing.T) {
	t.Setenv("PATH", "")
	if err := runGofmt([]string{"contoh.go"}); err == nil {
		t.Error("runGofmt() ingin error saat gofmt tidak ditemukan")
	}
}

func TestRunGofmtSuccess(t *testing.T) {
	t.Setenv("PATH", fakeBinaryDir(t, "gofmt")+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := runGofmt([]string{"contoh.go"}); err != nil {
		t.Errorf("runGofmt() error = %v", err)
	}
}

func TestRunGofmtWithoutGoFiles(t *testing.T) {
	if err := runGofmt([]string{"contoh.txt"}); err != nil {
		t.Errorf("runGofmt() error = %v untuk file non-Go", err)
	}
}

func TestRunBuildMissingBinary(t *testing.T) {
	t.Setenv("PATH", "")
	if err := runBuild(); err == nil {
		t.Error("runBuild() ingin error saat go tidak ditemukan")
	}
}

func TestRunBuildSuccess(t *testing.T) {
	t.Setenv("PATH", fakeBinaryDir(t, "go")+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := runBuild(); err != nil {
		t.Errorf("runBuild() error = %v", err)
	}
}
