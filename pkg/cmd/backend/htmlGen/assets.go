package htmlGen

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
	"apiList.html": {
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccV]o\xdb6\x14}n~ŭ\x86\r-0Y\xf6\xbe\x90)\xb2\x02'K\x80>\x14\b`\x17\xdd[A\x91W\x12\x17\x8aW#\xaf\xec\xc9\xc1\xfe\xfb \xc9n\xec%\xb5\xbbl\x19\xfaD\xfa\xf2\x1c\xea܃C\xd2\xc9KE\x92\xdb\x1a\xa1\xe4ʤ'I7\x80\x11\xb6\x98\x06h\x83\xae\x80B\xa5'\x00I\x85,@\x96\xc2y\xe4i\xd0p\x1e\x9e\x06\xf7\v%s\x1d\xe2\xef\x8d^N\x83_\xc3w\xb3\xf0\x92\xaaZ\xb0\xce\f\x06 \xc92Z\x9e\x06o\xae\xa6\xa8\n\xfcV\x96\x8e*\x9cN\x86\rX\xb3\xc1\xf4\x97\xf9\x02f7o\xe0\xea\x8fڐC\x97DC\xbdCxn\x87\xd9\v\xa5\x97#G\xab\xd8r\x19\xcaR\x1b\xf5\x8a\x94z\rw\x90\ty[8j\xac\n%\x19r1\xacJ\xcd\xe8+\xba\xc53\xf8\xf3q..\xd1\x1e w\xbc\x8e8j\xacX\nmDf\xf0\x83g\xc1\x8d\x87;\xd8 \v'\xda\xcd\a\xf6p\n\xbd܀\xe3X䌮\xe7\xf4N\xc4\x10\xc0\xab\x1d\xf0\xeb`\xbb\x83\x90\xac\x97\x0f?\x92\x99f\xdb\xc5Ha\xedP\nFuP\xcb\x0e쨔{\xec\xbd\x12+*\xfc \xc9\xc0\x1d(\xedk#\xda\x18\xb45\xdab\x98\x19\x92\xb7g\xb0Ҋ\xcb\x18&?b\xb5%-\xd1yM\xf6\xb3x;\xb4^\xe0\xe7p~\x1aߓD\xad\xaf\xac\xaaI[\xde\xe5m\b\x95p\x85\xb6\xa1\xc1\x9cc\xf8a\x8f5ǎ\x90\x93\xe5\xd0\xeb5\xc60\x99\x8c\xbf>\x1b\n+\xd4E\xc91dd\xd4@I\xa2M\xfa\x92h8\vIF\xaa\xedN\xc6$\xdd\xcfk9IO\x12\xa5\x97\xe9I?\x804\xc2\xfbi\xe0h\x15\xa4'/vK\x83\x88\xae\xbaW\xdeZ\x1e\xfcm\xe3~\xcf}\xe8ֱ \xad\x84\x15\x05\x02\x97\b\x8f\xb1\xb6\xe3\x03\x01\x1f\xbdۋ\xdc\x01M\x89\x80\xd2a>\r\xba\xb3\x1eGQ\xa6\x8b.\x96##l|:>\x1dG~%\x8a\x02]\xd8\xe8\xe8\xbcqfz\x04\x17ytK-q\xb4\xf9=\xfa͓\r\xd2ڑJ\"\x91\x1el{+z'\xd9A\u007f\x83\xdc4\x99\xd1\xf2\u007fv\u0090\x14\xa6$\xcf\a}x\x1c\xf5\t\x17>\x82\x9ff\xc5\x02=\x83\xb6\x9e\x85\x95\b\xae\xb1V\xdb\x02\xc8BK\x8d\x83J\xc8R[|`\xccf\xf2\xaf\xd3{9\u007f\u007f$\xb4\x97\x8dg\xaa\xd0\xc1|\xe8\x1eޓ\xbb\xed\xc4k\xb2\xcf&k1\xfbn<\x1e\x1fQ\xf6\xb6\xe1F\x18\xb8n\xac\x82\x85\x13\xd6\xe7\xe8`V\xa0\x95-\xcc[\xcfX=\u007f\x9e\xb6QP\x9e%\xb9zd\x91\xf7\x92T\xe5\xd9\xde\xda\xe1(\xbd\xbd\xbexZ\x88.\xba\x87\xbdOҍ\x11\x9c\x93\xfb\"Z\x17\xff\xa8\xf5\xd9\xd3Z\x9f\x99\xba<\xd6\xfb\u007f\x15˼\xc9ıK\xfe\x9a(\xbc\x10\xee\xd94\x14\xe4\xf2#\x12\x16ڡ\x02\xcaᛯ\xbe\xff\xf9\xac\x17\xdd\xcf\xce\xe1\x9d\x1f\x9e\x1f\x8b\xaba\xb1ۭ\x9ft\xf7\xf0\xcbg\x13\xbd^\xaf\xd7GD\x8bZw\xb7\x1fxC+Ӟ\u007fZ\xca0&\xd1\xe6m\x8f\xfa\xbf\xc3\u007f\x05\x00\x00\xff\xff\xaf\xb0\x03a\x1e\v\x00\x00",
		hash:  "5107670aad46ac1806b992c7a7fc6f52d926e372c06c2fb2ef2e253b705aefc4",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1489432673, 0),
		size:  2846,
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
