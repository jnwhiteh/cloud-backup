package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type ClientSecrets struct {
	Installed struct {
		Client_id     string `json:"client_id"`
		Client_secret string `json:"client_secret"`
		Auth_uri      string `json:"auth_uri"`
		Token_uri     string `json:"token_uri"`
	} `json:"installed"`
}

func OAuthConfigFromFile(filename string, scopes []string) *oauth2.Config {
	// Fetch the secrets from the JSON file
	var secrets ClientSecrets
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading %q: %v", filename, err)
	}
	err = json.Unmarshal(contents, &secrets)
	if err != nil {
		log.Fatalf("Could not decode client credentials: %v", err)
	}

	return &oauth2.Config{
		ClientID:     secrets.Installed.Client_id,
		ClientSecret: secrets.Installed.Client_secret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  secrets.Installed.Auth_uri,
			TokenURL: secrets.Installed.Token_uri,
		},
	}
}

// tokenCacheFilename returns the local cache filename for a given oauth
// configuration tuple (id, secret, scope).
func tokenCacheFilename(appName string, config *oauth2.Config) string {
	hash := fnv.New32a()
	hash.Write([]byte(config.ClientID))
	hash.Write([]byte(config.ClientSecret))
	scopes := strings.Join(config.Scopes, ",")
	hash.Write([]byte(scopes))
	fn := fmt.Sprintf("%s-token%v", appName, hash.Sum32())
	return filepath.Join(osUserCacheDir(), url.QueryEscape(fn))
}

// osUserCacheDir returns the cache directory used for oAuth tokens depending
// on operating system.
func osUserCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	log.Printf("TODO: osUserCacheDir on GOOS %q", runtime.GOOS)
	return "."
}

// tokenFromFile returns the oAuth token stored in a given filename.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth2.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

// saveToken stores an oAuth token in the given filename.
func saveToken(file string, token *oauth2.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

// tokenFromWeb attempts to authorize the application by directing the user to
// Google. This spawns a web server on localhost which the user is eventually
// redirected to.
func tokenFromWeb(debug bool, config *oauth2.Config) *oauth2.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	listener := newLocalListener(*redirectPort)
	go http.Serve(listener, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer listener.Close()

	config.RedirectURL = fmt.Sprintf("http://%s:%s/", *redirectHost, redirectPort)
	authUrl := config.AuthCodeURL(randState)
	go openUrl(authUrl)

	log.Printf("Authorize this app at: %s", authUrl)
	code := <-ch
	log.Printf("Got code: %s", code)

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return tok
}

// historyListener keeps track of all connections that it's ever
// accepted.
type historyListener struct {
	net.Listener
	sync.Mutex // protects history
	history    []net.Conn
}

func (hs *historyListener) Accept() (c net.Conn, err error) {
	c, err = hs.Listener.Accept()
	if err == nil {
		hs.Lock()
		hs.history = append(hs.history, c)
		hs.Unlock()
	}
	return
}

func newLocalListener(port string) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func openUrl(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.")
}

// New OAuthClient creates a new client against the Microsoft OAuth2 API
func OAuthClient(appName string, debug bool, config *oauth2.Config) *http.Client {
	cacheFile := tokenCacheFilename(appName, config)
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token = tokenFromWeb(debug, config)
		saveToken(cacheFile, token)
	} else {
		log.Printf("Using cached token from %q", cacheFile)
	}

	return config.Client(oauth2.NoContext, token)
}
