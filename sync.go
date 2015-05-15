package main

import (
	"fmt"
	"sort"
)

type HashedFile struct {
	Folder   string // the path to the parent folder
	Filename string
	Hash     string // a hex digest of the file contents
}

type byName []HashedFile

func (a byName) Len() int {
	return len(a)
}
func (a byName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byName) Less(i, j int) bool {
	return a[i].Filename < a[j].Filename
}

type Filer interface {
	// Return a list of files in the given path/folder
	Files(path string) ([]HashedFile, error)
}

type Status string

var (
	STATUS_ALREADY       Status = "Already synchronized"
	STATUS_NEED_SYNC     Status = "Needs sync"
	STATUS_UPLOADED      Status = "Uploaded"
	ERR_REMOTE_NOT_CLEAN error  = fmt.Errorf("Remote folder is not clean")
	ERR_LOCAL_NO_HASH    error  = fmt.Errorf("Local file has no hash")
)

type SyncStatus struct {
	HashedFile
	Status Status
	Error  error
}

type Syncer struct {
	local  Filer
	remote Filer
}

func NewSyncer(local Filer, remote Filer) Syncer {
	return Syncer{local, remote}
}

func (s Syncer) SyncStatus(localPath, remotePath string) ([]*SyncStatus, error) {
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

	// fetch files from the remote folder
	remoteFiles, err := s.remote.Files(remotePath)
	if err != nil {
		return nil, err
	}

	return s.Worklist(localFiles, remoteFiles)
}

func (s Syncer) Worklist(localFiles, remoteFiles []HashedFile) ([]*SyncStatus, error) {
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

func (s Syncer) addWithStatus(files []*SyncStatus, file HashedFile, status Status) []*SyncStatus {
	return append(files, &SyncStatus{
		HashedFile: file,
		Status:     status,
	})
}
