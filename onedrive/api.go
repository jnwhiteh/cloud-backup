package main

//go:generate go run tools/generate.go resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	PathNotFound = fmt.Errorf("PathNotFound")
)

type OneDriveAPI struct {
	client  *http.Client
	baseURL string
}

func (api *OneDriveAPI) Quota() (*Drive, error) {
	var response Drive
	resp, err := api.client.Get(api.baseURL + "/drive")
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	return &response, err
}

func (api *OneDriveAPI) Metadata(path string) (*Item, error) {
	endpoint := &url.URL{
		Path:     api.baseURL + "/drive/root:/" + path,
		RawQuery: "select=id,folder,file",
	}
	resp, err := api.client.Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, PathNotFound
	}

	var response Item
	err = json.NewDecoder(resp.Body).Decode(&response)
	return &response, err
}

type FileHash struct {
	Name string
	Hash string
}

type ByName []FileHash

func (a ByName) Len() int {
	return len(a)
}
func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

func (api *OneDriveAPI) ChildHashes(folderPath string) ([]FileHash, error) {
	endpoint := &url.URL{
		Path:     api.baseURL + "/drive/root:/" + folderPath + ":/children",
		RawQuery: "select=id,name,folder,file",
	}
	resp, err := api.client.Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, PathNotFound
	}

	var response ViewChanges
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	var result []FileHash
	for response.Value != nil && len(response.Value) > 0 {
		for _, metadata := range response.Value {
			if metadata.Folder != nil {
				log.Fatalf("Cannot handle subdirectories yet")
			}

			result = append(result, FileHash{
				metadata.Name,
				strings.ToLower(metadata.File.Hashes.Sha1Hash),
			})
		}

		if response.Instanceodata_nextLink != "" {
			var nextLink = response.Instanceodata_nextLink
			log.Printf("Collected %d results, fetching next page: %s", len(result), nextLink)
			resp, err := api.client.Get(nextLink)
			if err != nil {
				log.Fatalf("Failed when fetching %s: %s", nextLink, err)
			}

			response = ViewChanges{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				log.Fatalf("Failed decoding next page response: %s", err)
			}
		} else {
			break
		}
	}

	return result, nil
}

func (api *OneDriveAPI) Upload(filename, remotePath string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	endpoint := &url.URL{
		Path: api.baseURL + "/drive/root:/" + remotePath + ":/content",
	}
	sreader := &SpeedReader{file: file, start: time.Now()}

	req, err := http.NewRequest("PUT", endpoint.String(), sreader)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := api.client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", body), nil
}

type SpeedReader struct {
	file  *os.File
	start time.Time
	bytes uint64
	last  uint64
}

func (r *SpeedReader) Read(p []byte) (int, error) {
	n, err := r.file.Read(p)
	r.bytes += uint64(n)
	duration := uint64(time.Now().Sub(r.start).Seconds())
	if duration > 0 && duration != r.last {
		r.last = duration
		log.Printf("... %s /s (%s sent)", humanize.Bytes(r.bytes/duration), humanize.Bytes(r.bytes))
	}
	return n, err
}
func (r *SpeedReader) Close() error {
	return r.file.Close()
}

func (api *OneDriveAPI) Mkdir(parent, name string) (string, error) {
	endpoint := &url.URL{
		Path: api.baseURL + "/drive/root:/" + parent + ":/children",
	}
	type folderPayload struct {
		Name   string   `json:"name"`
		Folder struct{} `json:"folder"`
	}
	payload := folderPayload{Name: name}
	body := getIndentedJSON(payload)
	resp, err := api.client.Post(endpoint.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", body), nil
}
