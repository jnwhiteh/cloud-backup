package sync

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"reflect"
	"testing"
)

type mockFS struct {
	files  map[string][]File
	errors map[string]error
}

func CreateMock(path string, names ...string) *mockFS {
	fs := &mockFS{
		make(map[string][]File),
		make(map[string]error),
	}
	fs.addFiles(path, names...)
	return fs
}

func (a mockFS) Mkdir(path string) error {
	err, _ := a.errors[path]
	if err != nil {
		return err
	}
	_, ok := a.files[path]
	if !ok {
		a.files[path] = nil
	}
	return nil
}

func (a mockFS) Files(path string) ([]File, error) {
	err, _ := a.errors[path]
	if err != nil {
		return nil, err
	}

	files, ok := a.files[path]
	if !ok {
		log.Fatalf("Tried to get invalid folder key %s from mock", path)
	}
	return files, nil
}

func (a mockFS) addFiles(path string, names ...string) {
	for _, name := range names {
		h := md5.New()
		io.WriteString(h, name)
		hash := fmt.Sprintf("%x", h.Sum(nil))
		a.files[path] = append(a.files[path], File{path, name, hash})
	}
}

func TestNormalCases(t *testing.T) {
	type testCase struct {
		path         string
		local        FilestoreAPI
		remote       FilestoreAPI
		expected     []status
		expected_err error
	}

	testCases := []testCase{
		// a clean remote should results in needing sync
		testCase{
			path:     "pics/foo",
			local:    CreateMock("pics/foo", "a", "b", "c"),
			remote:   CreateMock("pics/foo"),
			expected: []status{STATUS_NEED_SYNC, STATUS_NEED_SYNC, STATUS_NEED_SYNC},
		},
		// dirty remote should result in an error
		testCase{
			path:         "pics/foo",
			local:        CreateMock("pics/foo", "a"),
			remote:       CreateMock("pics/foo", "z"),
			expected_err: ERR_REMOTE_NOT_CLEAN,
		},
	}

	for _, test := range testCases {
		runTestCase(t, test.path, test.local, test.remote, test.expected, test.expected_err)
	}
}

func TestNoLocalHash(t *testing.T) {
	local := CreateMock("pics/foo", "a", "b")
	remote := CreateMock("pics/foo")

	local.files["pics/foo"][0].Hash = ""
	runTestCase(t, "pics/foo", local, remote, nil, ERR_LOCAL_NO_HASH)
}

func TestRemoteHashMismatch(t *testing.T) {
	local := CreateMock("pics/foo", "a")
	remote := CreateMock("pics/foo", "a")
	remote.files["pics/foo"][0].Hash = "wronghash"
	runTestCase(t, "pics/foo", local, remote, nil, ERR_REMOTE_NOT_CLEAN)
}

func runTestCase(t *testing.T, path string, local, remote FilestoreAPI, expected []status, expected_err error) {
	syncer := Syncer{local, remote}
	files, err := syncer.SyncFolder(path, path)
	if expected_err != nil && err != expected_err {
		t.Fatalf("Expected error %s, got %s", expected_err, err)
	}

	var result []status
	for _, file := range files {
		result = append(result, file.Status)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("status mismatch: expected %#v, got %#v", expected, result)
	}
}
