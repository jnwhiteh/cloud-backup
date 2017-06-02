package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
)

var (
	secretFile   = flag.String("secret_file", "client_secrets.json", "client secrets JSON file")
	redirectHost = flag.String("redirect_host", "localtest.me", "host to redirect with oauth success")
	redirectPort = flag.String("redirect_port", "31337", "host to redirect with oauth success")
	debug        = flag.Bool("debug", true, "show HTTP traffic")
	localFolder  = flag.String("local", "", "path of a local folder to synchronize")
	remoteFolder = flag.String("remote", "", "path of the destination remote folder")
)

func main() {
	flag.Parse()
	config := OAuthConfigFromFile(*secretFile, []string{"wl.signin", "wl.offline_access", "onedrive.readwrite"})
	client := OAuthClient("onedrive-sync", *debug, config)
	api := OneDriveAPI{client, "https://api.onedrive.com/v1.0"}

	quota, err := api.Quota()
	if err != nil {
		log.Fatalf("Error fetching quota: %s", err)
	}

	log.Printf("Connected to %s's drive (%s of %s available)",
		quota.Owner.User.DisplayName,
		humanize.Bytes(uint64(quota.Quota.Remaining)),
		humanize.Bytes(uint64(quota.Quota.Total)))

	// ensure destination folder exists
	meta, err := api.Metadata(*remoteFolder)
	if err != nil && err == PathNotFound {
		parent := filepath.Dir(*remoteFolder)
		child := filepath.Base(*remoteFolder)

		resp, err := api.Mkdir(parent, child)
		if err != nil {
			log.Fatalf("Failed when creating folder: %s", err)
		}
		log.Printf("Mkdir response: %s", resp)
		meta, err = api.Metadata(*remoteFolder)
	} else if err != nil {
		log.Fatalf("Could not locate remote folder: %s", err)
	} else if meta.Folder == nil {
		log.Fatalf("Remote path is not a folder")
	}

	// fetch remote path recursive metadata
	remoteFiles, err := api.ChildHashes(*remoteFolder)
	if err != nil {
		log.Fatalf("Failed when fetching remote file hashes: %s", err)
	}

	// fetch local file hashes
	localFiles, err := LocalFileHashes(*localFolder)
	if err != nil {
		log.Fatalf("Failed when fetching local file hashes: %s", err)
	}

	tree, filenames := MergeTrees(localFiles, remoteFiles)

	// make sure we don't have any files that are only on the remote
	for filename, entry := range tree {
		if entry.LocalHash == "" {
			log.Fatalf("File %s is on remote but not local", filename)
		} else if entry.LocalHash != entry.RemoteHash && entry.RemoteHash != "" {
			log.Fatalf("File %s has different hashes (local: %s, remote: %s)",
				filename, entry.LocalHash, entry.RemoteHash)
		}
	}

	for _, file := range filenames {
		entry := tree[file]
		if entry.LocalHash == entry.RemoteHash {
			log.Printf("Skipping %s, already uploaded", file)
		} else {
			log.Printf("Hash mismatch (local: %s, remote: %s)", entry.LocalHash, entry.RemoteHash)
			log.Printf("Uploading %s...", file)
			body, err := api.Upload(filepath.Join(*localFolder, file),
				filepath.Join(*remoteFolder, file))
			log.Printf("Response: %s", body)
			if err != nil {
				log.Fatalf("Error uploading file: %s", err)
			}
		}
	}
}

func LocalFileHashes(path string) ([]FileHash, error) {
	// build a list of local files to check
	filenames, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}

	var result []FileHash
	for _, file := range filenames {
		filename := filepath.Base(file)
		if strings.HasPrefix(filename, ".") {
			continue
		}
		hash, err := Sha1Hash(file)
		if err != nil {
			return nil, err
		}

		result = append(result, FileHash{
			Name: filename,
			Hash: hash,
		})
	}
	return result, nil
}

func Sha1Hash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sha1er := sha1.New()
	io.Copy(sha1er, file)
	hash := fmt.Sprintf("%x", sha1er.Sum(nil))
	return hash, nil
}

type TreeHash struct {
	Name       string
	LocalHash  string
	RemoteHash string
}

func MergeTrees(localFiles, remoteFiles []FileHash) (map[string]*TreeHash, []string) {
	tree := make(map[string]*TreeHash)
	var filenames []string

	for _, file := range localFiles {
		entry, ok := tree[file.Name]
		if !ok {
			tree[file.Name] = &TreeHash{
				Name:      file.Name,
				LocalHash: file.Hash,
			}
			filenames = append(filenames, file.Name)
		} else {
			entry.LocalHash = file.Hash
		}
	}

	for _, file := range remoteFiles {
		entry, ok := tree[file.Name]
		if !ok {
			tree[file.Name] = &TreeHash{
				Name:       file.Name,
				RemoteHash: file.Hash,
			}
			filenames = append(filenames, file.Name)
		} else {
			entry.RemoteHash = file.Hash
		}
	}

	sort.Strings(filenames)
	return tree, filenames
}
