package main_test

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/jnwhiteh/cloud-backup"
)

type mockFS struct {
	files  map[string][]main.HashedFile
	errors map[string]error
}

func CreateMock(path string, names ...string) *mockFS {
	fs := &mockFS{
		make(map[string][]main.HashedFile),
		make(map[string]error),
	}
	fs.addFiles(path, names...)
	return fs
}

func (a mockFS) addFiles(path string, names ...string) {
	// Ensure the path is present in the map
	a.files[path] = a.files[path]
	for _, name := range names {
		h := md5.New()
		io.WriteString(h, name)
		hash := fmt.Sprintf("%x", h.Sum(nil))
		a.files[path] = append(a.files[path], main.HashedFile{path, name, hash})
	}
}

func (a mockFS) Files(path string) ([]main.HashedFile, error) {
	err, _ := a.errors[path]
	if err != nil {
		return nil, err
	}

	files, ok := a.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return files, nil
}

func TestNormalCases(t *testing.T) {
	type testCase struct {
		path         string
		local        main.Filer
		remote       main.Filer
		expected     []main.Status
		expected_err error
	}

	testCases := []testCase{
		// a clean remote should results in needing sync
		testCase{
			path:     "pics/foo",
			local:    CreateMock("pics/foo", "a", "b", "c"),
			remote:   CreateMock("pics/foo"),
			expected: []main.Status{main.STATUS_NEED_SYNC, main.STATUS_NEED_SYNC, main.STATUS_NEED_SYNC},
		},
		// dirty remote should result in an error
		testCase{
			path:         "pics/foo",
			local:        CreateMock("pics/foo", "a"),
			remote:       CreateMock("pics/foo", "z"),
			expected_err: main.ERR_REMOTE_NOT_CLEAN,
		},
	}

	for idx, test := range testCases {
		t.Logf("Running test %d: %v", idx, test)
		runTestCase(t, test.path, test.local, test.remote, test.expected, test.expected_err)
	}
}

func TestNoLocalHash(t *testing.T) {
	local := CreateMock("pics/foo", "a", "b")
	remote := CreateMock("pics/foo")

	local.files["pics/foo"][0].Hash = ""
	runTestCase(t, "pics/foo", local, remote, nil, main.ERR_LOCAL_NO_HASH)
}

func TestRemoteHashMismatch(t *testing.T) {
	local := CreateMock("pics/foo", "a")
	remote := CreateMock("pics/foo", "a")
	remote.files["pics/foo"][0].Hash = "wronghash"
	runTestCase(t, "pics/foo", local, remote, nil, main.ERR_REMOTE_NOT_CLEAN)
}

func runTestCase(t *testing.T, path string, local, remote main.Filer, expected []main.Status, expected_err error) {
	syncer := main.NewSyncer(local, remote)
	files, err := syncer.SyncStatus(path, path)
	if expected_err != nil && err != expected_err {
		t.Fatalf("Expected error %s, got %s", expected_err, err)
	}

	var result []main.Status
	for _, file := range files {
		result = append(result, file.Status)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("status mismatch: expected %#v, got %#v", expected, result)
	}
}
