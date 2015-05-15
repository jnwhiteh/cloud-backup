package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// File provides a subset of the *os.File struct interface
type File interface {
	io.Closer
	io.Reader
	Stat() (os.FileInfo, error)
}

type FileOpener func(string) (File, error)
type Globber func(string) ([]string, error)

// LocalFilesystem implements the Filer interface for the local file system
type LocalFilesystem struct {
	hasher  hash.Hash
	opener  FileOpener
	globber Globber
}

var OSOpener = func(name string) (File, error) {
	return os.Open(name)
}

func NewDefaultLocalFilesystem() LocalFilesystem {
	return NewLocalFilesystem(nil, nil, nil)
}

func NewLocalFilesystem(hasher hash.Hash, opener FileOpener, globber Globber) LocalFilesystem {
	// Set some sane default values
	if hasher == nil {
		hasher = md5.New()
	}
	if opener == nil {
		opener = OSOpener
	}
	if globber == nil {
		globber = filepath.Glob
	}

	return LocalFilesystem{hasher, opener, globber}
}

func (f *LocalFilesystem) Files(path string) ([]HashedFile, error) {
	matches, err := f.globber(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}

	var results []HashedFile
	for _, match := range matches {
		// skip any hidden files
		if strings.HasPrefix(match, ".") {
			continue
		}

		hash, err := f.Hash(match)
		if err == errIsDirectory {
			// we skip directories
			continue
		}

		if err != nil {
			return nil, err
		}

		results = append(results, HashedFile{
			Folder:   filepath.Dir(match),
			Filename: filepath.Base(match),
			Hash:     hash,
		})
	}

	return results, nil
}

var errIsDirectory = errors.New("File is a directory")

func (f *LocalFilesystem) Hash(path string) (string, error) {
	file, err := f.opener(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	// skip directories by returning an error
	if stat.IsDir() {
		return "", errIsDirectory
	}

	f.hasher.Reset()
	io.Copy(f.hasher, file)
	return fmt.Sprintf("%x", f.hasher.Sum(nil)), nil
}
