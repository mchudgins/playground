package backend

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type staticFilesFile struct {
	data  string
	mime  string
	mtime time.Time
	// size is the size before compression. If 0, it means the data is uncompressed
	size int
	// hash is a sha256 hash of the file contents. Used for the Etag, and useful for caching
	hash string
}

var staticFiles = map[string]*staticFilesFile{
	"service.swagger.json": {
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff̕\xc1n\xdb<\f\xc7\xef~\nB\xdf\a\xec\xb2\xc5Y\x8f\xb9\xae\xc0\xd6k\xb7ۖ\x83f3\xb6\n[RI:CP\xf8\xdd\a\xc9N\xachA6l-\xb6^\xaa\x88\x14\xc5\xff\x8f\x14\xfdT\x00(\xfe\xa6\x9b\x06Im@ݬ\xd6\xeau\xd83v\xe7\xd4\x06\x82\x1d@\x89\x91\x0e\x83\x9d\x91\xf6\xa6\u0095''.z\x02\xa8=\x12\x1bg\x83}^\x82u\x02\x8c\xa2\n\x801\xc6\xe3\xaa\xc5\x1eYm\xe0\xf3t\xa8\x15\xf1\xc7\x00a\xcd\xc1w\x1b}+gy8s\xd6\xdew\xa6\xd2b\x9c-\x1f\xd8\xd9\xc5ד\xab\x87\xea\x17}\xb5\xb4\xbc\x88*\xb57\xe5\xfem\x89U\xeb\xca'\xab{\x1cOF\x00\u0560$?\x83\x84\xa1\xef5\x1d\x82̏hk\x06\r\r!\x8a\xb1\xcd,$\xba9\x8f\x14o\xbf\xab\xa3\xab>|\xc0\xaes\xa9\v!{g\x19\xf9\xec\x02\x00u\xb3^g[\x00\xaaF\xae\xc8x\x99\x11'\x81\xa6\xb4\x02Y\xfd\xc31\x00\xf5?\xe1.\x9c\xf8\xaf\xacqg\xac\t\x11\xb8\x9cK\x18\x93\xbaG\xdf\x1d\xd4ٹ\xb1\xb8\xb4\x1e\x93\xec\xbd&ݣ -Ч\xbf,\xef@4\\\x1f\xffgI\x9b(%\x14$\xb7\x10>\x0e\x860\xb0\x13\x1a0\xb3\xca\xc1O}(t\x8e=Zw\x8ez-\x89\xfd\xa2\x96m\xa2Et\x93\xabP\xefCQ\x91\x96\xc3\xdb\"\r1\x9ez:\xa1\xbatՌ\xf7\x9d\xeb{g\xef\xf1q@N\xfb\xe8\xa4\xc0}}\xc0JN\nB#{$1YS\xa8\xca\x11a75\xd4m\xde/Wp\\\x81\x91\x16\x93\xbd~\x91\xb8\x03ǁ\xf2\x1cQ\x8b,\xfa2\x8d>\xb5\b4!\x86\x1e\x99u\x83P9+\xdaXc\x9b\xcd\x17\v\xf0\x06\xa4EЃ\xb4h%\x8c\x05\xac!\xe4\x06w\xb7\x8b\xf9\x8c\xf1\xb2=\xb3)\x92\xdb\xf3\xf2N\xef\xf8_\xab\xefψMY\x1f\x91]\x14\x98\x8c\x87?\x12\x170媮Υ\x8c\xec\xc5\xe6:&\xfe7h%\r\x16\x9b\xe4\xf8\t\xe0k\x18\x9fa\n\xfc6\xc8\xe9\xee\x8b\x1c\xe7\t\xfd\"\x10\xb3\xaf\xd6\xf5\xa7\x1aI\x86g\xf9\x8a!$\xb5R˨-\xc6\xe2{\x00\x00\x00\xff\xffa\xff\xe8\b\xa6\b\x00\x00",
		hash:  "513e00312e3f57ec5e674314557569fb8cbe68e582b2b12aacd405febde56de6",
		mime:  "application/json",
		mtime: time.Unix(1489351868, 0),
		size:  2214,
	},
}

// NotFound is called when no asset is found.
// It defaults to http.NotFound but can be overwritten
var NotFound = http.NotFound

// ServeHTTP serves a request, attempting to reply with an embedded file.
func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f, ok := staticFiles[strings.TrimPrefix(req.URL.Path, "/")]
	if !ok {
		NotFound(rw, req)
		return
	}
	header := rw.Header()
	if f.hash != "" {
		if hash := req.Header.Get("If-None-Match"); hash == f.hash {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("ETag", f.hash)
	}
	if !f.mtime.IsZero() {
		if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && f.mtime.Before(t.Add(1*time.Second)) {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("Last-Modified", f.mtime.UTC().Format(http.TimeFormat))
	}
	header.Set("Content-Type", f.mime)

	// Check if the asset is compressed in the binary
	if f.size == 0 {
		header.Set("Content-Length", strconv.Itoa(len(f.data)))
		io.WriteString(rw, f.data)
	} else {
		if header.Get("Content-Encoding") == "" && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			header.Set("Content-Encoding", "gzip")
			header.Set("Content-Length", strconv.Itoa(len(f.data)))
			io.WriteString(rw, f.data)
		} else {
			header.Set("Content-Length", strconv.Itoa(f.size))
			reader, _ := gzip.NewReader(strings.NewReader(f.data))
			io.Copy(rw, reader)
			reader.Close()
		}
	}
}

// Server is simply ServeHTTP but wrapped in http.HandlerFunc so it can be passed into net/http functions directly.
var Server http.Handler = http.HandlerFunc(ServeHTTP)

// Open allows you to read an embedded file directly. It will return a decompressing Reader if the file is embedded in compressed format.
// You should close the Reader after you're done with it.
func Open(name string) (io.ReadCloser, error) {
	f, ok := staticFiles[name]
	if !ok {
		return nil, fmt.Errorf("Asset %s not found", name)
	}

	if f.size == 0 {
		return ioutil.NopCloser(strings.NewReader(f.data)), nil
	}
	return gzip.NewReader(strings.NewReader(f.data))
}

// ModTime returns the modification time of the original file.
// Useful for caching purposes
// Returns zero time if the file is not in the bundle
func ModTime(file string) (t time.Time) {
	if f, ok := staticFiles[file]; ok {
		t = f.mtime
	}
	return
}

// Hash returns the hex-encoded SHA256 hash of the original file
// Used for the Etag, and useful for caching
// Returns an empty string if the file is not in the bundle
func Hash(file string) (s string) {
	if f, ok := staticFiles[file]; ok {
		s = f.hash
	}
	return
}
