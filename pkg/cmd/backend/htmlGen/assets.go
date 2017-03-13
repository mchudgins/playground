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
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccV]o\xdb6\x14}\xae\u007fŭ\x86\r-0Yv\xf7\x95)\x92\x02'K\x80>\x14\b`\x17\xdd[qE^I\\(R#)yr\xb0\xff>H\xb2k{I\xed,\x9b\x87>\x91\xbe<\xe7\xfa\xf0\xe0\x90b\xf4\x92k\xe6ڊ\xa0p\xa5LFQ7\x80D\x95\xc7\x1e)\xaf+\x10\xf2d\x04\x10\x95\xe4\x10X\x81ƒ\x8b\xbd\xdae\xfe\x99\xb7](\x9c\xab|\xfa\xbd\x16M\xec\xfd꿟\xf9W\xba\xacЉT\x92\aL+G\xca\xc5\xde\xdb\xeb\x98xN߲\xc2\xe8\x92\xe2\xe9\xd0\xc0\t')\xf9e\xbe\x80\xd9\xed[\xb8\xfe\xa3\x92ڐ\x89\x82\xa1\xde!\xack\x87\xd9\v.\x9a\xb1\xd1\xcbP\xb9\xc2g\x85\x90\xfc\x95\xe6\xfc5\xdcC\x8a\xec.7\xbaV\xdcgZj\x13²\x10\x8el\xa9\xef\xe8\x1c\xfe|\x9cK\r\xa9\x03\xe4\x8e\xd7\x11ǵ\xc2\x06\x85\xc4T\xd2G\xeb\xd0\xd5\x16\xeea\x8d\xcc\r\xb6\xeb?\xd8\xc3q\xb2l\r\x0eC\xcc\x1c\x99\x9e\xd3;\x11\x82\a\xafv\xc0\xaf\xbdM\adN4\x0f\xff$\x95\xf5f\x17cN\x95!\x86\x8e\xf8A-;\xb0\xa3R\xb6ح\x12\x85%}dZ\xc2=pa+\x89m\bBI\xa1\xc8O\xa5fw\xe7\xb0\x14\xdc\x15!\xfcD\xe5\x86Ӑ\xb1B\xab'\xd1~\xd8\xd2z}O\xe1\xfc8ْ\xb0\x12\u05caWZ(\xb7\xcb[\x13J4\xb9P\xbe\xa4̅\xf0\xfd\x1ekN\x1d!\xd3\xca\xf9V\xac(\x84\xe9t\xf2\xf5\xf9PX\x92\xc8\v\x17B\xaa%\x1f(Q\xb0\x0e_\x14\fG!J5o\xbb\x831M\xf6\xe3ZL\x93Q\xc4E\x93\x8c\xfa\x01\x98Dkc\xcf襗\x8c^\xec\x96\x06\x11]u\xaf\xbcq\xdc\xfb[\xe3\xbe\xe7>t㘗\x94\xa80'p\x05\xc1c\xac\xcd\xf8@\xc0'\xef\xf6\x12w@S\x84P\x18\xcab\xaf;\xeaa\x10\xa4\"\xefR9\x96\xa8³\xc9\xd9$\xb0K\xccs2~-\x82\x8b\xda\xc8\xf8\b.\xb0d\x1a\xc1h\xbc\xfe=\xfe\xcdj\xe5%\x95\xd1<\n0yt\xdb;\xf9\xf2\x92fzК\xcd\xc6v\xc2\xef\xf5\x97\xccm\x9dJ\xc1\xfeg\xb7\xa4f(\vm\xddA\xaf\x1eG}ƩO\xe0\xd3ٵ \xeb@(\xebP1\x02S+%T\x0eZA\xabk\x03%\xb2B(z`\xdez\xf2\xafO\xc1\xd5\xfcÑ\xf0_\xd5\xd6\xe9\x92\f\xcc\a\x87\xe0\x836w\x9dx\xa1\xd5\xc9d-fo&\x93\xc9\x11e\xefjW\xa3\x84\x9bZqX\x18T6#\x03\xb3\x9c\x14ka\xdeZG\xe5\xe93\xb7\x89\v\xb7\x8eiS\x8d\x15\xb9\xbd\xb4\x95Y\xba\xb7v8n\xefn.\x9f\x1a\xb4\xf13\xa2vٽ\"\xfa\xbc\xddJt\x996_\x84A\xf8\x8f\f\x9a=Ѡ7(\xab\x02\x9faҬ\xe3\x1dq鿊yV\xa7x\xec\xe3s\xa3\xb5\u007f\x89\xe6d\x1arm\xb2#\x12\x16\xc2\x10\a\x9d\xc17_}\xf7\xf3y/\xba\x9f]\xc0{;|\x16\x15-\x87Ů[?\xe9\xee\xfe\x97'\x13\xbdZ\xadVGDc%\xba\xdb\x14\xac\xd4K\xd9^|^\xca0F\xc1\xfa\xcd\x11\xf4\xaf\xf4\xbf\x02\x00\x00\xff\xff\xe1j8ǵ\v\x00\x00",
		hash:  "0686991b2b952e19c48c5442a44524a5620a8f311bc62db92d050f21369a3ee7",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1489433081, 0),
		size:  2997,
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
