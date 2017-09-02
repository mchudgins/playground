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
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccW\xefo\xdb6\x10\xfd\\\xff\x157\r\x1bZ`\xfa\xe1l\xc5\x02E\xd6\xe0f)P`\x05\x02$C\xf7-\xa0ȓą\"5\xf2$\xcf\v\xf6\xbf\x0f\xb2d[\x8e]'.\x1a`\x9f(\x9f\xdf;\x1e\xdfݣ\xec\xe4\x1ba8-k\x84\x92*\x95N\x92n\x01\xc5t1\xf3P{]\x00\x99H'\x00I\x85Ā\x97\xcc:\xa4\x99\xd7P\xee\x9f{\xdb/J\xa2\xdaǿ\x1a\xd9μ?\xfc\xdf\xe7\xfe\xa5\xa9jF2S\xe8\x017\x9aP\xd3\xcc\xfbp5CQ\xe0\x0f\xbc\xb4\xa6\xc2ٴO@\x92\x14\xa6\xbf\xde\xdc\xc2\xfc\xfa\x03\\\xfd]+c\xd1&a\x1f\xef\x10\x8e\x96\xfdӫr\n\x0f\x901~_X\xd3h\xe1s\xa3\x8c\x8d\xe1\xdb<ʣ<\xbf\x80\x9a\t!uᓩc\x98b\xb5\x8dd\x86\xc8TC\xf0\xdf.\x97\x90m`\xcd\x02\x1e\xa0b\xb6\x90z\x03\x89\x82\x9fߎ\xa9\xabd\xc1\xd9\xdbC\xe9\x86\xf08a\xac\xa9\xf4y)\x95xm\x84xs\xb0\xe0E)\t]e\xee\xf1\xb3\\lQ\x1f!w\xbc\x8e\x184\x9a\xb5L*\x96)\xbcsĨq\xf0\x00\x03\xb2\xb0l9l\xb0\x83\x13\xe8\xf8\x00\x8ec\x96\x13\xda\x15gզ\x18<x=\x02\xbf\xf1\xd6\x19\x18'\xd9\xeeo\x92\xa9f}\x8a@`m\x913Bq\xb4\x96\x11\xec\xc9R\xb6\xd8Q%\xb5\xbcA\xbaӬ\xc2;n\x14<\x80\x90\xaeVl\x19\x83\xd4Jj\xf43e\xf8\xfd#\xf8j\xabg\xc0\x9fN\xbb\x90\x82\xca\x18~ڴ>h\xd1:i\xf4\xa9\xb4\xa7K\x1a8?F[\x12\xab\xe5\x95\x16\xb5\x91\x9aƼ\x810\f\xb3\u009cv\xb6\xeaE\x80\aȍ&\xdf\xc9\u007f0\x86\xe94\xfa\xee\xa2\x0f,P\x16%Ő\x19%.\x1e;b4\xe5I8\xb81\t\xfb\xbb!ɌXv7\xc54\xdd\xf5o9M'\x89\x90m:Y-\xc0\x15sn\xe6Y\xb3\xf0\xd2ɫq\xa8/\xad\x8b\x1e\bo\xba\xec=\xcaߥ\xfe^g\xae\xbe\x88\xfb\xe5\x00w\xad\xaf\x97~d\x9a\x15\bT\"짙\xbcZ\xaf{\x85m\x94ޙ\xfe\xbdZ\xb7E&\fJ\x8b\xf9\xcc\xeb\xee\xc48\f\x95\xe1L\x95\xc6Q|\x1e\x9dG\xa1[\xb0\xa2@\xeb72\xfc\xa5\xb1jv\x14\x15:\xb4\xad\xe4\x18\f\x9f\x83?\x9d\xd1^\xba\x01'!Kו\xef\x943\x1aG/m\xa7\a1\x9b\xd9\x1bN6r\xa2\x97ޢ#\x90\xda\x11\xd3\x1c\xc16ZK]\x80Ѱ4\x8d\x85\x8a\xf1Rj|y\xf1\x04\xb6\xa8L\x8d\xd6\x05\xc2\x117\xb6\x0e\xa49*\xe4ӌψZ[#^N\xcf\xeb&S\x92\x1f\x1f\xbd\xe1\xe1k\xb9\xe5\xf2\xe6ө&\xb9l\x1c\x99\n-\xdc\xf4\x12\xc1'c\xef\xbb#H\xa3\x9f\xdf콗\xc0\t\r\xafr\xb6\xe9\x9bF\n\xbff_\x0e\xbfu\xbc\xf47,\x18_B\xb7Qû\xa3n\x06\xff\xa5[t;?\x8b\xa2\xe8䫬\xa1\x86)x\xdfh\x01\xb7\x96i\x97\xa3\x85y\x81\x9a/\xe1f\xe9\b\xab\x977\xe6\xda;\xe3^\x8dm\xf8\xb8\x8fǽ\xf7\xf1\xfd\xfc\x99->c\xaa.\xd9\x17\x18p\xde\xf1`u\xad]+F\xb9\xb1\xff\v\x95\xb2\x93Tz\xf7\\#\x04_\xa0л\xee\x87\xfcq\x816\xe3߯I8\xbc\xfb\xc3\xd5߇\xff\x02\x00\x00\xff\xffe_\xbe\xf5N\f\x00\x00",
		hash:  "f5ddc2321e69a1e4f4c5aa60f0fc578fe21343f764f541bfe2602bac1295fe9f",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1504375858, 0),
		size:  3150,
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
