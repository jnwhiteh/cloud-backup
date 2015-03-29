package sync

import (
	"fmt"
	"sort"
)

type File struct {
	Folder   string // the path to the parent folder
	Filename string
	Hash     string // a hex digest of the file contents
}

type byName []File

func (a byName) Len() int {
	return len(a)
}
func (a byName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byName) Less(i, j int) bool {
	return a[i].Filename < a[j].Filename
}

type FilestoreAPI interface {
	// Return a list of files in the given path/folder
	Files(path string) ([]File, error)
	// Create a folder at the remote path, if it doesn't exist
	Mkdir(path string) error
}

type status string

var (
	STATUS_ALREADY       status = "Already synchronized"
	STATUS_NEED_SYNC     status = "Needs sync"
	STATUS_UPLOADED      status = "Uploaded"
	ERR_REMOTE_NOT_CLEAN error  = fmt.Errorf("Remote folder is not clean")
	ERR_LOCAL_NO_HASH    error  = fmt.Errorf("Local file has no hash")
)

type SyncStatus struct {
	File
	Status status
	Error  error
}

type Syncer struct {
	local  FilestoreAPI
	remote FilestoreAPI
}

func (s Syncer) SyncFolder(localPath, remotePath string) ([]*SyncStatus, error) {
	// verify that the local folder exists
	localFiles, err := s.local.Files(localPath)
	if err != nil {
		return nil, err
	}

	for _, file := range localFiles {
		if file.Hash == "" {
			return nil, ERR_LOCAL_NO_HASH
		}
	}

	// check if the remote folder exists
	err = s.remote.Mkdir(remotePath)
	if err != nil {
		return nil, err
	}

	// fetch files from the remote folder
	remoteFiles, err := s.remote.Files(remotePath)
	if err != nil {
		return nil, err
	}

	return s.Worklist(localFiles, remoteFiles)
}

func (s Syncer) Worklist(localFiles, remoteFiles []File) ([]*SyncStatus, error) {
	sort.Sort(byName(localFiles))
	sort.Sort(byName(remoteFiles))

	var files []*SyncStatus
	var localIdx = 0
	var remoteIdx = 0
	for localIdx < len(localFiles) || remoteIdx < len(remoteFiles) {
		if localIdx >= len(localFiles) {
			// no more local files, something's wrong here
			return nil, ERR_REMOTE_NOT_CLEAN
		} else if remoteIdx >= len(remoteFiles) {
			// any more local files can be uploaded
			files = s.addWithStatus(files, localFiles[localIdx], STATUS_NEED_SYNC)
			localIdx++
			continue
		}

		local := localFiles[localIdx]
		remote := remoteFiles[remoteIdx]

		if local.Filename == remote.Filename {
			if local.Hash == remote.Hash {
				files = s.addWithStatus(files, local, STATUS_ALREADY)
				localIdx++
				remoteIdx++
			} else {
				return nil, ERR_REMOTE_NOT_CLEAN
			}
		} else if local.Filename < remote.Filename {
			files = s.addWithStatus(files, local, STATUS_NEED_SYNC)
			localIdx++
		} else if local.Filename > remote.Filename {
			// Something on the remote we don't know about
			return nil, ERR_REMOTE_NOT_CLEAN
		}
	}

	return files, nil
}

func (s Syncer) addWithStatus(files []*SyncStatus, file File, status status) []*SyncStatus {
	return append(files, &SyncStatus{
		File:   file,
		Status: status,
	})
}
