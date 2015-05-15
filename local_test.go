package main_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/jnwhiteh/cloud-backup"
)

// mockFile implements the main.File interface
type mockFile struct {
	*bytes.Buffer        // the buffer that contains the contents
	name          string // the ostensible "name" of the file
	isDir         bool   // is the file a directory
}

func fakeFile(path string) *mockFile {
	return &mockFile{
		bytes.NewBufferString(fmt.Sprintf("contents:%s", path)),
		filepath.Base(path),
		false,
	}
}

func fakeDir(path string) *mockFile {
	return &mockFile{bytes.NewBuffer(nil), path, true}
}

func (f *mockFile) Close() error {
	return nil
}

func (f *mockFile) Stat() (os.FileInfo, error) {
	return &mockFileInfo{
		name:  f.name,
		size:  int64(f.Len()),
		isDir: f.isDir,
	}, nil
}

// mockFileInfo implements the os.FileInfo interface
type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (fi mockFileInfo) Name() string {
	return fi.name
}

func (fi mockFileInfo) Size() int64 {
	return fi.size
}

func (fi mockFileInfo) Mode() os.FileMode {
	if fi.isDir {
		return 0755
	} else {
		return 0644
	}
}

func (fi mockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi mockFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi mockFileInfo) Sys() interface{} {
	return nil
}

func TestNormal(t *testing.T) {
	// return two files
	globber := func(path string) ([]string, error) {
		return []string{"foo", "bar"}, nil
	}

	// fake some contents
	opener := func(path string) (main.File, error) {
		return fakeFile(path), nil
	}

	fs := main.NewLocalFilesystem(nil, opener, globber)
	files, err := fs.Files(".")
	if err != nil {
		log.Fatalf("Error getting files: %s", err)
	}
	expected := []main.HashedFile{
		{".", "foo", "5172373545499a04bc8a03681dc9ab55"},
		{".", "bar", "29228460db10bba1415bd0106de0e974"},
	}

	ok := reflect.DeepEqual(expected, files)
	if !ok {
		t.Errorf("Result did not match expected: got %v, expected %v", files, expected)
	}
}

func TestIgnoreHiddenFiles(t *testing.T) {
	globber := func(path string) ([]string, error) {
		return []string{".hidden", "foo", "bar"}, nil
	}
	opener := func(path string) (main.File, error) {
		if path == ".hidden" {
			return fakeDir(path), nil
		} else {
			return fakeFile(path), nil
		}
	}
	fs := main.NewLocalFilesystem(nil, opener, globber)
	files, err := fs.Files(".")
	if err != nil {
		log.Fatalf("Error getting files: %s", err)
	}
	if len(files) != 2 {
		t.Errorf("Unexpected length of file list: got %d, expected 2", len(files))
	}
	for _, file := range files {
		if file.Filename == ".hidden" {
			t.Errorf("Hidden file was included in file list")
		}
	}
}

func TestDontRecurse(t *testing.T) {
	globber := func(path string) ([]string, error) {
		return []string{"subdir", "foo", "bar"}, nil
	}
	opener := func(path string) (main.File, error) {
		if path == "subdir" {
			return fakeDir(path), nil
		} else {
			return fakeFile(path), nil
		}
	}
	fs := main.NewLocalFilesystem(nil, opener, globber)
	files, err := fs.Files(".")
	if err != nil {
		log.Fatalf("Error gettingn files: %s", err)
	}
	if len(files) != 2 {
		t.Errorf("Unexpected length of file list: got %d, expected 2", len(files))
	}
	for _, file := range files {
		if file.Filename == "subdir" {
			t.Errorf("Subdirectory was included in file list")
		}
	}
}
