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
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccUAo\xdb:\f\xbe\xfbW\x10z\x0fx\x97\xb7\xba\xeb1\xb7a\x05\xb6b\xb7v;u9h6c\xab\x88%\x95\xa4;d\x85\xff\xfb َ\x155Ȇ\xad\xc5\xd6K\x1d\x93\"\xbf\xef\xe3'\xfa\xb1\x00P\xfcU7\r\x92Z\x81\xba8;W\xff\x87w\xc6n\x9cZA\x88\x03(1\xb2\xc5\x10g\xa4\aS\xe1\x99''.f\x02\xa8\a$6Ά\xf8\xf4\b\xd6\t0\x8a*\x00\x86X\x8f\xab\x16;d\xb5\x82\xdb\xf1P+\xe2\xe7\x02\xe1\x99C\xee:\xe6V\xcer\u007f\x90\xac\xbdߚJ\x8bq\xb6\xbccg\x97\\O\xae\ueadf\xcbe\xacz2\xb2\xbbč\xb1&$\xf0DQ}b\xa4\x9b)\x9c\xd0\xde\xf9\xc8Z{\xf3\x01w3Z\x13\x99\xb6\xa8k\xa4\xf9\x9d\xd5]\xcc|\xd3K\xeb\xc8|\x8b\xed#\xfbb8\xe8\xbd\xc7\xf9\x98\xf5\xbc]\x0f\v)--/0J\xedM\xf9\xf0\xbaĪu\xe5ch5\xec\x83\x00\xaaAI~\x86^}\xd7i\n\xad\xd4\rښACC\x88bl3\xe1\x8di\xce#E\x9cWuLջ\xf7\xb8ݺ4\x85\x90\xbd\xb3\x8c|\xd0\x00@]\x9c\x9fg\xaf\x00T\x8d\\\x91\xf12y!)4\xc2\n\x16\xd0O\x8e\x01\xa8\u007f\t7\xe1\xc4?e\xbd\x8c\xa6\x9c\xbc\x16A]\xa3\xdf\xee\xd4\xc1\xb9\xa18\xf6<$\xe8\xbd&ݡ -\xee\x18\xff2\xdc\xf3\xf0\xe2\xff\f\xf48\xec0\x90<Bx\xdf\x1b\u00a0\x9dP\x8fYt\xb6\x0e\v\x1d\xca\x1e\xa3\x1bG\x9d\x96$~\x94\xcb:\xe1\"\xba\xc9Y\xa8wa\xa8H\xcb\xe1u\x91\x96\x18\xf6\x97\xaf~bx\xd8_差뜽\xc6\xfb\x1e9\xf5ў\x81\xfbr\x87\x95\xec\x19\x84\x1b\xe7\x91\xc4d\xa6P\x95#\xc2\xedh\xa8\xcb\xdc/'\xe48!F:L\xf6\xfaE\xea\xf6\x1c7\xdfsT-\xb2\xea\xcb\xda\xfc\xd8\"\xd0(1tȬ\x1b\x84\xcaY\xd1\xc6\x1a۬>[\x80W -\x82\xee\xa5E+a\u007fa\r\x01\x1b\\].\xe1\x03\x8d\x97ד6E\xd2=\x1f\xefx\x8f\xff\xb6\xf9\xfeH\xb1\x11\xf5,\xd9Q\x82\xc9z\xf8-rA\xa6\x9c\xd5ɽ\x94){\xd4\\3\xf0?\xa1Vb\xb0h\x92\xf9\x13\xc0\xa7d|\x86-\xf0\xcbB\x8e\xbd\x8f\xea8m\xe8\x17\x111\xfbj\x9d\xbe\xaaQ\xc9p-\xffc\b\xa0\xceԲj\x8b\xa1\xf8\x1e\x00\x00\xff\xff\xeb\x1bT\xdcO\t\x00\x00",
		hash:  "21ff02c236fca646e1c0ddbe0d74fda383e26d96a1187710c6acbc0be44f8bf4",
		mime:  "application/json",
		mtime: time.Unix(1489508779, 0),
		size:  2383,
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
