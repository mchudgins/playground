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
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccV]o\xdb6\x14}n~ŭ\x86\r-0Iv\xf6\xd1L\x91\x158Y\x02\xf4\xa1@\x00\xbb\xe8\xde\n\x8a\xbc\x92\xb8P\xa4F^ٳ\x83\xfd\xf7A\x1f\x8e\xad%\xb1\xb3n\x1e\xf6D\xfa\xea\x9c\xebãC\x8a\xf1ka8\xad*\x84\x82J\x95\x9c\xc4\xcd\x00\x8a\xe9|\xe2\xa1\xf6\x9a\x022\x91\x9c\x00\xc4%\x12\x03^0\xeb\x90&^M\x99\u007f\xe6m\x1f\x14D\x95\x8f\xbf\xd5r1\xf1~\xf1?N\xfd+SV\x8cd\xaa\xd0\x03n4\xa1\xa6\x89\xf7\xfez\x82\"\xc7oyaM\x89\x93q׀$)L~\x9e\xcdaz\xfb\x1e\xae\u007f\xaf\x94\xb1h㰫7\bG\xabn\xf6J\xc8E`\xcd\x12\xee\xa1d6\x97\xdaO\r\x91)#\x18\x05\xef~\xc0\xf2\x1c*&\x84ԹO\xa6\x8a 8\x1d\xd46ؾ\xfc\xc7N\xc3HS\xe1\xf3B*\xf1\xc6\b\xf1\x16\xee!e\xfc.\xb7\xa6\xd6\xc2\xe7F\x19\x1b\xc1\xb2\x90\x84\xae4w\xf8,\x17\x17\xa8\xf7\x90\x1b^C\fj\xcd\x16L*\x96*\xfc\xec\x88Q\xed\xe0\x1ezdn٪\xff\x83\x01N\xa0\xe3=8\x8aXFh[Nkm\x04\x1e\xbc\xd9\x01\xbf\xf56\x1d\x18'\xb9x\xfc'\xa9\xaa7\xab\b\x04V\x169#\x14{\xb5\xec\xc0\x0eJ\xd9b\xb7J4+\xf137\n\xeeAHW)\xb6\x8a@j%5\xfa\xa92\xfc\xee\x1c\x96RP\x11\xc1\xbb\x87\x97\x13,\xd0:i\xf4\x8bh\xdbw\x1a\xb4\xfa^\xc2\xf9q\xb4%\xb1J^kQ\x19\xa9i\x97\xd7\x13\xfa\xb8)\xcc(\x82\xef\a\xac\x196\x84\xcch\xf2\x9d\\c\x04\xe3\xf1\xe8\xeb\xf3\xae\xb0D\x99\x17\x14Aj\x94\xe8(qا9\x0e\xbb\xbd\x15\xa7F\xac\x9a\x9d6N\x86\xf9/\xc6\xc9I,\xe4\"9i\a\xe0\x8a97\xf1\xacYz\xc9ɫ\xddR'\xa2\xa9\x0e\xca\x1bǽ\xbf4n{\x0e\xa1\x1bǼ\xa4d\x9a\xe5\bT <\xc5ڌ\x8f\x04<x7H\xdc\x1eM1\x83\xc2b6\xf1\x9a\xb3#\n\xc3T\xe6M*\x03\xc5tt6:\x1b\x85n\xc9\xf2\x1c\xad_\xcb\xf0\xa2\xb6jr\x00\x17:\xb4\v\xc91\xe8\u007f\a\xbf:\xa3\xbd\xa4\xb2F\xc4!K\x9e\\\xf6N\xbe\xbcd1\xdek\xcdfa;\xe1\xf7\xdaS\xeb\xb6N\x95\xe4\xff\xb1[\xcap\xa6\n\xe3h\xafWO\xa3\x9eq\xea\x01|<\xbb\xe6\xe8\b\xa4v\xc44G\xb0\xb5\xd6R\xe7`4\xacLm\xa1d\xbc\x90\x1a\x1f\x99\xd7O\xfe\xf1.\xb8\x9a}:\x10\xfe\xabڑ)\xd1¬s\b>\x19{\u05c8\x97F\x1fM\xd6|z:\x1a\x8d\x0e(\xfbPS\xcd\x14\xdc\xd4Z\xc0\xdc2\xed2\xb40\xcdQ\xf3\x15\xccV\x8e\xb0<~\xe66q\x11\x8e\xb8\xb1U\xa0\x91\x06i+\xb3t\xf0l\u007f\xdc>\xdc\\\xbe4h\xc1\x17D\xed\xb2\xb9\x96\xb4y\xbbU\x8c2c\xff\x17\x06\xb1\xbfe\xd0\xf4\x85\x06\x9d2U\x15\xec\vL\x9a6\xbc\x03.\xfd[1\xcf\xea\x94\x1d\xfa\xf8\xdc\x18\xe3_2{4\r\xb9\xb1\xd9\x01\tsiQ\x80\xc9\xe0\x9b\xaf\xbe\xfb\xe9\xbc\x15\xdd\xce.\xe0\xa3\xeb>\x8b\x1a\x97\xddæ[;i\xce\xfe\xd7G\x13\xbd^\xaf\xd7\aD\xb3J6\xa7)8e\x96ju\xf1\xbc\x94n\x8c\xc3\xfe\xce\x11\xb6\xd7\xfe?\x03\x00\x00\xff\xffA\xa5\x10~\x06\f\x00\x00",
		hash:  "c92147d17b904e596bd5f7dd3381fa422b635f2d6e0d69e5aed766fd70472825",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1489506083, 0),
		size:  3078,
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
