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
		data:  "\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xff\xccW\xefo\xdb6\x10\xfd\x1c\xff\x157\r\x1bZ`\xfa\xe1l\xc5\x02E\xd6\xe0f)P`\x05\x02$C\xf7-\xa0ȓą\"5\xf2d\xcf3\xf6\xbf\x0f\xb2d[\x8e\xdd8.\x9a\xa1\x9f\xa8\x9c\xde;\x1e\xdfݣ\xe2\xe4\x1ba8-j\x84\x92*\x95\x8e\x92v\x01\xc5t1\xf1P{m\x00\x99HG\x00I\x85Ā\x97\xcc:\xa4\x89\xd7P\xee_x\xdb\x17%Q\xed\xe3_\x8d\x9cM\xbc?\xfcߧ\xfe\x95\xa9jF2S\xe8\x017\x9aP\xd3\xc4{\u007f=AQ\xe0\x0f\xbc\xb4\xa6\xc2ɸK@\x92\x14\xa6\xbf\xde\xde\xc1\xf4\xe6=\\\xff]+c\xd1&a\x17o\x11\x8e\x16\xdd\xd3Y9\x86%d\x8c?\x14\xd64Z\xf8\xdc(cc\xf86\x8f\xf2(\xcf/\xa1fBH]\xf8d\xea\x18\xc6Xm#\x99!2U\x1f\xfc\xb7\xcd%\xe4,\xb0f\x0eK\xa8\x98-\xa4\xde@\xa2\xe0\xe77C\xea*Yp\xfe\xe6P\xba><L\x18k*}^J%^\x19!^\x1f,x^JBW\x99\a\xfc$\x17g\xa8\x9f \xb7\xbc\x96\x184\x9a͘T,Sx\xef\x88Q\xe3`\t=\xb2\xb0l\xd1o\xb0\x83\x13\xe8x\x0f\x8ec\x96\x13\xda\x15gզ\x18<x5\x00\xbf\xf6\xd6\x19\x18'9\xdb\xdf$S\xcd\xfa\x14\x81\xc0\xda\"g\x84\xe2\xc9Z\x06\xb0\xa3\xa5l\xb1\x83Jjy\x8bt\xafY\x85\xf7\xdc(X\x82\x90\xaeVl\x11\x83\xd4Jj\xf43e\xf8\xc3#\xf8j\xabg\xc0\x8f\xa7\x9dKAe\f?mZ\x1f\xcc\xd0:i\xf4\xa9\xb4\xe3%\xf5\x9c\x1f\xa3-\x89\xd5\xf2Z\x8b\xdaHMC^O\xe8\x87YaN;[u\"\xc0\x12r\xa3\xc9w\xf2\x1f\x8ca<\x8e\xbe\xbb\xec\x02s\x94EI1dF\x89\xcbǎ\x18Ly\x12\xf6nL\xc2\xeenH2#\x16\xedM1Nw\xfd[\x8e\xd3Q\"\xe4,\x1d\xad\x16\xe0\x8a97\xf1\xac\x99{\xe9\xe8l\x18\xeaJk\xa3\a\u009b.{\x8f\U000b7a7fי\xab/\xe3n9\xc0]\xeb\xeb\xa5\x1f\x98f\x05\x02\x95\b\xfbiFg\xebu\xaf\xb0\x8d\xd2;ӿW\xeb\xb6ȄAi1\x9fx\xed\x9d\x18\x87a&\x8b\xd6!\x81b:\xbe\x88.\xa2\xd0\xcdYQ\xa0\xf5\x1b\x19\xfe\xd2X59\x82\v\x1dڙ\xe4\x18\xf4\u007f\a\u007f:\xa3\xbd\xb4\xb6F$!Kׅ\xefT3\x98F/\x9d\x8d\x0fb6\xa3\xd7\x1fl`D/\xbdi2%\xf9\xff\xac\x942\x9c\xa9\xd28zR\xa7èO\xa8\xb4\x01\xbf\x9cTw\xe8\b\xa4v\xc44G\xb0\x8d\xd6R\x17`4,Lc\xa1b\xbc\x94\x1a\xf7\xc4\xeb\x1f\xbe\x943\xaen?\x9ej\x88\xabƑ\xa9\xd0\xc2m\xa7\x1c|4\xf6\xa1=\x944\xfa\xf9\xbd\u07bb\xf0O\xe8w\x95\xb3@8\xe2\xc6ցF\n\xbf\xe4P\x1f\xfe\xc2x\xe9oX0\xbe\x80v\xa3\x86\xb7Gݴ\xee\xa5[t7=\x8f\xa2\xe8\xe4k\xab\xa1\x86)x\xd7h\x01w\x96i\x97\xa3\x85i\x81\x9a/\xe0v\xe1\b\xab\x97\xf7\xe5\xdaR\xc3^\r\x1dY\xe5\xd9λ\xa7-\xf9\xe1\xdd\xdb\xe7\xb68\xf8\f;\xbem\xff\x1d]y\xf2F1ʍ\xfd*\x04b'\t4}\xa6@\xe7L\xd5%\xfb\f\x91\xa6-\xef\x88J\x9b\xf1\xef\xd6$\xec\xbf\xf3\xe1\xea\xa7\xc2\u007f\x01\x00\x00\xff\xff\x81FA\xa1:\f\x00\x00",
		hash:  "b6d6bbf40612406bf8714b114ce17a7c5080a6d52bc39aec7b9e8c5b754b1b85",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1489527581, 0),
		size:  3130,
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
